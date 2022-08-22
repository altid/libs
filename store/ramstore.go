package store

import (
	"fmt"

	"github.com/altid/libs/store/internal/ramstore"
)

// RamStore is a type that implements Filer as an in-memory data store
type RamStore struct {
	files map[string]*ramstore.File
}

func NewRamStore() *RamStore {
	return &RamStore{
		files: make(map[string]*ramstore.File),
	}
}

func (rs *RamStore) List() ([]string) {
	var list []string
	for _, file := range rs.files {
		list = append(list, file.Name())
	}

	return list
}

func (rs *RamStore) Open(path string) (File, error) {
	f, ok := rs.files[path]
	if !ok || f == nil {
		f = ramstore.Open(path)
		rs.files[path] = f
		return f, nil
	}

	return f, nil
}

func (rs *RamStore) Delete(path string) error {
	f, ok := rs.files[path]
	if !ok {
		return fmt.Errorf("no file exists at path %s", path)
	}

	if f.InUse() {
		return fmt.Errorf("attempting to delete an active file")
	}

	delete(rs.files, path)
	return nil
}