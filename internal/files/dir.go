package files

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/altid/server/files"
)

type dirHandler struct{}

type dir struct {
	name  string
	c     chan os.FileInfo
	done  chan struct{}
	count int64
	total int64
}

func (*dirHandler) Normal(msg *files.Message) (interface{}, error) {
	c := make(chan os.FileInfo)
	done := make(chan struct{})
	fp := path.Join(msg.Service, msg.Buffer)

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

func (*dirHandler) Stat(msg *files.Message) (os.FileInfo, error) {
	fp := path.Join(msg.Service, msg.Buffer)

	_, count, err := listDir(msg, fp)
	if err != nil {
		return nil, err
	}

	d := &dir{
		count: count,
	}

	return d, nil
}

func listDir(msg *files.Message, fp string) ([]os.FileInfo, int64, error) {
	var count int64

	list, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, 0, err
	}

	// We need to trim duplicate entries from here should they exist
	for _, entry := range list {
		switch entry.Name() {
		case "feed", "input", "tabs", "ctl":
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
	c := &ctlHandler{}
	if cstat, e := c.Stat(msg); e == nil {
		list = append(list, cstat)
		count++
	}

	e := &errHandler{}
	if estat, e := e.Stat(msg); e == nil {
		list = append(list, estat)
		count++
	}

	f := &feedHandler{}
	if fstat, e := f.Stat(msg); e == nil {
		list = append(list, fstat)
		count++
	}

	t := &tabsHandler{}
	if tstat, e := t.Stat(msg); e == nil {
		list = append(list, tstat)
		count++
	}

	i := &inputHandler{}
	if istat, e := i.Stat(msg); e == nil {
		list = append(list, istat)
		count++
	}

	return list, count, nil
}

func (d *dir) Name() string       { return d.name }
func (d *dir) IsDir() bool        { return true }
func (d *dir) ModTime() time.Time { return time.Now().Truncate(time.Hour) }
func (d *dir) Mode() os.FileMode  { return os.ModeDir | 0755 }
func (d *dir) Sys() interface{}   { return d }
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
