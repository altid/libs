package ninep

import (
	"context"

	"github.com/altid/libs/config"
	"github.com/altid/server/client"
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
	services map[string]*service
}

func NewSettings(debug, chatty bool, path string, port int, factotum, usetls bool) *Settings {
	return &Settings{
		debug:    debug,
		chatty:   chatty,
		path:     path,
		port:     port,
		factotum: factotum,
		usetls:   usetls,
	}
}

func (s *Settings) registerFiles(t *tabs.Manager, e chan *tail.Event, c chan *command.Command) *files.Manager {
	h := files.Handle(s.path)

	h.Add("/", my.NewDir())
	h.Add("/ctl", my.NewCtl(c))
	h.Add("/error", my.NewError())
	h.Add("/feed", my.NewFeed(e))
	h.Add("/input", my.NewInput())
	h.Add("/tabs", my.Tabs(t))
	h.Add("default", my.NewNormal())

	return h
}

func (s *Settings) BuildServices(ctx context.Context) error {
	services := make(map[string]*service)

	for _, entry := range config.ListAll() {
		events, err := tail.WatchEvents(ctx, s.path, entry.Name)
		if err != nil {
			return err
		}

		tabs := tabs.Manager{}
		commands := make(chan *command.Command)
		feed := make(chan struct{})
		files := s.registerFiles(tabs, events, commands)

		srv := &service{
			client:   &client.Manager{},
			files:    files,
			config:   entry,
			tabs:     tabs,
			commands: commands,
			events:   events,
			feed:     feed,
			basedir:  s.path,
			log:      s.debug,
			chatty:   s.chatty,
			tls:      s.usetls,
			factotum: s.factotum,
			debug:    func(string, ...interface{}) {},
		}

		if s.debug {
			srv.debug = serviceDebugLog
		}

		services[entry.Name] = srv
	}

	s.services = services
	return nil
}
