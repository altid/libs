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
	cert     string
	key      string
	factotum bool
	usetls   bool
	services map[string]*service
}

func NewSettings(debug, chatty bool, cert, key, path string, factotum, usetls bool) *Settings {
	return &Settings{
		debug:    debug,
		chatty:   chatty,
		cert:     cert,
		key:      key,
		path:     path,
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
		events, err := tail.WatchEvents(ctx, s.path, entry)
		if err != nil && s.debug {
			serviceDebugLog("%v", err)
			continue
		}

		tabs, err := tabs.FromFile(path.Join(s.path, entry))
		if err != nil && s.debug {
			serviceDebugLog("%v", err)
			continue
		}

		listen, port := config.GetListenAddress(entry)

		srv := &service{
			command:  make(chan *command.Command),
			client:   &client.Manager{},
			name:     entry,
			cert:     s.cert,
			key:      s.key,
			tabs:     tabs,
			listen:   listen,
			port:     port,
			events:   events,
			basedir:  s.path,
			log:      s.debug,
			chatty:   s.chatty,
			tls:      s.usetls,
			factotum: s.factotum,
			debug:    func(string, ...interface{}) {},
		}

		srv.files, srv.feed = registerFiles(srv)

		if s.debug {
			srv.debug = serviceDebugLog
		}

		services[entry] = srv
	}

	s.services = services
	return nil
}
