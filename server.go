package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
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
		case notifyEvent:
			// TODO(halfwit) Figure out notifications
			continue
		case feedEvent:
			// Increment our unread count for any inactive buffers
			srv := s.services[e.service]

			t, ok := srv.tablist[e.name]
			if !ok {
				// We have a new buffer
				t := &tab{
					count:  1,
					active: false,
				}

				srv.tablist[e.name] = t
				continue
			}

			if !t.active {
				t.count++
				continue
			}

			for _, cl := range srv.clients {
				cl.feed <- struct{}{}
			}
		}
	}
}

func (s *server) start() {
	for _, svc := range s.services {
		go s.run(svc)
	}
}

func (s *server) run(svc *service) {
	port := fmt.Sprintf(":%d", *listenPort)
	t := &styx.Server{
		Addr:     svc.addr + port,
		ErrorLog: log.New(os.Stderr, "", 0),
		TraceLog: log.New(os.Stderr, "", 0),
		//Auth: auth,
	}

	t.Handler = styx.HandlerFunc(func(sess *styx.Session) {
		uuid := rand.Int63()
		current := "server"

		if len(sess.Access) > 1 {
			current = sess.Access
		}

		c := &client{
			uuid:    uuid,
			feed:    make(chan struct{}),
			target:  svc.name,
			current: current,
		}
		svc.clients = append(svc.clients, c)
		svc.tablist[current].active = true

		for sess.Next() {
			q := sess.Request()
			c.reading = q.Path()
			handleReq(s, c, q)
		}
	})

	switch *usetls {
	case true:
		if e := t.ListenAndServeTLS(*certfile, *keyfile); e != nil {
			log.Fatal(e)
		}
	case false:
		if e := t.ListenAndServe(); e != nil {
			log.Fatal(e)
		}
	}
}

func (s *server) getPath(c *client) string {
	return path.Join(*inpath, c.target, c.current, c.reading)
}
