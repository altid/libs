package files

import (
	"io"
	"io/fs"
	"time"

	"github.com/altid/libs/store"
)

type FeedFile struct {
	feed store.File
	offset int64
}

func Feed(feed store.File, err error) (*FeedFile, error) {
	feed.Seek(0, io.SeekStart)
	f := &FeedFile{
		feed: feed,
		offset: 0,
	}
	return f, err
}

// Attempt to do the right thing here with the reads
func (f *FeedFile) Read(b []byte) (n int, err error) {
	LOOP:
		n, err = f.feed.Read(b)
		f.offset += int64(n)
		switch err {
		case io.EOF:
			time.Sleep(time.Second)
			f.feed.Seek(f.offset, io.SeekStart)
			if n > 0 {
				return len(b), nil
			}
			goto LOOP
		case nil:
			return len(b), nil
		default:
			// Just exit the loop cleanly, as this can be a number of different errors depending on the backing store
			return n, io.EOF
		}
}

func (f *FeedFile) Write(p []byte) (n int, err error)            { return f.feed.Write(p) }
func (f *FeedFile) Truncate(cap int64) error                     { return f.feed.Truncate(cap) }
func (f *FeedFile) Seek(offset int64, whence int) (int64, error) { return f.feed.Seek(offset, whence) }
func (f *FeedFile) Name() string                                 { return f.feed.Name() }
func (f *FeedFile) Stat() (fs.FileInfo, error)                   { return f.feed.Stat() }
func (f *FeedFile) Close() error                                 { return f.feed.Close() }
