package services

import (
	"context"
	"os"
	"path"

	"github.com/altid/libs/config"
	"github.com/altid/server/client"
	"github.com/altid/server/command"
	"github.com/altid/server/files"
	"github.com/altid/server/internal/routes"
	"github.com/altid/server/internal/tabs"
	"github.com/altid/server/internal/tail"
)

type Service struct {
	ctx     context.Context
	Files   *files.Files
	Tabs    *tabs.Manager
	Client  *client.Manager
	Feed    *routes.FeedHandler
	Command chan *command.Command
	Events  chan *tail.Event
	Basedir string
	Name    string
	Log     bool
	Debug   func(string, ...interface{})
}

func FindServices(ctx context.Context, dir string) (map[string]*Service, error) {
	services := make(map[string]*Service)

	list, err := config.ListAll()
	if err != nil {
		return nil, err
	}

	for _, entry := range list {
		sdir := path.Join(dir, entry)

		events, err := tail.WatchEvents(ctx, dir, entry)
		if err != nil {
			continue
		}

		tabs, err := tabs.FromFile(sdir)
		if err != nil {
			continue
		}

		srv := &Service{
			Command: make(chan *command.Command),
			Client:  &client.Manager{},
			Feed:    routes.NewFeed(),
			Name:    entry,
			Tabs:    tabs,
			Events:  events,
			Basedir: dir,
			ctx:     ctx,
			Debug:   func(string, ...interface{}) {},
		}

		srv.Files = files.NewFiles(sdir, srv.Command, srv.Tabs, srv.Feed)
		services[entry] = srv

		ctlpath := path.Join(srv.Basedir, srv.Name, "ctl")
		ctl, err := os.OpenFile(ctlpath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		go srv.listenCommands(ctl)
		go srv.listenEvents()
	}

	return services, nil
}
