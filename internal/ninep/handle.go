package ninep

import (
	"os"
	"path"

	"github.com/altid/server/client"
	"github.com/altid/server/files"
	"github.com/go9p/styx"
)

type handler struct {
	c      *client.Client
	target string
}

func (h *handler) walk() (os.FileInfo, error) {
	m, f := h.build()
	return f.Stat(m)
}

func (h *handler) open() (interface{}, error) {
	m, f := h.build()
	return f.Normal(m)
}

//all this needs to do is get a handler, the rest can happen in 9pd top level
func (h *handler) build() (*files.Message, *files.Handler) {
	s := h.c.Aux.(*service)
	m := &files.Message{
		Server: s.addr,
		Buffer: h.c.Current(),
		Target: h.target,
	}

	return m, s.files
}

func (h *handler) path() string {
	service := h.c.Aux.(*service)
	return path.Join(service.dir, h.c.Current(), h.target)
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
