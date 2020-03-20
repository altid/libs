package routes

import (
	"errors"
	"io"
	"os"
	"path"
	"sync"

	"github.com/altid/server/files"
)

// FeedHandler wraps a `feed` file with special blocking read semantics
type FeedHandler struct {
	feeds map[uint32]chan struct{}
	sync.Mutex
}

// NewFeed returns a FeedHandler
func NewFeed() *FeedHandler {
	fh := &FeedHandler{
		feeds: make(map[uint32]chan struct{}),
	}

	return fh
}

// Send an event to the feed file specified by the uuid, if one exists
func (fh *FeedHandler) Send(uuid uint32) {
	fh.Lock()
	defer fh.Unlock()

	if f, ok := fh.feeds[uuid]; ok {
		f <- struct{}{}
	}
}

// Done sends EOF for any blocking reads to a client on a `feed`
func (fh *FeedHandler) Done(uuid uint32) {
	if _, ok := fh.feeds[uuid]; ok {
		fh.Lock()
		defer fh.Unlock()

		close(fh.feeds[uuid])
		delete(fh.feeds, uuid)
	}
}

// Normal returns a readwritecloser tied to "feed", which tails it
// until a new file is read by the client
func (fh *FeedHandler) Normal(msg *files.Message) (interface{}, error) {
	event := make(chan struct{})
	f := &feed{
		event: event,
		path:  path.Join(msg.Service, msg.Buffer, "feed"),
		buff:  path.Join(msg.Service, msg.Buffer),
	}

	// Feed to match this specific client
	fh.feeds[msg.UUID] = event

	return f, nil

}

// Stat returns a normal stat to an underlying feed file
func (*FeedHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	return os.Lstat(path.Join(msg.Service, msg.Buffer, "feed"))
}

// feed files are special in that they're blocking
type feed struct {
	tailing bool
	path    string
	buff    string
	event   chan struct{}
}

func (f *feed) ReadAt(p []byte, off int64) (n int, err error) {
	fp, err := os.Open(f.path)
	if err != nil {
		return 0, err
	}

	defer fp.Close()

	if !f.tailing {
		n, err = fp.ReadAt(p, off)
		if err != nil && err != io.EOF {
			return
		}

		if err == io.EOF {
			f.tailing = true
		}
		return n, nil
	}

	for range f.event {
		n, err = fp.ReadAt(p, off)
		if err == io.EOF {
			return n, nil
		}

		return
	}

	return 0, io.EOF
}

func (f *feed) WriteAt(p []byte, off int64) (int, error) {
	return 0, errors.New("writing to feed files is currently unsupported")
}

func (f *feed) Close() error { return nil }
