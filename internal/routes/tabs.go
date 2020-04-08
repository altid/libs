package routes

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/altid/server/internal/message"
	"github.com/altid/server/internal/tabs"
)

type TabsHandler struct {
	// Put handler to tabs here
	manager *tabs.Manager
}

type tab struct {
	data []byte
	size int64
}

func NewTabs(manager *tabs.Manager) *TabsHandler { return &TabsHandler{manager} }

func (th *TabsHandler) Normal(msg *message.Message) (interface{}, error) {
	var b bytes.Buffer

	for _, t := range th.manager.List() {
		fmt.Fprintf(&b, "%s\n", t.String())
	}

	t := &tab{
		size: int64(b.Len()),
		data: b.Bytes(),
	}

	return t, nil
}

func (*TabsHandler) Stat(msg *message.Message) (os.FileInfo, error) {
	return os.Stat(path.Join(msg.Service, "tabs"))
}
func (t *tab) ReadAt(p []byte, off int64) (n int, err error) {
	if off > t.size {
		return n, io.EOF
	}

	n = copy(p, t.data[off:])
	if int64(n)+off > t.size {
		return n, io.EOF
	}

	return
}

func (t *tab) WriteAt(p []byte, off int64) (int, error) {
	return 0, errors.New("tabs file does not allow modification")
}
