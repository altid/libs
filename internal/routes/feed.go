package routes

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/altid/server/internal/message"
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
		close(fh.feeds[uuid])

		fh.Lock()
		defer fh.Unlock()
		delete(fh.feeds, uuid)
	}
}

// Normal returns a readwritecloser tied to "feed", which tails it
// until a new file is read by the client
func (fh *FeedHandler) Normal(msg *message.Message) (interface{}, error) {
	f := &feed{
		fh:   fh,
		uuid: msg.UUID,
		path: path.Join(msg.Service, msg.Buffer, "feed"),
		buff: path.Join(msg.Service, msg.Buffer),
	}

	return f, nil

}

// Stat returns a normal stat to an underlying feed file
func (*FeedHandler) Stat(msg *message.Message) (os.FileInfo, error) {
	return os.Lstat(path.Join(msg.Service, msg.Buffer, "feed"))
}

// feed files are special in that they're blocking
type feed struct {
	fh      *FeedHandler
	uuid    uint32
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

		fmt.Println("Tailin'")
		if err == io.EOF {
			f.tailing = true
			f.event = make(chan struct{})

			f.fh.Lock()
			defer f.fh.Unlock()
			f.fh.feeds[f.uuid] = f.event
		}

		return n, nil
	}

	// If the loop is live
	for range f.event {
		fmt.Println("Loopin'")
		n, err = fp.ReadAt(p, off)
		switch err {
		case io.EOF:
			return n, nil
		case nil:
			return
		default:
			return 0, err
		}
	}

	fmt.Println("End of filin'")
	return 0, io.EOF
}

func (f *feed) WriteAt(p []byte, off int64) (int, error) {
	return 0, errors.New("writing to feed files is currently unsupported")
}

func (f *feed) Close() error { return nil }
