package main

import (
	"os"
	"path"
)

func init() {
	s := &fileHandler{
		fn:   getNormal,
		stat: getNormalStat,
	}
	addFileHandler("/default", s)
}

func getNormal(msg *message) (interface{}, error) {
	fp := path.Join(*inpath, msg.service, msg.buff, msg.file)
	return os.Open(fp)
}

func getNormalStat(msg *message) (os.FileInfo, error) {
	fp := path.Join(*inpath, msg.service, msg.buff, msg.file)
	return os.Lstat(fp)
}
