package ninep

import (
	"context"
	"path"

	"github.com/altid/libs/config"
	"github.com/altid/server/client"
	"github.com/altid/server/command"
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

func (s *Settings) BuildServices(ctx context.Context) error {
	services := make(map[string]*service)

	list, err := config.ListAll()
	if err != nil {
		return err
	}

	for _, entry := range list {
		events, err := tail.WatchEvents(ctx, s.path, entry.Name)
		if err != nil && s.debug {
			serviceDebugLog("%s", err)
			continue
		}

		tabs, err := tabs.FromFile(path.Join(s.path, entry.Name))
		if err != nil && s.debug {
			serviceDebugLog("%s", err)
			continue
		}

		commands := make(chan *command.Command)
		feed := make(chan struct{})
		files := registerFiles(tabs, feed, commands, path.Join(s.path, entry.Name))

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
