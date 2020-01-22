package main

import (
	"os"

	"aqwari.net/net/styx"
)

type fileHandler struct {
	stat func(msg *message) (os.FileInfo, error)
	fn   func(msg *message) (interface{}, error)
}

var handlers = make(map[string]*fileHandler)

type message struct {
	state   chan *update
	service string
	buff    string
	file    string
}

func addFileHandler(path string, fh *fileHandler) {
	handlers[path] = fh
}

func walk(svc *service, c *client) (os.FileInfo, error) {
	h, m := handler(svc, c)
	return h.stat(m)
}

func open(svc *service, c *client) (interface{}, error) {
	h, m := handler(svc, c)
	return h.fn(m)
}

func handler(svc *service, c *client) (*fileHandler, *message) {
	m := &message{
		state:   svc.state,
		service: svc.name,
		buff:    c.current,
		file:    c.reading,
	}
	h, ok := handlers[c.reading]
	if !ok {
		return handlers["/default"], m
	}
	return h, m
}

func handleReq(s *server, c *client, req styx.Request) {
	service, ok := s.services[c.target]
	if !ok {
		req.Rerror("%s", "No such service")
		return
	}
	switch msg := req.(type) {
	case styx.Twalk:
		msg.Rwalk(walk(service, c))
	case styx.Topen:
		msg.Ropen(open(service, c))
	case styx.Tstat:
		msg.Rstat(walk(service, c))
	case styx.Tutimes:
		switch msg.Path() {
		case "/tabs", "/ctl":
			msg.Rutimes(nil)
		default:
			fp := s.getPath(c)
			msg.Rutimes(os.Chtimes(fp, msg.Atime, msg.Mtime))
		}
	case styx.Ttruncate:
		switch msg.Path() {
		case "/tabs", "/ctl":
			msg.Rtruncate(nil)
		default:
			fp := s.getPath(c)
			msg.Rtruncate(os.Truncate(fp, msg.Size))
		}
	case styx.Tremove:
		switch msg.Path() {
		case "/notification":
			fp := s.getPath(c)
			msg.Rremove(os.Remove(fp))
		default:
			msg.Rerror("%s", "permission denied")
		}
	}
}
