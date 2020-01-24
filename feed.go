package main

import (
	"os"
	"path"
)

// feed files are special in that they're blocking
//type feed struct{}

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
	return os.Open(path.Join(*inpath, msg.service, msg.buff, "feed"))
}

func getFeedStat(msg *message) (os.FileInfo, error) {
	return os.Stat(path.Join(*inpath, msg.service, msg.buff, "feed"))
}
