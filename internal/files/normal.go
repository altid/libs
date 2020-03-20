package files

import (
	"os"
	"path"

	"github.com/altid/server/files"
)

type NormalHandler struct{}

func NewNormal() *NormalHandler { return &NormalHandler{} }

func (*NormalHandler) Normal(msg *files.Message) (interface{}, error) {
	fp := path.Join(msg.Server, msg.Buffer, msg.Target)
	return os.Open(fp)
}

func (*NormalHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	fp := path.Join(msg.Server, msg.Buffer, msg.Target)
	return os.Lstat(fp)
}
