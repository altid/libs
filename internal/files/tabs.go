package files

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/altid/server/files"
	"github.com/altid/server/tabs"
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

func (th *TabsHandler) Normal(msg *files.Message) (interface{}, error) {
	var b bytes.Buffer

	for _, tab := range th.manager.List() {
		if tab.Alert {
			b.WriteString("!")
		}

		fmt.Fprintf(&b, "%s [%d]\n", tab.Name, tab.Count)
	}

	t := &tab{
		size: int64(b.Len()),
		data: b.Bytes(),
	}

	return t, nil
}

func (*TabsHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	return os.Stat(path.Join(msg.Service, "tabs"))
}
func (t *tab) ReadAt(p []byte, off int64) (n int, err error) {
	n = copy(p, t.data[off:])
	if int64(n)+off > t.size {
		return n, io.EOF
	}

	return
}

func (t *tab) WriteAt(p []byte, off int64) (int, error) {
	return 0, errors.New("tabs file does not allow modification")
}
