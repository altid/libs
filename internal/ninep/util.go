package ninep

import (
	"path"

	"github.com/altid/server/files"
	"github.com/altid/server/internal/routes"
)

func registerFiles(s *service) (*files.Files, *routes.FeedHandler) {
	h := files.Handle(path.Join(s.basedir, s.name))
	fh := routes.NewFeed()

	h.Add("/", routes.NewDir())
	h.Add("/ctl", routes.NewCtl(s.command))
	h.Add("/error", routes.NewError())
	h.Add("/input", routes.NewInput())
	h.Add("/tabs", routes.NewTabs(s.tabs))
	h.Add("default", routes.NewNormal())
	h.Add("/feed", fh)

	return h, fh
}
