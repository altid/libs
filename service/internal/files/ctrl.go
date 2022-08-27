package files

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"
)

type CtrlFile struct {
	Current chan string
	offset  int64
}

func Ctrl() (*CtrlFile, error) {
	cf := &CtrlFile{
		Current: make(chan string),
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

func (c *CtrlFile) Read(b []byte) (n int, err error) {
	return 0, errors.New("reading not implemented on ctrl")
}

func (c *CtrlFile) Write(p []byte) (n int, err error) {
	n = len(p)
	c.offset += int64(n)

	go func(c *CtrlFile, p []byte) {
		if bytes.HasPrefix(p, []byte("buffer ")) {
			buffer := bytes.TrimPrefix(p, []byte("buffer "))
			c.Current <- string(buffer)
			return
		}
	}(c, p)

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
	close(c.Current)
	c.offset = 0
	return nil
}

func (c *CtrlFile) Stream() (io.ReadCloser, error) {
	return nil, errors.New("streams not supported on control file")
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
func (s *CtrlStat) Sys() interface{}   { return nil }
func (s *CtrlStat) ModTime() time.Time { return s.modtime }
func (s *CtrlStat) IsDir() bool        { return false }
func (s *CtrlStat) Mode() os.FileMode  { return 0644 }
func (s *CtrlStat) Size() int64        { return s.size }
