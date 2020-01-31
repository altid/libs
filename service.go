package main

import (
	"log"
	"sync"
)

type cmdKey int

const (
	bufferCmd cmdKey = iota
	reloadCmd
	linkCmd
	openCmd
	closeCmd
)

type service struct {
	commands chan *cmd
	clients  map[int64]*client
	tablist  map[string]*tab
	addr     string
	name     string
}

type client struct {
	uuid    int64
	feed    chan struct{}
	target  string
	reading string
	current string
	sync.Mutex
}

type cmd struct {
	uuid  int64
	key   cmdKey
	value string
}

func getServices(cfg *config) map[string]*service {
	services := make(map[string]*service)

	for _, svc := range cfg.listServices() {
		tlist, err := listInitialTabs(svc)
		if err != nil {
			log.Printf("Unable to add service %s, no tabs file found\n", svc)
			continue
		}

		service := &service{
			clients:  make(map[int64]*client),
			commands: make(chan *cmd),
			tablist:  tlist,
			addr:     cfg.getAddress(svc),
			name:     svc,
		}

		go service.watchCommands(cfg)
		services[svc] = service
	}

	return services
}

func (s *service) watchCommands(cfg *config) {
	for cmd := range s.commands {
		cl, ok := s.clients[cmd.uuid]
		if !ok {
			continue
		}

		switch cmd.key {
		case reloadCmd:
			s.addr = cfg.getAddress(s.name)
		case bufferCmd:
			s.move(cl, cmd.value)
			cl.sendFeedEOF()
		case openCmd:
			s.open(cl, cmd.value)
			cl.sendFeedEOF()
		case closeCmd:
			s.close(cl)
			cl.sendFeedEOF()
		case linkCmd:
			s.close(cl)
			s.open(cl, cmd.value)
			cl.sendFeedEOF()
		}
	}
}

func (s *service) open(c *client, name string) {
	s.checkInactive(c)
	c.current = name
}

func (s *service) close(c *client) {
	delete(s.tablist, c.current)

	for _, cl := range s.clients {
		if cl.current != c.current {
			continue
		}

		// Grab first item
		for _, t := range s.tablist {
			cl.current = t.name
			t.active = true
			t.count = 0
			break
		}
	}

}

func (s *service) move(c *client, name string) {
	t, ok := s.tablist[name]
	if !ok {
		t = &tab{
			name: name,
		}
		s.tablist[name] = t
	}

	t.active = true
	t.count = 0

	s.checkInactive(c)
	c.current = name
}

func (s *service) checkInactive(c *client) {
	for _, cl := range s.clients {
		if cl.uuid == c.uuid {
			continue
		}

		// At least one listener, no need to update
		if cl.current == c.current {
			return
		}
	}

	if t, ok := s.tablist[c.current]; ok {
		t.active = false
	}
}

func (cl *client) sendFeedEOF() {
	cl.Lock()
	close(cl.feed)
	cl.feed = make(chan struct{})
	cl.Unlock()
}
