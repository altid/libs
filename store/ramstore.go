package store

import (
	"github.com/altid/libs/store/internal/ramstore"
)

// Forward errors from internal
const (
	ErrInvalidTrunc = ramstore.ErrInvalidTrunc
	ErrShortSeek    = ramstore.ErrShortSeek
	ErrActiveStream = ramstore.ErrActiveStream
	ErrDirExists    = ramstore.ErrDirExists
	ErrFileClosed   = ramstore.ErrFileClosed
)

// RamStore is a type that implements Filer as an in-memory data store
type RamStore struct {
	root *ramstore.Dir
}

func NewRamStore(debug bool) *RamStore {
	return &RamStore{
		root: ramstore.NewRoot(debug),
	}
}

func (rs *RamStore) List() []string {
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

func (rs *RamStore) Type() string {
	return "ram"
}

func (rs *RamStore) Mkdir(dir string) error {
	return rs.root.Mkdir(dir)
}
