package ramstore

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

// This is not great for multiple readers/writers - needs to be reworked
// Very large structures will cause fairly large allocations due to the recursive maps here
type Dir struct {
	name  string
	dirs  map[string]*Dir
	files map[string]*File
}

func NewRoot() *Dir {
	return &Dir{
		name:  "/",
		dirs:  make(map[string]*Dir),
		files: make(map[string]*File),
	}
}

// List of files and dirs; dirs get a trailing slash
func (d *Dir) List() []string {
	var list []string

	for _, dir := range d.dirs {
		list = append(list, dir.name+string(os.PathSeparator))
	}

	for _, f := range d.files {
		list = append(list, f.Name())
	}

	return list
}

func (d *Dir) Root(buffer string) (*File, error) {
	var err error
	f := &File{
		path:    "/",
		data:    []byte(""),
		offset:  0,
		isdir:   true,
		closed:  false,
		modTime: time.Now(),
		readdir: make(chan fs.FileInfo, 10),
		done:    make(chan struct{}),
	}

	go listRoot(d, f, buffer)
	return f, err
}

// Open works by either returning a file/directory, or recursing if we are still rooted in a path
func (d *Dir) Open(name string) (*File, error) {
	paths := strings.Split(name, string(os.PathSeparator))
	// Ignore leading slashes
	if paths[0] == "" {
		paths = paths[1:]
	}

	// Fix root pathing
	if name == "/" {
		paths[0] = "/"
	}

	// We have a base-level file for the given dir
	// We have to assume it's a regular file and not a directory
	if len(paths) == 1 && name != "/" {
		file, ok := d.files[paths[0]]
		if ok {
			file.closed = false
			return file, nil
		}

		f := &File{
			path:    name,
			data:    []byte(""),
			offset:  0,
			closed:  false,
			isdir:   false,
			streams: make(map[string]*Stream),
			modTime: time.Now(),
		}

		d.files[paths[0]] = f
		return f, nil
	}

	// If we have a good entry, return it
	dir, ok := d.dirs[paths[0]]
	if ok {
		// We're still nested, recurse
		if len(paths) > 1 {
			paths[0] = "/"
			name = path.Join(paths...)
			return dir.Open(name)
		}

		f := &File{
			path:    paths[0],
			data:    []byte(""),
			offset:  0,
			isdir:   true,
			closed:  false,
			modTime: time.Now(),
			readdir: make(chan fs.FileInfo, 10),
			done:    make(chan struct{}),
		}

		go listDir(d, f)
		return f, nil
	}

	// Accidental files that were supposed to be dirs can populate here
	// Make sure we get rid of 'em before we create a dir
	delete(d.files, paths[0])
	wd := &Dir{
		name:  paths[0],
		dirs:  make(map[string]*Dir),
		files: make(map[string]*File),
	}

	//if (d.name != "/") {
	d.dirs[paths[0]] = wd
	//}

	if len(paths) > 1 {
		paths[0] = "/"
		name = path.Join(paths...)
		return wd.Open(name)
	}

	f := &File{
		path:    paths[0],
		data:    []byte(""),
		offset:  0,
		isdir:   true,
		closed:  false,
		modTime: time.Now(),
		readdir: make(chan fs.FileInfo, 10),
		done:    make(chan struct{}),
	}

	go listDir(d, f)
	return f, nil
}

func (d *Dir) Delete(name string) error {
	// TODO: Walk the dirs and delete the entry if found
	return nil
}

type Stream struct {
	data chan []byte
	done chan struct{}
	uuid string
	f    *File
}

func (s *Stream) Read(b []byte) (n int, err error) {
	for {
		select {
		case inc := <-s.data:
			n = copy(b, inc)
			return n, nil
		case <-s.done:
			return 0, io.EOF
		}
	}
}

func (s *Stream) Close() error {
	close(s.done)
	close(s.data)

	delete(s.f.streams, s.uuid)
	return nil
}

type File struct {
	path    string
	data    []byte
	offset  int64
	isdir   bool
	closed  bool
	streams map[string]*Stream
	modTime time.Time
	readdir chan os.FileInfo
	done    chan struct{}
}

func (f *File) Read(b []byte) (n int, err error) {
	if f.closed {
		return 0, fmt.Errorf("attempted to read on closed file")
	}

	if int64(len(f.data)) < f.offset {
		return 0, io.EOF
	}

	f.modTime = time.Now()
	n = copy(b, f.data)

	f.offset += int64(n)
	if f.offset >= int64(len(f.data)) {
		return n, io.EOF
	}

	return n, nil
}

