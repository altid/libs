package files

import (
	"errors"
	"os"
)

// Handler represents calls to a synthetic file
type Handler interface {
	Stat(msg *Message) (os.FileInfo, error)
	Normal(msg *Message) (interface{}, error)
}

// Files is the main
type Files struct {
	run     map[string]Handler
	path    string
	service string
}

// Message contains information about which file is being requested
type Message struct {
	Service string
	Buffer  string
	Target  string
}

// Handle listens for calls to its Stat and Normal functions, and returns a stat or os.File
// Writes and Reads to real files will be rooted at `dir`
// Many resulting files are synthesized on demand, others map to a real file that may or may not
// be rooted at the given directory
// For example, to send a command to open a buffer
//
// 		h := Handle("/path/to/my/server")
//
//		fp, err := h.Normal("mybuffer", "ctl")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		defer fp.Close()
//		fp.WriteString("open foo")
func Handle(dir string) *Files {
	// TODO(halfwit) split dir
	return &Files{
		service: dir,
		run:     make(map[string]Handler),
	}
}

// Add a Handler to the stack
func (f *Files) Add(path string, h Handler) {
	f.run[path] = h
}

// Stat will synthesize an os.FileInfo (stat) for the named file, if available
func (f *Files) Stat(buffer, req string) (os.FileInfo, error) {
	h, ok := f.run[req]
	if !ok {
		return nil, errors.New("Unable to find handler for named file")
	}

	msg := &Message{
		Service: f.service,
		Buffer:  buffer,
		Target:  req,
	}

	return h.Stat(msg)
}

// Normal will return an interface satisfying io.ReaderAt, io.WriterAt, and io.Closer if the file requested is a regular file
// If the file requested is a directory, it will synthesize an *os.FileInfo with the correct semantics
func (f *Files) Normal(buffer, req string) (interface{}, error) {
	h, ok := f.run[req]
	if !ok {
		return nil, errors.New("Unable to find handler for named file")
	}

	msg := &Message{
		Service: f.service,
		Buffer:  buffer,
		Target:  req,
	}

	return h.Normal(msg)
}
