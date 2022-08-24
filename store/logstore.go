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
	base string
	files map[string]*ramstore.File
	mains map[string]*logstore.File
}

func NewLogStore(base string) *LogStore {
	return &LogStore{
		base: base,
		files: make(map[string]*ramstore.File),
		mains: make(map[string]*logstore.File),
	}

}

func (ls *LogStore) List() ([]string) {
	var list []string
	for _, file := range ls.files {
		list = append(list, file.Name())
	}
	for _, main := range ls.mains {
		list = append(list, main.Name())
	}

	return list
}


func (ls *LogStore) Open(name string) (File, error) {
	// Check if our path ends with "/main"
	if path.Base(name) == "main" {
		os.MkdirAll(path.Join(ls.base, path.Dir(name)), 0777)
		return logstore.Open(path.Join(ls.base, name))
	}

	f, ok := ls.files[name]
	if !ok || f == nil {
		f = ramstore.Open(name)
		ls.files[name] = f
		return f, nil
	}
	return f, nil
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

	f, ok := ls.files[name]
	if !ok {
		return fmt.Errorf("to file exists at path %s", name)
	}
	
	if f.InUse() {
		return fmt.Errorf("attempting to delete an active file")
	}
	
	delete(ls.files, name)
	return nil
}