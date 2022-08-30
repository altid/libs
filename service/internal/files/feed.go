package files

import (
	"io"
	"io/fs"

	"github.com/altid/libs/store"
)

type FeedFile struct {
	feed store.File
	data io.ReadCloser
}

func Feed(buffer string, stream store.Streamer, feed store.File) (*FeedFile, error) {
	data, err := stream.Stream(buffer)
	if err != nil {
		return nil, err
	}

	f := &FeedFile{
		feed: feed,
		data: data,
	}

	return f, nil
}

// Attempt to do the right thing here with the reads
func (f *FeedFile) Read(b []byte) (n int, err error) { 
	n, err = f.data.Read(b)
	if err == io.EOF {
		return n, nil
	}

	return
}

func (f *FeedFile) Write(p []byte) (n int, err error)            { return f.feed.Write(p) }
func (f *FeedFile) Truncate(cap int64) error                     { return f.feed.Truncate(cap) }
func (f *FeedFile) Seek(offset int64, whence int) (int64, error) { return f.feed.Seek(offset, whence) }
func (f *FeedFile) Name() string                                 { return "/feed" }
func (f *FeedFile) Close() error                                 { return f.feed.Close() }
func (f *FeedFile) Stat() (fs.FileInfo, error)                   { return f.feed.Stat() }
