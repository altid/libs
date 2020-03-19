package files

import (
	"errors"
	"io"
	"os"
	"path"

	"github.com/altid/server/files"
	"github.com/altid/server/tail"
)

type FeedHandler struct {
	event chan *tail.Event
}

func (fh *FeedHandler) NewFeed(event chan *tail.Event) *FeedHandler {
	return &FeedHandler{event}
}

func (fh *FeedHandler) Normal(msg *files.Message) (interface{}, error) {
	done := make(chan struct{})
	f := &feed{
		event: fh.event,
		path:  path.Join(msg.Service, msg.Buffer, "feed"),
		buff:  path.Join(msg.Service, msg.Buffer),
		done:  done,
	}

	return f, nil

}

func (*FeedHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	return os.Lstat(path.Join(msg.Service, msg.Buffer, "feed"))
}

// feed files are special in that they're blocking
type feed struct {
	event   chan struct{}
	tailing bool
	path    string
	buff    string
	done    chan struct{}
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
