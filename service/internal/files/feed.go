package files

import (
	"io"
	"io/fs"
	"log"
	"time"

	"github.com/altid/libs/store"
)

type FeedFile struct {
	feed store.File
}

func Feed(feed store.File, err error) (*FeedFile, error) {
	feed.Seek(0, io.SeekStart)
	f := &FeedFile{
		feed: feed,
	}
	return f, err
}

// Attempt to do the right thing here with the reads
func (f *FeedFile) Read(b []byte) (n int, err error) {
	n, err = f.feed.Read(b)
	switch err {
	case store.ErrFileClosed:
		return 0, io.EOF
	case io.EOF:
		time.Sleep(time.Millisecond * 500)
		fallthrough
	case nil:
		log.Printf("bits read %d", n)
		return n, nil
	default:
		return 0, err
	}
}

func (f *FeedFile) Write(p []byte) (n int, err error)            { return f.feed.Write(p) }
func (f *FeedFile) Truncate(cap int64) error                     { return f.feed.Truncate(cap) }
func (f *FeedFile) Seek(offset int64, whence int) (int64, error) { return f.feed.Seek(offset, whence) }
func (f *FeedFile) Name() string                                 { return f.feed.Name() }
func (f *FeedFile) Stat() (fs.FileInfo, error)                   { return f.feed.Stat() }
func (f *FeedFile) Close() error                                 { return f.feed.Close() }