func (f *File) Write(p []byte) (n int, err error) {
	if f.closed {
		return 0, fmt.Errorf("attemted to write on a closed file")
	}

	// Enable seek
	f.data = append(f.data[:f.offset], p...)
	n = len(p)
	f.offset += int64(n)
	f.modTime = time.Now()

	// Write to all the open Streams
	for _, c := range f.streams {
		go func(c *Stream, f *File) {
			// Guard against close channel race condition
			for {
				select {
				case c.data <- f.data:
					return
				case <-c.done:
					return
				}
			}
		}(c, f)
	}

	return n, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		return 0, fmt.Errorf("attempted to seek on a closed file")
	}

	switch whence {
	case io.SeekStart:
		f.offset = offset
	case io.SeekCurrent:
		f.offset += offset
	case io.SeekEnd:
		f.offset = int64(len(f.data)) + offset
	}

	if f.offset < 0 {
		return 0, fmt.Errorf("attempted to seek before start of file")
	}

	if f.offset > int64(len(f.data)) {
		return 0, io.EOF
	}

	return f.offset, nil
}

func (f *File) InUse() bool {
	return len(f.streams) > 0
}

func (f *File) Close() error {
	if f.closed {
		return fmt.Errorf("attempted to close a file which was already closed")
	}

	if len(f.streams) > 0 {
		return fmt.Errorf("attempted to close a file with active streams")
	}

	if f.isdir {
		close(f.done)
	}

	f.offset = 0
	f.closed = true
	return nil
}

func (f *File) Truncate(cap int64) error {
	if cap > int64(len(f.data)) {
		return fmt.Errorf("truncation beyond size of file")
	}

	if cap < 0 {
		return fmt.Errorf("truncation to less than 0 invalid")
	}

	f.data = f.data[:cap]
	return nil
}

func (f *File) Stream() (io.ReadCloser, error) {
	uuid := uuid.New()
	s := &Stream{
		f:    f,
		uuid: uuid.String(),
		done: make(chan struct{}),
		data: make(chan []byte),
	}

	// Load out the initial, existing data to the stream
	// Don't continue to block if the ReadCloser is closed
	go func(s *Stream, data []byte) {
		for {
			select {
			case s.data <- data:
				return
			case <-s.done:
				return
			}
		}
	}(s, f.data)

	f.streams[s.uuid] = s
	return s, nil
}

func (f *File) Name() string {
	return f.path
}

func (f *File) Stat() (fs.FileInfo, error) {
	if !f.isdir {
		fi := FileInfo{
			len:     int64(len(f.data)),
			name:    f.path,
			modtime: f.modTime,
		}

		return fi, nil
	}

	di := &DirInfo{
		name:    f.Name(),
		modtime: time.Now(),
	}

	return di, nil
}

// Sometimes you have to use an ugly global

func (f *File) Readdir(n int) ([]fs.FileInfo, error) {
	var err error

	fi := make([]os.FileInfo, 0, 10)
	for i := 0; i < n; i++ {
		s, ok := <-f.readdir
		if !ok {
			err = io.EOF
			break
		}
		fi = append(fi, s)
	}

	return fi, err

}

func listRoot(d *Dir, root *File, buffer string) {
	var list []fs.FileInfo
	for _, file := range d.files {
		fi := &FileInfo{
			len:     int64(len(file.data)),
			name:    file.path,
			modtime: file.modTime,
		}

		list = append(list, fi)
	}

	// Then go into the buffer dir
	if dir, ok := d.dirs[buffer]; ok {
		for _, file := range dir.files {
			fi := &FileInfo{
				len:     int64(len(file.data)),
				name:    path.Join("/", path.Base(file.path)),
				modtime: file.modTime,
			}

			list = append(list, fi)
		}
	}

	go func([]os.FileInfo, *File) {
		for _, d := range list {
			select {
			case root.readdir <- d:
			case <-root.done:
				goto FINISH
			}
		}
	FINISH:
		close(root.readdir)
	}(list, root)
}

func listDir(d *Dir, f *File) {
	var list []fs.FileInfo
	for _, file := range d.files {
		fi := &FileInfo{
			len:     int64(len(file.data)),
			name:    file.path,
			modtime: file.modTime,
		}

		list = append(list, fi)
	}

	for _, dir := range d.dirs {
		fi := &DirInfo{
			name:    dir.name,
			modtime: time.Now(),
		}

		list = append(list, fi)
	}

	go func([]os.FileInfo, *File) {
		for _, d := range list {
			select {
			case f.readdir <- d:
			case <-f.done:
				goto FINISH
			}
		}
	FINISH:
		close(f.readdir)
	}(list, f)
}

type DirInfo struct {
	name    string
	modtime time.Time
}

func (di DirInfo) Size() int64        { return 0 }
func (di DirInfo) Name() string       { return di.name }
func (di DirInfo) IsDir() bool        { return true }
func (di DirInfo) ModTime() time.Time { return di.modtime }
func (di DirInfo) Mode() os.FileMode  { return os.ModeDir }
func (di DirInfo) Sys() interface{}   { return nil }

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
func (fi FileInfo) Sys() interface{}   { return nil }
