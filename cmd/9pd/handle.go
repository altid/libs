package main

import (
	"os"
	"path"

	"github.com/altid/server/client"
	"github.com/altid/server/files"
	"github.com/go9p/styx"
)

type handler struct {
	c       *client.Client
	f       *files.Files
	basedir string
	target  string
}

// TODO: Give server access to the client manager
func (h *handler) walk() (os.FileInfo, error) {
	return h.f.Stat(h.c.Current(), h.target, uint32(h.c.UUID))
}

func (h *handler) open() (interface{}, error) {
	return h.f.Normal(h.c.Current(), h.target, uint32(h.c.UUID))
}

func (h *handler) path() string {
	return path.Join(h.basedir, h.c.Current(), h.target)
}

func handleReq(c *client.Client, r *files.Files, req styx.Request, base string) {
	h := &handler{
		c:       c,
		f:       r,
		basedir: base,
		target:  req.Path(),
	}

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
