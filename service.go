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
	clients  []*client
	tablist  map[string]*tab
	addr     string
	name     string
	sync.Mutex
}

type client struct {
	feed    chan struct{}
	target  string
	reading string
	current string
	uuid    int64
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
			commands: make(chan *cmd),
			tablist:  tlist,
			addr:     cfg.getAddress(svc),
			name:     svc,
		}

		go service.watch(cfg)
		services[svc] = service
	}

	return services
}

func (s *service) watch(cfg *config) {
	for cmd := range s.commands {
		switch cmd.key {
		case reloadCmd:
			s.addr = cfg.getAddress(s.name)
		case bufferCmd:
			// A client is switching buffers. Go through
			// and update all unread counts to reflect this
			// Mark old buffer as inactive if there are no
			// more readers on it

			// We close feed so that all readers can send the EOF
			for _, cl := range s.clients {
				if cl.uuid != cmd.uuid {
					continue
				}
				s.Lock()
				close(cl.feed)
				cl.feed = make(chan struct{})
				s.Unlock()
			}

			continue
		case openCmd:
			// A client is moving to a new buffer much like above
			// Validate we have no listeners on the old one
			continue
		case closeCmd:
			// We're moving back to an old buffer. Only update
			// The active status on the buffer we're moving to
			continue
		case linkCmd:
			// We're renaming a buffer outright
			continue
		}
	}
}
