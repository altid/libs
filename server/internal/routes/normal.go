package routes

import (
	"os"
	"path"

	"github.com/altid/server/internal/message"
)

type NormalHandler struct{}

func NewNormal() *NormalHandler { return &NormalHandler{} }

func (*NormalHandler) Normal(msg *message.Message) (interface{}, error) {
	fp := path.Join(msg.Service, msg.Buffer, msg.Target)
	return os.Open(fp)
}

func (*NormalHandler) Stat(msg *message.Message) (os.FileInfo, error) {
	fp := path.Join(msg.Service, msg.Buffer, msg.Target)
	return os.Lstat(fp)
}
