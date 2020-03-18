package files

import (
	"os"
	"path"

	"github.com/altid/server/files"
)

type normalHandler struct{}

func (*normalHandler) Normal(msg *files.Message) (interface{}, error) {
	fp := path.Join(msg.Server, msg.Buffer, msg.Target)
	return os.Open(fp)
}

func getNormalStat(msg *files.Message) (os.FileInfo, error) {
	fp := path.Join(msg.Server, msg.Buffer, msg.Target)
	return os.Lstat(fp)
}
