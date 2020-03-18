package files

import (
	"os"
	"path"

	"github.com/altid/server/files"
)

type inputHandler struct{}

func (*inputHandler) Normal(msg *files.Message) (interface{}, error) {
	fp := path.Join(msg.Server, msg.Buffer, "input")
	i := &input{
		path: fp,
	}

	return i, nil
}

func (*inputHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	return os.Stat(path.Join(msg.Server, msg.Buffer, "input"))
}

type input struct {
	path string
}

// Simple wrapper around an open call
func (i *input) ReadAt(b []byte, off int64) (n int, err error) {
	fp, err := os.OpenFile(i.path, os.O_RDONLY, 0600)
	if err != nil {
		return
	}

	defer fp.Close()
	return fp.ReadAt(b, off)
}

// Open in correct modes
func (i *input) WriteAt(p []byte, off int64) (n int, err error) {
	fp, err := os.OpenFile(i.path, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}

	defer fp.Close()
	return fp.Write(p)
}

func (i *input) Close() error { return nil }
