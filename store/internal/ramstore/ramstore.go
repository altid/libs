package ramstore

// TODO: Decouple the data from a "file", the File is just an accessor to the data with offsets
// TODO: sync.RWMutex is better for many-reader performance
import (
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"time"
)

var l *log.Logger

type storeMsg int

const (
	storeErr storeMsg = iota
	storeMkdir
	storeStream
	storeRoot
	storeDir
	storeData
	storeReadStream
	storeOpen
)

type errRamstore string

const (
	ErrInvalidTrunc = errRamstore("truncation invalid")
	ErrInvalidPath  = errRamstore("invalid path supplied to Mkdir")
	ErrShortSeek    = errRamstore("attempted negative seek")
	ErrActiveStream = errRamstore("attempted to close a file with active streams")
	ErrDirExists    = errRamstore("directory exists")
	ErrFileClosed   = errRamstore("invalid action on closed file")
)

// This is not great for multiple readers/writers - needs to be reworked
// Very large structures will cause fairly large allocations due to the recursive maps here
type Dir struct {
	name  string
	dirs  map[string]*Dir
	files map[string]*File
	debug func(storeMsg, ...interface{})
}

// Internal data store
type store struct {
	bytes []byte
}

func NewRoot(debug bool) *Dir {
	d := &Dir{
		name:  "/",
		dirs:  make(map[string]*Dir),
		files: make(map[string]*File),
		debug: func(storeMsg, ...interface{}) {},
	}

	if debug {
		d.debug = storeLogger
		l = log.New(os.Stdout, "store ", 0)
	}

	return d
}

func (d *Dir) Mkdir(name string) error {
	if _, ok := d.dirs[name]; ok {
		return ErrDirExists
	}

	if path.Clean(name) != name {
		d.debug(storeErr, ErrInvalidPath)
		return ErrInvalidPath
	}

	dir := &Dir{
		name:  name,
		dirs:  make(map[string]*Dir),
		files: make(map[string]*File),
		debug: d.debug,
	}

	d.debug(storeMkdir, name)
	d.dirs[name] = dir
	return nil
}

func (d *Dir) List() []string {
	var list []string

	for dname, dir := range d.dirs {
		for fname := range dir.files {
			list = append(list, path.Join("/", dname, fname))
		}
	}

	for _, f := range d.files {
		list = append(list, f.Name())
	}

	return list
}

func (d *Dir) Root(buffer string) (*File, error) {
	// Clean the path
	buffer = path.Join("/", buffer)

	f := &File{
		path:    buffer,
		data:    nil,
		offset:  0,
		isdir:   true,
		closed:  false,
		modTime: time.Now(),
		readdir: make(chan fs.FileInfo, 10),
		done:    make(chan struct{}),
		debug:   d.debug,
	}

	go listRoot(d, f, buffer)
	d.debug(storeRoot, buffer)
	return f, nil
}

// Open works by either returning a file/directory, or recursing if we are still rooted in a path
func (d *Dir) Open(name string) (*File, error) {
	// Use strings split os.pathesparator
	// switch on the len of that array
	// do d[path[0]], make if missing up to n times
	// like for i := 0; i < len(path); i++
	// then build out the dirs if they miss recursively
	// grab the final file
	// Or even i < len(path) - 1, do the final file/dir thing after
	if _, ok := d.dirs[name]; ok {
		return d.Root(name)
	}

	// For example, `/errors` or `/tabs`
	if f, ok := d.files[name]; ok {
		return copyFile(f)
	}

	// Say we look up `#altid/feed`
	base := path.Base(name)
	for _, val := range d.dirs {
		if f, ok := val.files[path.Join("/", base)]; ok {
			return copyFile(f)
		}
	}

	// We're here, we need a new dir with a files entry
	data := &store{
		bytes: make([]byte, 256),
	}

	f := &File{
		path:    path.Join("/", base),
		data:    data,
		offset:  0,
		closed:  false,
		isdir:   false,
		modTime: time.Now(),
		debug:   d.debug,
	}

	// Create the directory we need, and assign our file in it
	// If we're on the base dir, simply add to our top level files and return
	root := path.Dir(name)
	if root == "/" {
		d.files[name] = f
		return f, nil
	}

	d.Mkdir(root)
	d.dirs[root].files[path.Join("/", base)] = f
	//d.debug(storeOpen, f)

	return f, nil
}

func (d *Dir) Delete(name string) error {
	// TODO: Walk the dirs and delete the entry if found
	return nil
}

