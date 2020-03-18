package files

import (
	"os"
	"path"

	"github.com/altid/server/files"
)

// Nothing needs to be synthesized here
// Just make sure we grab the top level file

type errHandler struct{}

func (*errHandler) Normal(msg *files.Message) (interface{}, error) {
	fp := path.Join(msg.Service, "errors")
	return os.Open(fp)
}

func (*errHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	fp := path.Join(msg.Service, "errors")
	return os.Lstat(fp)
}
