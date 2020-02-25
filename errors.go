package main

import (
	"os"
	"path"
)

func init() {
	s := &fileHandler{
		fn:   getErrors,
		stat: getErrorsStat,
	}
	addFileHandler("/errors", s)
}

func getErrors(msg *message) (interface{}, error) {
	fp := path.Join(*inpath, msg.svc.name, "errors")
	return os.Open(fp)
}

func getErrorsStat(msg *message) (os.FileInfo, error) {
	fp := path.Join(*inpath, msg.svc.name, "errors")
	return os.Lstat(fp)
}
