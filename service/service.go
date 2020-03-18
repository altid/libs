package main

// We'll have to hold the handler to the tab state in service, and pass it down to tabs when it wants to read on it.
// This should be mostly fine regardless

// So we need access to the commands as they're made
// commands
// clients... maybe?
// list of client buffers
// Then from tabs, we can call 

import (
	"log"
	"os"
	"sync"
)

type cmdKey int

const (
	bufferCmd cmdKey = iota
	reloadCmd
	linkCmd
	openCmd
	closeCmd
	startCmd
	checkCmd
	sendFeedCmd
	sendEOFCmd
)

type service struct {
	chatty   func(cmdKey, ...string)
	debug    func(string, ...interface{})
	commands chan *cmd
	clients  map[int64]*client
	tablist  map[string]*tabs
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
			// Eventually this should go away.
			log.Printf("Unable to add service %s, no tabs file found", svc)
			continue
		}

		chlog := serviceChatlog
		if !*chatty {
			chlog = func(cmdKey, ...string) {}
		}

		dblog := serviceDebugLog
		if !*debug {
			dblog = func(string, ...interface{}) {}
		}
		addr := cfg.getAddress(svc)

		service := &service{
			debug:    dblog,
			chatty:   chlog,
			clients:  make(map[int64]*client),
			commands: make(chan *cmd),
			tablist:  tlist,
			addr:     addr,
			name:     svc,
		}

		chlog(startCmd, addr, svc)

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
			s.chatty(reloadCmd, s.name)
			s.addr = cfg.getAddress(s.name)
		case bufferCmd:
			s.move(cl, cmd.value)
			s.chatty(sendEOFCmd, s.name, cl.current)
			s.sendFeedEOF(cl.uuid)
		case openCmd:
			s.open(cl, cmd.value)
			s.chatty(sendEOFCmd, s.name, cl.current)
			s.sendFeedEOF(cl.uuid)
		case closeCmd:
			s.close(cl)
			s.chatty(sendEOFCmd, s.name, cl.current)
			s.sendFeedEOF(cl.uuid)
		case linkCmd:
			s.chatty(linkCmd)
			s.close(cl)
			s.open(cl, cmd.value)
			s.chatty(sendEOFCmd, s.name, cl.current)
			s.sendFeedEOF(cl.uuid)
		}
	}
}

func (s *service) open(c *client, name string) {
	s.checkInactive(c)
	s.chatty(openCmd, name)
	c.current = name
}

func (s *service) close(c *client) {
	delete(s.tablist, c.current)
	s.chatty(closeCmd, c.current)

	for _, cl := range s.clients {
		if cl.current != c.current {
			continue
		}

		// Grab first item
		for _, t := range s.tablist {
			s.chatty(bufferCmd, c.current, t.name)
			cl.current = t.name
			c.current = t.name
			t.active = true
			t.alert = false
			t.count = 0
			break
		}
	}

}

func (s *service) move(c *client, name string) {
	old := &client{
		current: c.current,
		uuid:    c.uuid,
	}

	defer s.checkInactive(old)

	if name == "none" {
		s.chatty(bufferCmd, c.current, "none")
		c.current = "none"
		return
	}

	t, ok := s.tablist[name]
	if !ok {

		t = &tabs{
			name: name,
		}
		s.tablist[name] = t
	}

	t.active = true
	t.alert = false
	t.count = 0
	s.chatty(bufferCmd, c.current, name)
	c.current = name
}

func (s *service) checkInactive(c *client) {
	for _, cl := range s.clients {
		if cl.uuid == c.uuid {
			continue
		}

		// At least one listener, no need to update
		if cl.current == c.current {
			s.chatty(checkCmd, "found", c.current)
			return
		}
	}

	s.chatty(checkCmd, "notfound", c.current)
	if t, ok := s.tablist[c.current]; ok {
		t.active = false
	}
}

func (s *service) sendFeed() {
	for _, cl := range s.clients {
		s.chatty(sendFeedCmd, s.name, cl.current)
		cl.feed <- struct{}{}
	}
}

func (s *service) sendFeedEOF(uuid int64) {
	close(s.clients[uuid].feed)
	s.clients[uuid].feed = make(chan struct{})
}

func serviceChatlog(key cmdKey, args ...string) {
	// Set logger format to match
	l := log.New(os.Stdout, "", 0)

	cmd := '↓'
	switch key {
	case bufferCmd:
		l.Printf("%c CMD Buffer from=\"%s\" to=\"%s\"", cmd, args[0], args[1])
	case reloadCmd:
		l.Printf("%c CMD Reload service=%s", cmd, args[0])
	case linkCmd:
		l.Printf("%c CMD Link", cmd)
	case openCmd:
		l.Printf("%c CMD Open buffer=\"%s\"", cmd, args[0])
	case closeCmd:
		l.Printf("%c CMD Close buffer=\"%s\"", cmd, args[0])
	}

	cmd = '!'
	switch key {
	case startCmd:
		l.Printf("%c INF Start address=%s name=%s", cmd, args[0], args[1])
	case checkCmd:
		if args[0] == "found" {
			l.Printf("%c INF Check buffer=\"%s\" active=true", cmd, args[1])
			return
		}
		l.Printf("%c INF Check buffer=\"%s\" active=false", cmd, args[1])
	case sendFeedCmd:
		l.Printf("%c INF Efeed op=event target=\"%s/%s\"", cmd, args[0], args[1])
	case sendEOFCmd:
		l.Printf("%c INF Efeed op=eof target=\"%s/%s\"", cmd, args[0], args[1])
	}
}

func serviceDebugLog(format string, v ...interface{}) {
	l := log.New(os.Stdout, "9pd: ", 0)
	l.Printf(format, v...)
}
