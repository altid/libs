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
type Ramstore struct {
	Base *ramstore.Dir
}

func NewRamstore(debug bool) *Ramstore {
	return &Ramstore{
		Base: ramstore.RootDir(debug),
	}
}

// Returns a dir or file, or any errors encountered
func (rs *Ramstore) Walk(name string) (any, error) {
	return rs.Base.Walk(name)
}

func (rs *Ramstore) List() []string {
	return rs.Base.List()
}

func (rs *Ramstore) Root(name string) (File, error) {
	return rs.Base.Root(name)
}

func (rs *Ramstore) Open(name string) (File, error) {
	return rs.Base.Open(name)
}

func (rs *Ramstore) Delete(path string) error {
	return rs.Base.Delete(path)
}

func (rs *Ramstore) Type() string {
	return "ram"
}

func (rs *Ramstore) Mkdir(dir string) error {
	return rs.Base.Mkdir(dir)
}
