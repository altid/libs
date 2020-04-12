package files

import (
	"os"

	"github.com/altid/server/command"
	"github.com/altid/server/internal/message"
	"github.com/altid/server/internal/routes"
	"github.com/altid/server/internal/tabs"
)

// Handler represents calls to a synthetic file
type Handler interface {
	Stat(msg *message.Message) (os.FileInfo, error)
	Normal(msg *message.Message) (interface{}, error)
}

// Files facilitates access to the functions of the sythetic files
type Files struct {
	run     map[string]Handler
	service string
}

func NewFiles(dir string, cmd chan *command.Command, tabs *tabs.Manager) *Files {
	run := make(map[string]Handler)

	run["/"] = routes.NewDir()
	run["/ctl"] = routes.NewCtl(cmd)
	run["/errors"] = routes.NewError()
	run["/input"] = routes.NewInput()
	run["/tabs"] = routes.NewTabs(tabs)
	run["default"] = routes.NewNormal()
	run["/feed"] = routes.NewFeed()

	return &Files{
		service: dir,
		run: run,
	}
}

// Stat will synthesize an os.FileInfo (stat) for the named file, if available
func (f *Files) Stat(buffer, req string, uuid uint32) (os.FileInfo, error) {
	h, ok := f.run[req]
	if !ok {
		h = f.run["default"]
	}

	msg := &message.Message{
		Service: f.service,
		Buffer:  buffer,
		Target:  req,
		UUID:    uuid,
	}

	return h.Stat(msg)
}

// Normal will return an interface satisfying io.ReaderAt, io.WriterAt, and io.Closer if the file requested is a regular file
// If the file requested is a directory, it will synthesize an *os.FileInfo with the correct semantics
func (f *Files) Normal(buffer, req string, uuid uint32) (interface{}, error) {
	h, ok := f.run[req]
	if !ok {
		h = f.run["default"]
	}

	msg := &message.Message{
		Service: f.service,
		Buffer:  buffer,
		Target:  req,
		UUID:    uuid,
	}

	return h.Normal(msg)
}
