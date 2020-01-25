package main

import (
	"log"
	"sync"
)

type updateKey int

const (
	bufferUpdate updateKey = iota
	configUpdate
	linkUpdate
	openUpdate
	closeUpdate
)

type service struct {
	state   chan *update
	clients []*client
	tabs    map[string]*tab
	addr    string
	name    string
	sync.Mutex
}

type client struct {
	feed    chan struct{}
	target  string
	reading string
	current string
	uuid    int64
}

type update struct {
	uuid  int
	key   updateKey
	value string
}

func getServices(cfg *config) map[string]*service {
	services := make(map[string]*service)

	for _, svc := range cfg.listServices() {
		tabs, err := listInitialTabs(svc)
		if err != nil {
			log.Printf("Unable to add service %s, no tabs file found\n", svc)
			continue
		}

		service := &service{
			state:   make(chan *update),
			tabs:    tabs,
			addr:    cfg.getAddress(svc),
			name:    svc,
		}

		go service.watch(cfg)
		services[svc] = service
	}

	return services
}

func (s *service) watch(cfg *config) {
	for update := range s.state {
		switch update.key {
		case configUpdate:
			s.addr = cfg.getAddress(s.name)
		case bufferUpdate:
			// A client is switching buffers. Go through
			// and update all unread counts to reflect this
			// Mark old buffer as inactive if there are no
			// more readers on it

			// We close feed so that all readers can send the EOF
			/*
				s.clients
				s.Lock()
				close(s.feed)
				s.feed = make(chan struct{})
				s.Unlock()
			*/
			continue
		case openUpdate:
			// A client is moving to a new buffer much like above
			// Validate we have no listeners on the old one
			continue
		case closeUpdate:
			// We're moving back to an old buffer. Only update
			// The active status on the buffer we're moving to
			continue
		case linkUpdate:
			// We're renaming a buffer outright
			continue
		}
	}
}
