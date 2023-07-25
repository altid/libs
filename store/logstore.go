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
type Logstore struct {
	base  string
	root  *ramstore.Dir
	mains map[string]fs.FileInfo
}

func NewLogstore(base string, debug bool) *Logstore {
	os.MkdirAll(base, 0777)
	return &Logstore{
		base:  base,
		root:  ramstore.RootDir(debug),
		mains: make(map[string]fs.FileInfo),
	}
}

func (ls *Logstore) List() []string {
	list := ls.root.List()
	for dir, main := range ls.mains {
		list = append(list, path.Join(dir, main.Name()))
	}
	return list
}

func (ls *Logstore) Open(name string) (File, error) {
	// Check if our path ends with "/main"
	if path.Base(name) == "main" || path.Base(name) == "feed" {
		os.MkdirAll(path.Join(ls.base, path.Dir(name)), 0777)
		f, err := os.OpenFile(path.Join(ls.base, name), os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return nil, err
		}
		if _, ok := ls.mains[path.Dir(name)]; !ok {
			ls.mains[path.Dir(name)], err = f.Stat()
		}
		return f, err
	}
	return ls.root.Open(name)
}

func (ls *Logstore) Delete(name string) error {
	if path.Base(name) == "main" || path.Base(name) == "feed" {
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

func (ls *Logstore) Root(name string) (Dir, error) { return ls.root.Root(name) }
func (ls *Logstore) Type() string                  { return "log" }
func (ls *Logstore) Mkdir(name string) error       { return ls.root.Mkdir(name) }
