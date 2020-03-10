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

type input struct {
	path string
}

// Simple wrapper around an open call
func (i *input) ReadAt(b []byte, off int64) (n int, err error) {
	fp, err := os.OpenFile(i.path, os.O_RDONLY, 0600)
	if err != nil {
		return
	}

	defer fp.Close()
	return fp.ReadAt(b, off)
}

// Open in correct modes
func (i *input) WriteAt(p []byte, off int64) (n int, err error) {
	fp, err := os.OpenFile(i.path, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}

	defer fp.Close()
	return fp.Write(p)
}

func (i *input) Close() error { return nil }
func (i *input) Uid() string  { return defaultUID }
func (i *input) Gid() string  { return defaultGID }

func getInput(msg *message) (interface{}, error) {
	fp := path.Join(*inpath, msg.svc.name, msg.buff, "input")
	i := &input{
		path: fp,
	}

	return i, nil
}

func getInputStat(msg *message) (os.FileInfo, error) {
	return os.Stat(path.Join(*inpath, msg.svc.name, msg.buff, "input"))
}
