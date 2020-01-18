package main

import (
	"log"
)

type updateKey int

const (
	bufferUpdate updateKey = iota
	openUpdate
	closeUpdate
)

type service struct {
	state   chan *update
	clients map[int64]*client
	tabs    map[string]*tab
	addr    string
	name    string
}

type client struct {
	target  string
	reading string
	current string
}

type update struct {
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
			clients: make(map[int64]*client),
			state:   make(chan *update),
			tabs:    tabs,
			addr:    cfg.getAddress(svc),
			name:    svc,
		}
		service.watch()
		services[svc] = service
	}
	return services
}

func (s *service) watch() {
	for update := range s.state {
		switch update.key {
		case bufferUpdate:
			// A client is switching buffers. Go through
			// and update all unread counts to reflect this
			// Mark old buffer as inactive if there are no
			// more readers on it
		case openUpdate:
			// A client is moving to a new buffer much like above
			// Validate we have no listeners on the old on
		case closeUpdate:
			// We're moving back to an old buffer. Only update
			// The active status on the buffer we're moving to
		}
	}
}
