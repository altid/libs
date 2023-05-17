package ramstore

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

type errRamstore string
const (
	ErrInvalidTrunc = errRamstore("truncation invalid")
	ErrInvalidPath  = errRamstore("invalid path supplied")
	ErrInvalidDir   = errRamstore("invalid directory supplied")
	ErrShortSeek    = errRamstore("attempted negative seek")
	ErrSeekOver     = errRamstore("attempted seek past end of file")
	ErrDirExists    = errRamstore("directory exists")
	ErrFileClosed   = errRamstore("invalid action on closed file")
)

var l *log.Logger

type dirMsg int
const (
	dirErr dirMsg = iota
	dirMkdir
	dirRoot
	dirDir
	dirData
	dirReaddir
	dirOpen
	dirInfo
)

type Dir struct {
	name 	string
	path	string
	files	map[string]any
	readdir	chan os.FileInfo
	done	chan struct{}
	debug	func(dirMsg, ...any)
}

func RootDir(debug bool) *Dir {
	d := &Dir{
		name:		"/",
		path:		"/",
		files:		make(map[string]any),
		debug: 		func(dirMsg, ...any) {},
	}

	if debug {
		d.debug = dirLogger
		l = log.New(os.Stdout, "store: directory: ", 0)
	}

	return d
}

func (d *Dir) Walk(name string) (any, error) {
	if name == "/" || name == "." {
		return d, nil
	}
	paths := strings.Split(name, string(os.PathSeparator))
	for _, a := range d.files {
		d.debug(dirInfo, paths)
		switch v := a.(type) {
		case *Dir:
			if v.name == paths[0] {
				if len(paths) > 1 {
					return v.Walk(path.Join( paths[1:]...))
				}
				return a, nil
			}
		case *File:
			if v.Name() == paths[0] {
				v.closed = false
				a = v
				return a, nil
			}
		}
	}

	return nil, ErrInvalidPath
}

func (d *Dir) Mkdir(name string) error {
	d.debug(dirInfo, "mkdir", name)
	v, err := d.Walk(path.Dir(name))
	if err != nil {
		return err
	}
	// check if file exists
	if dir, ok := v.(*Dir); ok {
		// Check if there's already something at that path
		if _, ok := dir.files[name]; ok {
			return ErrDirExists
		}
		// Good dir, make new file
		l := &Dir{
			name:		path.Base(name),
			path:		name,
			files:		make(map[string]any),
			debug:		d.debug,
		}
		d.debug(dirMkdir, name)
		dir.files[name] = l
		return nil
	}

	return ErrInvalidDir
}

func (d *Dir) List() []string {
	var list []string
	for _, a := range d.files {
		switch v := a.(type) {
		case *Dir:
			l := v.List()
			list = append(list, l...)
		case *File:
			list = append(list, v.path)
		}
	}
	return list
}

func (d *Dir) Root(buffer string) (*Dir, error) {
	d.readdir = make(chan fs.FileInfo)
	go listRoot(d, buffer)
	d.debug(dirRoot, buffer)
	return d, nil
}

// Open works by either returning a file/directory, or recursing if we are still rooted in a path
func (d *Dir) Open(name string) (*File, error) {
	d.debug(dirInfo, "opening file/dir", name)
	a, err := d.Walk(path.Dir(name))
	if err != nil {
		return nil, ErrInvalidPath
	}
LOOP:
	// Good file, we can return
	switch v := a.(type) {
	case *Dir:
		// Look and see if we have a good file/dir
		// This may be broken because path strips the leading path
		if nv, ok := v.files[path.Base(name)]; ok {
			a = nv
			goto LOOP
		}
		f := &File{
			path:	name,
			name:	path.Base(name),
			data:   &data{},
			offset: 0,
			closed: false,
			modTime: time.Now(),
		}
		v.files[path.Base(name)] = f
		return f, nil
	case *File:
		v.closed = false
		return v, nil
	default:
		return nil, ErrInvalidPath
	}
}

func (d *Dir) Close() error { return nil }

func (d *Dir) Delete(name string) error {
	// TODO: Walk the dirs and delete the entry if found
	return nil
}

func (d *Dir) Readdir(n int) ([]fs.FileInfo, error) {
	d.debug(dirReaddir, d.path)
	var err error
	fi := make([]os.FileInfo, 0, 10)
	for i := 0; i < n; i++ {
		s, ok := <-d.readdir
		if !ok {
			err = io.EOF
			break
		}
		fi = append(fi, s)
	}

	return fi, err

}

func listRoot(d *Dir, buffer string) {
	// List fileinfo for our complete root
	var list []fs.FileInfo
	for _, a := range d.files {
		switch v := a.(type) {
		case *Dir:

			if v.path == buffer {
				for _, b := range v.files {

					if f, ok := b.(*File); ok {
						// Stat won't fail as we're operating on our map of files
						stat, _ := f.Stat()
						list = append(list, stat)
					}
				}
			}
		case *File:
			fi := &FileInfo{
				len:     int64(len(v.data.bytes)),
				name:    v.path,
				modtime: v.modTime,
			}
	
			list = append(list, fi)
		}
	}

	go func([]os.FileInfo, *Dir) {
		for _, item := range list {
			select {
			case d.readdir <- item:
			case <-d.done:
				goto FINISH
			}
		}
	FINISH:
		close(d.readdir)
	}(list, d)
}

func (d *Dir) Info() (fs.FileInfo, error) {
	di := &DirInfo{
		name: d.name,
		modtime: time.Now(),
	}
	return di, nil
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
func (di DirInfo) Sys() any			  { return nil }

func dirLogger(msg dirMsg, args ...any) {
	switch msg {
	case dirErr:
		l.Printf("error: %s", args[0])
	case dirMkdir:
		l.Printf("mkdir: %s", args[0])
	case dirRoot:
		l.Printf("root: created at %s", args[0])
	case dirReaddir:
		l.Printf("Readdir entered with %s", args[0])
	case dirData:
		l.Printf("incoming data: file=\"%s\" %s", args[0], args[1])
	case dirDir:
		l.Printf("opening dir: %s", args[0])
	case dirOpen:
		if f, ok := args[0].(*File); ok {
			l.Printf("open: name=\"%s\"", f.path)
		}
	case dirInfo:
		l.Printf("info: %v", args)
	}
}
