package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"
)

type CtrlFile struct {
	cb     func([]byte) error
	data   []byte
	offset int64
}

// Each ctrl open should have potentially new data of available commands
// So we send callbacks to generate from the current good commandlist
// back in the session.go file
func Ctrl(cb func([]byte) error, data func() []byte) (*CtrlFile, error) {
	cf := &CtrlFile{
		cb:   cb,
		data: data(),
	}

	return cf, nil
}

// Support seeking for potential large control writes
func (c *CtrlFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		c.offset = offset
	case io.SeekCurrent:
		c.offset += offset
	case io.SeekEnd:
		c.offset = 0 + offset
	}

	if c.offset < 0 {
		return 0, fmt.Errorf("attempted to seek before start of file")
	}

	if c.offset > 0 {
		return 0, io.EOF
	}

	return c.offset, nil
}

// Reads return our ctrl message group
func (c *CtrlFile) Read(b []byte) (n int, err error) {
	if (c.offset) >= int64(len(c.data)) {
		return 0, io.EOF
	}

	n = copy(b, c.data)
	c.offset += int64(n)
	return
}

func (c *CtrlFile) Write(p []byte) (n int, err error) {
	n = len(p)
	c.offset += int64(n)
	err = c.cb(p)

	return
}

func (c *CtrlFile) Truncate(cap int64) error {
	if cap > c.offset {
		return errors.New("truncation on file requested was larger than file")
	}

	c.offset = cap
	return nil
}

func (c *CtrlFile) Close() error {
	c.offset = 0
	return nil
}

func (c *CtrlFile) Name() string { return "/ctrl" }
func (c *CtrlFile) Stat() (fs.FileInfo, error) {
	cs := &CtrlStat{
		name:    "/ctrl",
		size:    c.offset,
		modtime: time.Now(),
	}

	return cs, nil
}

type CtrlStat struct {
	name    string
	size    int64
	modtime time.Time
}

func (s *CtrlStat) Name() string       { return s.name }
func (s *CtrlStat) Sys() any           { return nil }
func (s *CtrlStat) ModTime() time.Time { return s.modtime }
func (s *CtrlStat) IsDir() bool        { return false }
func (s *CtrlStat) Mode() os.FileMode  { return 0644 }
func (s *CtrlStat) Size() int64        { return s.size }
