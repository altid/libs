package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"sync"

	"aqwari.net/net/styx"
	"aqwari.net/net/styx/styxauth/factotum"
	"github.com/altid/libs/auth"
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
		srv := s.services[e.service]

		switch e.etype {
		//case streamEvent:
		case notifyEvent:
			t, ok := srv.tablist[e.name]
			if !ok {
				addTab(srv, e)
			}

			if !t.active {
				t.alert = true
			}
		case docEvent:
			t, ok := srv.tablist[e.name]
			if !ok {
				addTab(srv, e)
				continue
			}

			if !t.active {
				t.count++
			}
		case feedEvent:
			t, ok := srv.tablist[e.name]
			if !ok {
				addTab(srv, e)
				continue
			}

			if !t.active {
				t.count++
				continue
			}

			srv.sendFeed()
		}
	}
}

func (s *server) start() {
	var wg sync.WaitGroup

	for _, svc := range s.services {
		wg.Add(1)

		go func(svc *service) {
			defer wg.Done()

			if e := s.run(svc); e != nil {
				log.Println(e)
			}
		}(svc)
	}

	wg.Wait()
	s.ctx.Done()
}

func (s *server) run(svc *service) error {
	var rpc func() (io.ReadWriteCloser, error)

	if *enableFactotum {
		rpc = auth.OpenRPC
	} else {
		rpc = s.cfg.mockFactotum

	}

	af, aof := factotum.Start(rpc, "p9any")
	t := &styx.Server{
		Addr:     svc.addr + fmt.Sprintf(":%d", *listenPort),
		Auth:     af,
		OpenAuth: aof,
	}

	if *verbose {
		t.TraceLog = log.New(os.Stderr, "", 0)
	}

	if *debug {
		t.ErrorLog = log.New(os.Stderr, "", 0)
	}

	t.Handler = styx.HandlerFunc(func(sess *styx.Session) {
		uuid := rand.Int63()
		current := "server"

		dirs, err := ioutil.ReadDir(path.Join(*inpath, svc.name))
		if err != nil {
			log.Print(err)
			return
		}

		for _, file := range dirs {
			if file.IsDir() {
				current = file.Name()
				break
			}
		}

		if len(sess.Access) > 1 {
			current = sess.Access
		}

		c := &client{
			uuid:    uuid,
			feed:    make(chan struct{}),
			target:  svc.name,
			current: current,
		}
		svc.clients[uuid] = c

		if tab, ok := svc.tablist[current]; ok {
			tab.active = true
			tab.count = 0
		}

		for sess.Next() {
			q := sess.Request()
			c.reading = q.Path()
			handleReq(s, c, q)
		}

		delete(svc.clients, uuid)
		close(c.feed)
	})

	switch *usetls {
	case true:
		if e := t.ListenAndServeTLS(*certfile, *keyfile); e != nil {
			return e
		}
	case false:
		if e := t.ListenAndServe(); e != nil {
			return e
		}
	}

	return nil
}

func (s *server) getPath(c *client) string {
	return path.Join(*inpath, c.target, c.current, c.reading)
}

func addTab(srv *service, e *event) {
	t := &tabs{
		count:  1,
		active: false,
	}

	srv.tablist[e.name] = t
}
