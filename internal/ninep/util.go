package ninep

import (
	"github.com/altid/server/command"
	"github.com/altid/server/files"
	"github.com/altid/server/internal/routes"
	"github.com/altid/server/tabs"
)

func registerFiles(t *tabs.Manager, e chan struct{}, c chan *command.Command, service string) *files.Files {
	h := files.Handle(service)

	h.Add("/", routes.NewDir())
	h.Add("/ctl", routes.NewCtl(0, c))
	h.Add("/error", routes.NewError())
	h.Add("/feed", routes.NewFeed(e))
	h.Add("/input", routes.NewInput())
	h.Add("/tabs", routes.NewTabs(t))
	h.Add("default", routes.NewNormal())

	return h
}
