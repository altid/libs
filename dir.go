package main

import (
	"io"
	"io/ioutil"
	"os"
	"path"
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
	off int64
}

func getDir(msg *message) (interface{}, error) {
	c := make(chan os.FileInfo, 10)
	done := make(chan struct{})
	fp := path.Join(*inpath, msg.service, msg.file)

	list, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, err
	}
	// We take the least resistance for error handling here
	// a missing entry may occur in the worst case
	// but a direct read of the file will correctly error
	// with all details we want
	cstat, err := getCtlStat(msg)
	if err == nil {
		list = append(list, cstat)
	}

	cfeed, err := getFeedStat(msg)
	if err == nil {
		list = append(list, cfeed)
	}

	ctabs, err := getTabsStat(msg)
	if err != nil {
		list = append(list, ctabs)
	}

	cinput, err := getInputStat(msg)
	if err == nil {
		list = append(list, cinput)
	}

	go func([]os.FileInfo) {
		defer close(c)

		for _, f := range list {
			select {
			case c <- f:
			case <-done:
				break
			}
		}
	}(list)

	return nil, &dir{
		c:    c,
		done: done,
		name: fp,
	}
}

func getDirStat(msg *message) (os.FileInfo, error) {
	return os.Stat(path.Join(*inpath, msg.service, msg.file))
}

func (d *dir) Error() string { return "" }
func (d *dir) Uid() string   { return defaultUID }
func (d *dir) Gid() string   { return defaultGID }

// Readdir for Directory interface
func (d *dir) Readdir(n int) ([]os.FileInfo, error) {
	var err error

	if n <= 0 {
		return d.readAllDir()
	}

	fi := make([]os.FileInfo, 0, 10)

	for i := 0; i < n; i++ {
		s, ok := <-d.c
		if !ok {
			err = io.EOF
			break
		}

		fi = append(fi, s)
	}


	d.done <- struct{}{}

	return fi, err
}

func (d *dir) readAllDir() ([]os.FileInfo, error) {
	fi := make([]os.FileInfo, 0, 10)

	for s := range d.c {
		fi = append(fi, s)
	}

	return fi, io.EOF
}

func (d *dir) Close() error {
	close(d.done)

	return nil
}
