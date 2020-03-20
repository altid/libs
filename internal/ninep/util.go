package ninep

import (
	"os"
	"path"

	"github.com/altid/server/command"
	"github.com/altid/server/files"
	"github.com/altid/server/internal/routes"
)

func registerFiles(s *service) *files.Files {
	h := files.Handle(path.Join(s.basedir, s.config.Name))

	file := path.Join(s.basedir, s.config.Name, "ctl")
	fp, _ := os.OpenFile(file, os.O_APPEND, 0644)

	// Feed and ctl are the only ones talking to eachother, we don't need tracking elsewhere
	commands := make(chan *command.Command)

	h.Add("/", routes.NewDir())
	h.Add("/ctl", routes.NewCtl(commands))
	h.Add("/error", routes.NewError())
	h.Add("/feed", routes.NewFeed(s.client, commands, fp))
	h.Add("/input", routes.NewInput())
	h.Add("/tabs", routes.NewTabs(s.tabs))
	h.Add("default", routes.NewNormal())

	return h
}
