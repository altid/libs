package logstore

import (
	"io"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/google/uuid"
)

// Wrap an os.File to support streaming
type Stream struct {
	data chan []byte
	done chan struct {}
	uuid string
	f	 *File
}

func (s *Stream) Read(b []byte) (n int, err error) {
	for {
		select {
		case inc := <- s.data:
			n = copy(b, inc)
			return n, nil
		case <- s.done:
			return 0, io.EOF
		}
	}
}

func (s *Stream) Close() error {
	close(s.done)
	close(s.data)

	delete(s.f.streams, s.uuid)
	return	nil
}

type File struct {
	f		*os.File
	streams map[string]*Stream
}

func Open(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &File{
		f: f,
		streams: make(map[string]*Stream),
	}, nil
}

func (f *File) Stream() (io.ReadCloser, error) {
	uuid := uuid.New()
	s := &Stream{
		f: f,
		uuid: uuid.String(),
		done: make (chan struct{}),
		data: make (chan []byte),
	}

	data, err := ioutil.ReadAll(f.f)
	if err != nil {
		return nil, err
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
	}(s, data)

	f.streams[s.uuid] = s
	return s, nil
}

// Wrap the rest of our interface to the raw files
func (f *File) Read(b []byte) (int, error) { return f.f.Read(b) }
func (f *File) Write(p []byte) (int, error) { return f.f.Write(p) }
func (f *File) Seek(offset int64, whence int) (int64, error) { return f.f.Seek(offset, whence )}
func (f *File) Close() error { return f.f.Close() }
func (f *File) Name() string { return f.f.Name() }
func (f *File) Stat() (fs.FileInfo, error) { return f.f.Stat() }
