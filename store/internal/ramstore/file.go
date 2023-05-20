package ramstore

import (
	"io"
	"io/fs"
	"os"
	"time"
)

type File struct {
	path    string
	name	string
	data	*data
	offset  int
	closed  bool
	modTime time.Time
	//debug   func(fileMsg, ...any)
}

func (f *File) Read(b []byte) (n int, err error) {
	if f.closed {
		return 0, ErrFileClosed
	}
	f.data.RLock()
	defer f.data.RUnlock()
	n = copy(b, f.data.bytes[f.offset:])
	f.offset += n
	if f.offset >= len(f.data.bytes) {
		return n, io.EOF
	}
	return n, nil
}

func (f *File) Write(p []byte) (n int, err error) {
	if f.closed {
		return 0, ErrFileClosed
	}
	f.data.Lock()
	defer f.data.Unlock()
	m := f.data.grow(f.offset, len(p))
	n = copy(f.data.bytes[m:], p)
	f.offset += n
	f.modTime = time.Now()

	return
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		return 0, ErrFileClosed
	}
	switch whence {
	case io.SeekStart:
		f.offset = int(offset)
	case io.SeekCurrent:
		f.offset += int(offset)
	case io.SeekEnd:
		f.offset = len(f.data.bytes) + int(offset)
	}

	if f.offset < 0 {
		//f.debug(storeErr, ErrShortSeek)
		return 0, ErrShortSeek
	}

	if f.offset > len(f.data.bytes) {
		f.offset = len(f.data.bytes)
	}

	return int64(f.offset), nil
}

func (f *File) Close() error {
	f.closed = true
	return nil
}

func (f *File) Truncate(cap int64) error {
	if f.closed {
		return ErrFileClosed
	}
	if cap > int64(len(f.data.bytes)) {
		//f.debug(storeErr, ErrInvalidTrunc)
		return ErrInvalidTrunc
	}
	// Make sure we make a fresh byte array on 0 truncation
	if cap == 0 {
		// Make a new buffer
		f.data.bytes = make([]byte, cap, 4096)
		return nil
	}
	// Just remake the data and set the cap
	f.data.Lock()
	defer f.data.Unlock()
	f.data.bytes = f.data.bytes[:cap]
	return nil
}

func (f *File) Name() string {
	return f.path
}

func (f *File) Stat() (fs.FileInfo, error) {
	fi := FileInfo{
		len:     int64(len(f.data.bytes)),
		name:    f.name,
		modtime: f.modTime,
	}

	return fi, nil
}

type FileInfo struct {
	len     int64
	name    string
	modtime time.Time
}

func (fi FileInfo) Size() int64        { return fi.len }
func (fi FileInfo) Name() string       { return fi.name }
func (fi FileInfo) IsDir() bool        { return false }
func (fi FileInfo) ModTime() time.Time { return fi.modtime }
func (fi FileInfo) Mode() os.FileMode  { return 0644 }
func (fi FileInfo) Sys() any		   { return nil }
func (e errRamstore) Error() string    { return string(e) }
 