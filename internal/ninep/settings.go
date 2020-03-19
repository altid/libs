package ninep

import (
	"github.com/altid/server/command"
	"github.com/altid/server/files"
	my "github.com/altid/server/internal/files"
	"github.com/altid/server/tabs"
	"github.com/altid/server/tail"
)

type Settings struct {
	debug    bool
	chatty   bool
	path     string
	port     int
	factotum bool
	usetls   bool
	subsets map[string]*subset
}

type subset struct {
	handler *files.Files
	tabs *tabs.Manager
	command chan *command.Command
	events chan *tail.Event
	feed chan struct{}
}
// *debug, *chatty, *dir, *port, *factotum, *usetls
func NewSettings(debug, chatty bool, path string, port int, factotum, usetls bool) *Settings {
	return &Settings{
		debug:   debug,
		chatty:  chatty,
		path:    path,
		port: port,
		factotum: factotum,
		usetls: usetls,
		subset: make(map[string]*subset)
	}
}


// BuildServices must be called before any of the underlying types are used
func (s *Settings) BuildServices() error {
	for _, entry := range config.ListAll() {
		v := &subset{}
		s.subsets[entry] = v
	}
}

func (s *Settings) registerFiles(svc string) error {
	h := files.Handle(s.path)

	h.Add("/", my.NewDir())
	h.Add("/ctl", my.NewCtl(s.command[svc]))
	h.Add("/error", my.NewError())
	h.Add("/feed", my.NewFeed(s.feed[svc])
	h.Add("/input", my.NewInput())
	h.Add("/tabs", my.Tabs(s.tabs[svc]))
	h.Add("default", my.NewNormal())

	s.handlers = h
	return nil
}

func (s *Settings) listServices() map[string]*service {
	services := make(map[string]*service)

	for name, set := range s.subset {
		conf, _ := config.New(nil, name, false)
		srv := &service{
			client: &client.Manager{},
			files: set.handler,
			config: conf,
			tabs: set.tabs,
			commands: set.command,
			events: set.events,
			feed: set.feed,
			basedir: set.path,
			log: set.debug,
			chatty: set.chatty,
			tls: set.tls,
			factotum: set.factotum,
			debug: func(string, ...interface{}){},
		}

		if set.debug {
			srv.debug = serviceDebugLogging
		}

		s.registerFiles(srv)
		services[name] = srv
	}


	return services
}
