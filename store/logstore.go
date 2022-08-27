package store

import (
	"fmt"
	"os"
	"path"

	"github.com/altid/libs/store/internal/logstore"
	"github.com/altid/libs/store/internal/ramstore"
)

// We need to include streamer on this
type LogStore struct {
	base  string
	root  *ramstore.Dir
	mains map[string]*logstore.File
}

func NewLogStore(base string) *LogStore {
	return &LogStore{
		base:  base,
		root:  ramstore.NewRoot(),
		mains: make(map[string]*logstore.File),
	}

}

func (ls *LogStore) List() []string {
	list := ls.root.List()

	for _, main := range ls.mains {
		list = append(list, main.Name())
	}

	return list
}

func (ls *LogStore) Root(name string) (File, error) {
	return ls.root.Root(name)
}

func (ls *LogStore) Open(name string) (File, error) {
	// Check if our path ends with "/main"
	if path.Base(name) == "main" {
		os.MkdirAll(path.Join(ls.base, path.Dir(name)), 0777)
		return logstore.Open(path.Join(ls.base, name))
	}

	return ls.root.Open(name)
}

func (ls *LogStore) Delete(name string) error {
	if path.Base(name) == "main" {
		f, ok := ls.mains[name]
		if !ok {
			return fmt.Errorf("no file exists at path %s", name)
		}

		os.Remove(f.Name())
		delete(ls.mains, name)
		return nil
	}

	return ls.root.Delete(name)
}

func (ls *LogStore) Type() string {
	return "log"
}
