package files

import (
	"os"
	"path"

	"github.com/altid/server/files"
)

// Nothing needs to be synthesized here
// Just make sure we grab the top level file

type ErrHandler struct{}

func NewError() *ErrHandler { return &ErrHandler{} }

func (*ErrHandler) Normal(msg *files.Message) (interface{}, error) {
	fp := path.Join(msg.Service, "errors")
	return os.Open(fp)
}

func (*ErrHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	fp := path.Join(msg.Service, "errors")
	return os.Lstat(fp)
}
