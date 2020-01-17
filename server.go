package main

import (
	"context"
	"log"
	"math/rand"
	"os"
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
	current string
}

type service struct {
	tabs map[string]*tab
	port int
	addr string
	name string
}

type message struct {
	service string
	data    string
	buff    string
}

type fileHandler struct {
	fn func(srv *service, msg *message) (interface{}, error)
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
			port: cfg.getPort(svc),
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

func run(s *server, srv *service) {
	h := styx.HandlerFunc(func(sess *styx.Session) {
		uuid := rand.Int63()
		c := &client{
			target: srv.name,
		}
		s.clients[uuid] = c
		for sess.Next() {
			handleReq(s, sess.Request())
		}
	})
	t := &styx.Server{
		Addr:    srv.addr + ":564",
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

func handleReq(s *server, req styx.Request) {
	switch msg := req.(type) {
	case styx.Twalk:
		msg.Rwalk(os.Stat(msg.Path()))
	case styx.Topen:
	case styx.Tstat:
	}
}
