package store

import (
	"fmt"

	"github.com/altid/libs/service/internal/ramstore"
)

// RamStor is a type that implements Filer as an in-memory data store
type RamStore struct {
	files map[string]*ramstore.File
}

func NewRamStorage() *RamStore {
	return &RamStore{
		files: make(map[string]*ramstore.File),
	}
}

func (rm *RamStore) List() ([]string) {
	var list []string
	for _, file := range rm.files {
		list = append(list, file.path)
	}

	return list
}

func (rm *RamStore) Open(path string) (File, error) {
	f, ok := rm.files[path]
	if !ok || f == nil {
		f = ramstore.Open(path)
		rm.files[path] = f
		return f, nil
	}

	return f, nil
}

func (rm *RamStore) Delete(path string) error {
	f, ok := rm.files[path]
	if !ok {
		return fmt.Errorf("No file exists at path %s", path)
	}

	if !f.closed {
		return fmt.Errorf("Attempting to delete an open file")
	}

	if len(f.streams) > 0 {
		return fmt.Errorf("Attempting to delete a file with active streams")
	}

	delete(rm.files, path)
	return nil
}