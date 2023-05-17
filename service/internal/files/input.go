package files

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/altid/libs/markup"
	"github.com/altid/libs/service/callback"
)

// Callback - forward handler from listener Registration
// Call back in Write
type InputFile struct {
	handle callback.Handler
	buffer string
	offset int64
}

func Input(buffer string, handle callback.Handler) (*InputFile, error) {
	i := &InputFile{
		buffer: buffer,
		handle: handle,
	}

	return i, nil
}

func (i *InputFile) Read(b []byte) (n int, err error) {
	return 0, errors.New("reads not supported on input")
}

func (i *InputFile) Write(p []byte) (n int, err error) {
	n = len(p)
	i.offset += int64(n)
	c := markup.NewLexer(p)
	err = i.handle.Handle(i.buffer, c)
	return
}

// Support seeking for potential large control writes
func (i *InputFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		i.offset = offset
	case io.SeekCurrent:
		i.offset += offset
	case io.SeekEnd:
		i.offset = 0 + offset
	}

	if i.offset < 0 {
		return 0, fmt.Errorf("attempted to seek before start of file")
	}

	if i.offset > 0 {
		return 0, io.EOF
	}

	return i.offset, nil
}

func (i *InputFile) Close() error {
	i.offset = 0
	return nil
}

func (i *InputFile) Truncate(cap int64) error {
	if cap > i.offset {
		return errors.New("truncation on file requested was larger than file")
	}

	i.offset = cap
	return nil
}

func (i *InputFile) Name() string { return "input" }
func (i *InputFile) Stat() (fs.FileInfo, error) {
	is := &InputStat{
		name:    "input",
		size:    0,
		modtime: time.Now(),
	}

	return is, nil
}

type InputStat struct {
	name    string
	size    int64
	modtime time.Time
}

func (s *InputStat) Name() string       { return s.name }
func (s *InputStat) Sys() any           { return nil }
func (s *InputStat) ModTime() time.Time { return s.modtime }
func (s *InputStat) IsDir() bool        { return false }
func (s *InputStat) Mode() os.FileMode  { return 0644 }
func (s *InputStat) Size() int64        { return s.size }
