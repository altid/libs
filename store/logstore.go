package store

// TODO: logging
import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/altid/libs/store/internal/ramstore"
)

// We need to include streamer on this
type LogStore struct {
	base  string
	root  *ramstore.Dir
	mains map[string]fs.File
}

func NewLogStore(base string, debug bool) *LogStore {
	os.MkdirAll(base, 0777)
	return &LogStore{
		base:  base,
		root:  ramstore.NewRoot(debug),
		mains: make(map[string]fs.File),
	}

}

func (ls *LogStore) List() []string {
	list := ls.root.List()

	for _, main := range ls.mains {
		st, _ := main.Stat()
		list = append(list, st.Name())
	}

	return list
}

func (ls *LogStore) Root(name string) (File, error) {
	return ls.root.Root(name)
}

func (ls *LogStore) Open(name string) (File, error) {
	// Check if our path ends with "/main"
	if path.Base(name) == "main" || path.Base(name) == "feed" {
		os.MkdirAll(path.Join(ls.base, path.Dir(name)), 0777)
		return os.OpenFile(path.Join(ls.base, name), os.O_RDWR|os.O_CREATE, 0755)
	}

	return ls.root.Open(name)
}

func (ls *LogStore) Delete(name string) error {
	if path.Base(name) == "main" || path.Base(name) == "feed" {
		f, ok := ls.mains[name]
		if !ok {
			return fmt.Errorf("no file exists at path %s", name)
		}

		st, _ := f.Stat()
		os.Remove(st.Name())
		delete(ls.mains, name)
		return nil
	}

	return ls.root.Delete(name)
}

func (ls *LogStore) Type() string {
	return "log"
}

func (ls *LogStore) Mkdir(name string) error {
	return ls.root.Mkdir(name)
}
