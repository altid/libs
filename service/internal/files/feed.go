package files

import (
	"io"
	"io/fs"
	"time"

	"github.com/altid/libs/store"
)

// On Feed() create a channel that will drain to Read() until close
// fill channel from Write as well as writing to underlying f.feed.Write
// QUESTION(halfwit) Do we need a really big queue of data? Readers should be on read without more than a few milliseconds delay
type FeedFile struct {
	feed store.File
}

// Hold our offset, read to EOF, return; but don't return EOF from here
func Feed(feed store.File, err error) (*FeedFile, error) {
	feed.Seek(0, io.SeekStart)
	f := &FeedFile{
		feed: feed,
	}

	return f, nil
}

// Attempt to do the right thing here with the reads
func (f *FeedFile) Read(b []byte) (n int, err error) {
	switch n, err = f.feed.Read(b); err {
	case io.EOF:
		time.Sleep(time.Millisecond * 400)
		return n, nil
	case nil:
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
