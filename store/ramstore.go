package store

import (
	"github.com/altid/libs/store/internal/ramstore"
)

// RamStore is a type that implements Filer as an in-memory data store
type RamStore struct {
	root *ramstore.Dir
}

func NewRamStore() *RamStore {
	return &RamStore {
		root: ramstore.NewRoot(),
	}
}

func (rs *RamStore) List() ([]string) {
	return rs.root.List()
}

func (rs *RamStore) Root(name string) (File, error) {
	return rs.root.Root(name)
}

func (rs *RamStore) Open(name string) (File, error) {
	return rs.root.Open(name)
}

func (rs *RamStore) Delete(path string) error {
	return rs.root.Delete(path)
}