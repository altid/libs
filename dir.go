package main

import (
	"io"
	"io/ioutil"
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
	name  string
	c     chan os.FileInfo
	done  chan struct{}
	count int64
	total int64
}

func getDir(msg *message) (interface{}, error) {
	c := make(chan os.FileInfo)
	done := make(chan struct{})
	fp := path.Join(*inpath, msg.svc.name, msg.buff)

	list, total, err := listDir(msg, fp)
	if err != nil {
		return nil, err
	}
	go func(list []os.FileInfo, c chan os.FileInfo, done chan struct{}) {
		defer close(c)

		for _, f := range list {
			select {
			case c <- f:
				continue
			case <-done:
				break
			}
		}
	}(list, c, done)

	d := &dir{
		c:     c,
		done:  done,
		name:  fp,
		total: total,
	}

	return d, nil
}

func getDirStat(msg *message) (os.FileInfo, error) {
	fp := path.Join(*inpath, msg.svc.name, msg.buff)

	_, count, err := listDir(msg, fp)
	if err != nil {
		return nil, err
	}

	d := &dir{
		count: count,
	}

	return d, nil
}

func listDir(msg *message, fp string) ([]os.FileInfo, int64, error) {
	var count int64

	list, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, 0, err
	}

	// We need to trim duplicate entries from here should they exist
	for _, entry := range list {
		switch entry.Name() {
		case "feed", "input", "tabs":
			if len(list) > 1 {
				list[count] = list[len(list)-1]
			}
			list = list[:len(list)-1]
		default:
			count++
		}
	}

	// We take the least resistance for error handling here
	// a missing entry may occur in the worst case
	// but a direct read of the file will correctly error
	// with all details we want
	cstat, err := getCtlStat(msg)
	if err == nil {
		list = append(list, cstat)
		count++
	}

	cfeed, err := getFeedStat(msg)
	if err == nil {
		list = append(list, cfeed)
		count++
	}

	ctabs, err := getTabsStat(msg)
	if err == nil {
		list = append(list, ctabs)
		count++
	}

	cinput, err := getInputStat(msg)
	if err == nil {
		list = append(list, cinput)
		count++
	}

	return list, count, nil
}

func (d *dir) Name() string       { return d.name }
func (d *dir) IsDir() bool        { return true }
func (d *dir) ModTime() time.Time { return time.Now().Truncate(time.Hour) }
func (d *dir) Mode() os.FileMode  { return os.ModeDir | 0755 }
func (d *dir) Sys() interface{}   { return d }
func (d *dir) Uid() string        { return defaultUID }
func (d *dir) Gid() string        { return defaultGID }
func (d *dir) Size() int64        { return 0 }

func (d *dir) Readdir(n int) ([]os.FileInfo, error) {
	var err error

	if d.count == d.total {
		return nil, io.EOF
	}

	if n <= 0 || int64(n) >= d.total {
		return d.readAllDir()
	}

	fi := make([]os.FileInfo, 0, n)

	for i := 0; i < n; i++ {
		s, ok := <-d.c
		if !ok {
			err = io.EOF
			break
		}
		d.count++
		fi = append(fi, s)
	}

	return fi, err
}

func (d *dir) readAllDir() ([]os.FileInfo, error) {
	fi := make([]os.FileInfo, 0, d.total)

	for s := range d.c {
		fi = append(fi, s)
		d.count++
	}

	return fi, nil
}

func (d *dir) Close() error {
	close(d.done)
	return nil
}
