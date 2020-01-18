package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

func init() {
	s := &fileHandler{
		fn:   getDir,
		stat: getDirStat,
	}
	addFileHandler("/", s)
}

type dir struct {
	name string
	c    chan os.FileInfo
	done chan struct{}
}

func getDir(msg *message) (interface{}, error) {
	c := make(chan os.FileInfo, 10)
	done := make(chan struct{})
	fp := path.Join(*inpath, msg.service)
	list, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, nil
	}
	cstat, err := getCtlStat(msg)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	list = append(list, cstat)
	cinput, err := getInputStat(msg)
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	list = append(list, cinput)
	//ctabs, err := getTabsStat(msg)
	go func([]os.FileInfo) {
		for _, f := range list {
			select {
			case c <- f:
			case <-done:
				break
			}
		}
		close(c)
	}(list)
	return nil, &dir{
		c:    c,
		done: done,
		name: fp,
	}
}

func getDirStat(msg *message) (os.FileInfo, error) {
	fp := path.Join(*inpath, msg.service)
	return os.Stat(fp)
}

func (d *dir) Error() string      { return "" }
func (d *dir) IsDir() bool        { return true }
func (d *dir) ModTime() time.Time { return time.Now() }
func (d *dir) Mode() os.FileMode  { return os.ModeDir }
func (d *dir) Name() string       { return d.name }
func (d *dir) Size() int64        { return 0 }
func (d *dir) Sys() interface{}   { return nil }

// Listen for os.FileInfo members to come in from mkdir
func (d *dir) Readdir(n int) ([]os.FileInfo, error) {
	var err error
	fi := make([]os.FileInfo, 0, 10)
	for i := 0; i < n; i++ {
		s, ok := <-d.c
		if !ok {
			err = io.EOF
			break
		}
		fi = append(fi, s)
	}
	return fi, err
}

func (d *dir) Close() error {
	close(d.done)
	return nil
}
