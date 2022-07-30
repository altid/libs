package ramstore

import (
	"fmt"
	"io"

	"github.com/google/uuid"
)


type Stream struct {
	data chan []byte
	done chan struct{}
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
	path	string
	data	[]byte
	offset	int64
	closed  bool
	streams map[string]*Stream
}

func Open(path string) *File {
	return &File{
		path: 	path,
		data: 	[]byte(""),
		offset: 0,
		closed: false,
		streams: make(map[string]*Stream),
	}
}

func (f *File) Read(b []byte) (n int, err error) {
	if int64(len(f.data)) < f.offset {
		return 0, io.EOF
	}

	if f.closed {
		return 0, fmt.Errorf("Attempted to read on closed file")
	}

	n = copy(b, f.data)

	f.offset += int64(n)
	if f.offset >= int64(len(f.data)) {
		return n, io.EOF
	}

	return n, nil
}

func (f *File) Write(p []byte) (n int, err error) {
	if f.closed {
		return 0, fmt.Errorf("Attemted to write on a closed file")
	}

	// Write to all the open Streams
	for _, c := range f.streams {
		go func(c *Stream, p []byte) {
			// Guard against close channel race condition
			for {
				select {
				case c.data <- p:
					return
				case <-c.done:
					return
				}
			}
		}(c, p)
	}

	f.data = append(f.data, p...)
	n = len(p)
	f.offset += int64(n)

	return n, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		return 0, fmt.Errorf("Attempted to seek on a closed file")
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
		return 0, fmt.Errorf("Attempted to seek before start of file")
	}

	if f.offset > int64(len(f.data)) {
		return 0, fmt.Errorf("Attempted to seek past end of file")
	}

	return f.offset, nil
}

func (f *File) Close() error {
	f.closed = true
	return nil
}

func (f *File) Stream() (io.ReadCloser, error) {
	uuid := uuid.New()
	s := &Stream{
		f: f,
		uuid: uuid.String(), 
		done: make (chan struct{}),
		data: make (chan []byte),
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

func (f *File) Path() string {
	return f.path
}