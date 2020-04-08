package services

import (
	"context"
	"crypto/tls"
	"path"

	"github.com/altid/libs/config"
	"github.com/altid/server/client"
	"github.com/altid/server/files"
	"github.com/altid/server/internal/command"
	"github.com/altid/server/internal/routes"
	"github.com/altid/server/internal/tabs"
	"github.com/altid/server/internal/tail"
	"github.com/altid/server/settings"
)

type Runner interface {
	Run(context.Context, *files.Files) error
}

type Service struct {
	ctx      context.Context
	Client   *client.Manager
	Files    *files.Files
	Tabs     *tabs.Manager
	Feed     *routes.FeedHandler
	Command  chan *command.Command
	Events   chan *tail.Event
	Cert     tls.Certificate
	Basedir  string
	Listen   string
	Name     string
	Port     string
	Log      bool
	Factotum bool
	Debug    func(string, ...interface{})
}

func FromSettings(ctx context.Context, s *settings.Settings) (map[string]*Service, error) {
	services := make(map[string]*Service)

	list, err := config.ListAll()
	if err != nil {
		return nil, err
	}

	for _, entry := range list {
		events, err := tail.WatchEvents(ctx, s.Path, entry)
		if err != nil {
			continue
		}

		tabs, err := tabs.FromFile(path.Join(s.Path, entry))
		if err != nil {
			continue
		}

		listen, port := config.GetListenAddress(entry)

		srv := &Service{
			Command:  make(chan *command.Command),
			Client:   &client.Manager{},
			Name:     entry,
			Tabs:     tabs,
			Listen:   listen,
			Port:     port,
			Events:   events,
			Basedir:  s.Path,
			Cert:     s.Cert,
			Factotum: s.Factotum,
			ctx:      ctx,
		}

		srv.Files = files.NewFiles(s.Path, srv.Command, srv.Tabs)
		services[entry] = srv
	}

	return services, nil
}

func (s *Service) Run(run Runner) error {
	return run.Run(s.ctx, s.Files)
}
