package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"sync"

	"aqwari.net/net/styx"
)

type server struct {
	services map[string]*service
	clients  map[int64]*client
	ctx      context.Context
	cfg      *config
	events   chan *event
	inputs   chan interface{}
	controls chan interface{}
	sync.Mutex
	errors []error
}

type client struct {
	target  string
	reading string
	current string
}

type service struct {
	tabs map[string]*tab
	addr string
	name string
}

type message struct {
	service string
	buff    string
	file    string
}

type fileHandler struct {
	stat func(msg *message) (os.FileInfo, error)
	fn   func(msg *message) (interface{}, error)
}

var handlers = make(map[string]*fileHandler)

func addFileHandler(path string, fh *fileHandler) {
	handlers[path] = fh
}

func newServer(ctx context.Context, cfg *config) (*server, error) {
	services := make(map[string]*service)
	for _, svc := range cfg.listServices() {
		tabs, err := listInitialTabs(svc)
		if err != nil {
			log.Printf("Unable to add service %s, no tabs file found\n", svc)
			continue
		}
		service := &service{
			tabs: tabs,
			addr: cfg.getAddress(svc),
			name: svc,
		}
		services[svc] = service
	}
	events, err := listenEvents(ctx, services)
	if err != nil {
		return nil, err
	}
	s := &server{
		services: services,
		events:   events,
		clients:  make(map[int64]*client),
		inputs:   make(chan interface{}),
		controls: make(chan interface{}),
		ctx:      ctx,
		cfg:      cfg,
	}
	return s, nil
}

func (s *server) listenEvents() {
	for e := range s.events {
		switch e.etype {
		case feedEvent:
			// Increment our unread count for any inactive buffers
			srv := s.services[e.service]
			t, ok := srv.tabs[e.name]
			if !ok {
				// We have a new buffer
				srv.tabs[e.name] = &tab{1, false}
				continue
			}
			if !t.active {
				t.count++
			}
		case notifyEvent:
			// TODO(halfwit) Figure out notifications
		}
	}
}

func (s *server) start() {
	for _, svc := range s.services {
		go run(s, svc)
	}
}

func (s *server) getPath(c *client) string {
	return path.Join(*inpath, c.target, c.current, c.reading)
}

func run(s *server, srv *service) {
	h := styx.HandlerFunc(func(sess *styx.Session) {
		uuid := rand.Int63()
		current := "server"
		if sess.Access != "/" {
			current = sess.Access
		}
		c := &client{
			target:  srv.name,
			current: current,
		}
		s.clients[uuid] = c
		for sess.Next() {
			q := sess.Request()
			c.reading = q.Path()
			handleReq(s, c, q)
		}
	})
	port := fmt.Sprintf(":%d", *listenPort)
	t := &styx.Server{
		Addr:    srv.addr + port,
		Handler: h,
		//Auth: auth,
	}
	var err error
	if *usetls {
		err = t.ListenAndServeTLS(*certfile, *keyfile)
	} else {
		err = t.ListenAndServe()
	}
	if err != nil {
		log.Fatal(err)
	}
}

func walk(svc *service, c *client) (os.FileInfo, error) {
	h, m := handler(svc, c)
	i, err := h.stat(m)
	if err != nil {
		log.Print(err)
	}
	info, ok := i.(os.FileInfo)
	if !ok {
		return nil, errors.New("requested file does not exist on server")
	}
	return info, err
}

func open(svc *service, c *client) (io.ReadWriteCloser, error) {
	h, m := handler(svc, c)

	i, err := h.fn(m)
	info, ok := i.(io.ReadWriteCloser)
	if !ok {
		return nil, errors.New("requested file does not exist on server")
	}
	return info, err
}

func handler(svc *service, c *client) (*fileHandler, *message) {
	h, ok := handlers[c.reading]
	if !ok {
		h = handlers["/default"]
	}
	m := &message{
		service: svc.name,
		buff:    c.current,
		file:    c.reading,
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
		case "/", "/tabs":
			msg.Rutimes(nil)
		default:
			fp := s.getPath(c)
			msg.Rutimes(os.Chtimes(fp, msg.Atime, msg.Mtime))
		}
	case styx.Ttruncate:
		switch msg.Path() {
		case "/", "/tabs":
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
