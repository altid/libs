package main

import (
	"os"
	"path"
)

func init() {
	s := &fileHandler{
		fn:   getInput,
		stat: getInputStat,
	}
	addFileHandler("/input", s)
}

func getInput(msg *message) (interface{}, error) {
	fp := path.Join(*inpath, msg.svc.name, msg.buff, "input")
	return os.OpenFile(fp, os.O_RDWR|os.O_APPEND, 0644)
}

func getInputStat(msg *message) (os.FileInfo, error) {
	return os.Stat(path.Join(*inpath, msg.svc.name, msg.buff, "input"))
}
