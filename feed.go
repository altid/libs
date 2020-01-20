package main

import (
	"os"
	"path"
)

// feed files are special in that they're blocking
type feed struct{}

func init() {
	s := &fileHandler{
		fn:   getFeed,
		stat: getFeedStat,
	}
	addFileHandler("/feed", s)
}

//func (f *feed) Read() {}
//func (f *feed) Close() {}
func getFeed(msg *message) (interface{}, error) {
	fp := path.Join(*inpath, msg.service, msg.buff, "feed")
	return os.Open(fp)
}
func getFeedStat(msg *message) (os.FileInfo, error) {
	fp := path.Join(*inpath, msg.service, msg.buff, "feed")
	return os.Lstat(fp)
}