type File struct {
	path    string
	data    *store
	offset  int64
	isdir   bool
	closed  bool
	modTime time.Time
	readdir chan os.FileInfo
	done    chan struct{}
	debug   func(storeMsg, ...interface{})
}

func (f *File) Read(b []byte) (n int, err error) {
	if f.closed {
		f.debug(storeErr, ErrFileClosed)
		return 0, ErrFileClosed
	}

	if f.offset >= int64(len(f.data.bytes)) {
		return 0, io.EOF
	}

	n = copy(b, f.data.bytes[f.offset:])
	f.offset += int64(n)
	return n, nil
}

func (f *File) Write(p []byte) (n int, err error) {
	if f.closed {
		f.debug(storeErr, ErrFileClosed)
		return 0, ErrFileClosed
	}

	if int64(len(f.data.bytes)+len(p)) >= f.offset {
		tmp := make([]byte, (cap(f.data.bytes)*2)+len(p))
		copy(tmp, f.data.bytes)
		f.data.bytes = tmp
	}

	var i int64
	for i = 0; i < int64(len(p)); i++ {
		f.data.bytes[f.offset] = p[i]
		f.offset++
	}

	f.modTime = time.Now()
	return len(p), nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		f.debug(storeErr, ErrFileClosed)
		return 0, ErrFileClosed
	}

	switch whence {
	case io.SeekStart:
		f.offset = offset
	case io.SeekCurrent:
		f.offset += offset
	case io.SeekEnd:
		f.offset = int64(len(f.data.bytes)) + offset
	}

	if f.offset < 0 {
		f.debug(storeErr, ErrShortSeek)
		return 0, ErrShortSeek
	}

	if f.offset > int64(len(f.data.bytes)) {
		return 0, io.EOF
	}

	return f.offset, nil
}

func (f *File) Close() error {
	if f.closed {
		f.debug(storeErr, ErrFileClosed)
		return ErrFileClosed
	}

	if f.isdir {
		close(f.done)
	}

	f.offset = 0
	f.closed = true
	return nil
}

func (f *File) Truncate(cap int64) error {
	if cap > int64(len(f.data.bytes)) {
		f.debug(storeErr, ErrInvalidTrunc)
		return ErrInvalidTrunc
	}

	if cap < 0 {
		f.debug(storeErr, ErrInvalidTrunc)
		return ErrInvalidTrunc
	}

	// Just remake the data and set the cap
	f.data.bytes = f.data.bytes[:cap]
	return nil
}

func (f *File) Name() string {
	return f.path
}

func (f *File) Stat() (fs.FileInfo, error) {
	if !f.isdir {
		fi := FileInfo{
			len:     int64(len(f.data.bytes)),
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

// We want to return our file wrapper to allow multiple simultaneous reader/writers
func copyFile(in *File) (*File, error) {
	f := &File{
		path:    in.path,
		data:    in.data,
		offset:  0,
		closed:  false,
		isdir:   in.isdir,
		modTime: in.modTime,
		debug:   in.debug,
	}

	return f, nil
}

func listRoot(d *Dir, root *File, buffer string) {
	var list []fs.FileInfo
	for _, file := range d.files {
		fi := &FileInfo{
			len:     int64(len(file.data.bytes)),
			name:    file.path,
			modtime: file.modTime,
		}

		list = append(list, fi)
	}

	// Then go into the buffer dir
	if dir, ok := d.dirs[buffer]; ok {
		for _, file := range dir.files {
			fi := &FileInfo{
				len:     int64(len(file.data.bytes)),
				name:    path.Join("/", path.Base(file.path)),
				modtime: file.modTime,
			}

			list = append(list, fi)
		}
	}

	// Early exit if we have nothing
	if len(list) == 0 {
		return
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
func (e errRamstore) Error() string    { return string(e) }

func storeLogger(msg storeMsg, args ...interface{}) {
	switch msg {
	case storeErr:
		l.Printf("error: %s", args[0])
	case storeMkdir:
		l.Printf("Mkdir: %s", args[0])
	case storeStream:
		l.Printf("stream: starting for %s", args[0])
	case storeRoot:
		l.Printf("root: created at %s", args[0])
	case storeData:
		l.Printf("incoming data: file=\"%s\" %s", args[0], args[1])
	case storeDir:
		l.Printf("opening dir: %s", args[0])
	case storeReadStream:
		l.Printf("stream: reading initial data for %s: %s", args[0], args[1])
	case storeOpen:
		if f, ok := args[0].(*File); ok {
			l.Printf("open: name=\"%s\"", f.path)
		}
	}
}
