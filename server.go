package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"path"
	"sync"

	"aqwari.net/net/styx"
)

type server struct {
	services map[string]*service
	ctx      context.Context
	cfg      *config
	events   chan *event
	sync.Mutex
	errors []error
}

func newServer(ctx context.Context, cfg *config) (*server, error) {
	services := getServices(cfg)
	events, err := tailEvents(ctx, services)
	if err != nil {
		return nil, err
	}
	s := &server{
		services: services,
		events:   events,
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
		go s.run(svc)
	}
}

func (s *server) run(svc *service) {
	h := styx.HandlerFunc(func(sess *styx.Session) {
		uuid := rand.Int63()
		current := "server"
		if len(sess.Access) > 1 {
			current = sess.Access
		}
		c := &client{
			target:  svc.name,
			current: current,
		}
		svc.clients[uuid] = c
		for sess.Next() {
			q := sess.Request()
			c.reading = q.Path()
			handleReq(s, c, q)
		}
	})
	port := fmt.Sprintf(":%d", *listenPort)
	t := &styx.Server{
		Addr:    svc.addr + port,
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

func (s *server) getPath(c *client) string {
	return path.Join(*inpath, c.target, c.current, c.reading)
}
