package ninep

import (
	"os"
	"path"

	"github.com/altid/server/client"
	"github.com/go9p/styx"
)

type handler struct {
	c      *client.Client
	target string
}

func (h *handler) walk() (os.FileInfo, error) {
	service := h.c.Aux.(*service)
	return service.files.Stat(h.c.Current(), h.target)
}

func (h *handler) open() (interface{}, error) {
	service := h.c.Aux.(*service)
	return service.files.Normal(h.c.Current(), h.target)
}

func (h *handler) path() string {
	service := h.c.Aux.(*service)
	return path.Join(service.basedir, h.c.Current(), h.target)
}

func handleReq(c *client.Client, req styx.Request) {
	h := &handler{c, req.Path()}

	switch msg := req.(type) {
	case styx.Twalk:
		msg.Rwalk(h.walk())
	case styx.Topen:
		msg.Ropen(h.open())
	case styx.Tstat:
		msg.Rstat(h.walk())
	case styx.Tutimes:
		switch msg.Path() {
		case "/tabs", "/ctl", "/feed":
			msg.Rutimes(nil)
		default:
			fp := h.path()
			msg.Rutimes(os.Chtimes(fp, msg.Atime, msg.Mtime))
		}
	case styx.Ttruncate:
		switch msg.Path() {
		case "/tabs", "/ctl", "/feed":
			msg.Rtruncate(nil)
		default:
			msg.Rtruncate(os.Truncate(h.path(), msg.Size))
		}
	case styx.Tremove:
		switch msg.Path() {
		case "/notification":
			msg.Rremove(os.Remove(h.path()))
		default:
			msg.Rerror("%s", "permission denied")
		}
	}
}
