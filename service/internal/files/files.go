package files

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/altid/libs/service/controller"
	"github.com/altid/libs/store"
)

var l *log.Logger

type fileMsg int

const (
	fileErr fileMsg = iota
	fileBuffer
	fileDelete
	fileNotify
	fileRemove
)

type Files struct {
	store   store.Filer
	errors  store.File
	tablist map[string]interface{}
	debug   func(fileMsg, ...interface{})
}

type WriteCloser struct {
	store store.File
	path  string
}

func (w WriteCloser) Write(b []byte) (int, error) {
	return w.store.Write(b)
}

func (w WriteCloser) Close() error {
	return nil
}

func New(store store.Filer, debug bool) *Files {
	ew, _ := store.Open("/errors")
	f := &Files{
		errors:  ew,
		store:   store,
		tablist: make(map[string]interface{}),
		debug:   func(fileMsg, ...interface{}) {},
	}

	if debug {
		l = log.New(os.Stdout, "files ", 0)
		f.debug = fileLogger
	}

	return f
}

func (c *Files) Cleanup() {
	c.errors.Close()
}

func (c *Files) CreateBuffer(name string) error {
	// Make a store item
	switch e := c.store.Mkdir(name); e {
	case nil:
	case store.ErrDirExists:
	default:
		c.debug(fileErr, e)
		return e
	}

	c.debug(fileBuffer, name)
	c.tablist[name] = nil
	writetab(c.store, c.tablist)

	return nil
}

func (c *Files) DeleteBuffer(name string) error {
	if e := c.store.Delete(name); e != nil {
		c.debug(fileErr, e)
		return e
	}

	delete(c.tablist, name)
	c.debug(fileDelete, name)
	if e := writetab(c.store, c.tablist); e != nil {
		c.debug(fileErr, e)
	}

	return nil
}

func (c *Files) HasBuffer(name string) bool {
	if _, ok := c.tablist[name]; ok {
		return true
	}

	return false
}

func (c *Files) Remove(buffer, filename string) error {
	doc := path.Join("/", buffer, filename)
	c.debug(fileRemove, doc)
	return c.store.Delete(doc)
}

func (c *Files) Notification(buff, from, msg string) error {
	nfile := path.Join("/", buff, "notification")
	f, err := c.store.Open(nfile)
	if err != nil {
		c.debug(fileErr, err)
		return err
	}

	defer f.Close()
	c.debug(fileNotify, buff, from, msg)
	fmt.Fprintf(f, "%s\n%s\n", from, msg)

	return nil
}

func (c *Files) FeedWriter(buffer string) (controller.WriteCloser, error) {
	return c.FileWriter(buffer, "feed")
}

func (c *Files) MainWriter(buffer string) (controller.WriteCloser, error) {
	return c.FileWriter(buffer, "main")
}

func (c *Files) NavWriter(buffer string) (controller.WriteCloser, error) {
	return c.FileWriter(buffer, "navi")
}

func (c *Files) SideWriter(buffer string) (controller.WriteCloser, error) {
	return c.FileWriter(buffer, "aside")
}

func (c *Files) StatusWriter(buffer string) (controller.WriteCloser, error) {
	return c.FileWriter(buffer, "status")
}

func (c *Files) TitleWriter(buffer string) (controller.WriteCloser, error) {
	return c.FileWriter(buffer, "title")
}

func (c *Files) FileWriter(buffer, target string) (controller.WriteCloser, error) {
	ep := path.Join("/", buffer, target)
	mf, err := c.store.Open(ep)
	if err != nil {
		c.debug(fileErr, err)
		return nil, err
	}

	wc := &WriteCloser{
		store: mf,
		path:  ep,
	}

	return wc, nil
}

func (c *Files) ErrorWriter() (controller.WriteCloser, error) {
	ew, err := c.store.Open("/errors")
	if err != nil {
		c.debug(fileErr, err)
		return nil, err
	}

	wc := &WriteCloser{
		store: ew,
		path:  "/errors",
	}

	return wc, nil
}

func (c *Files) ImageWriter(buffer, resource string) (controller.WriteCloser, error) {
	ep := path.Join("/", buffer, "images")
	return c.FileWriter(ep, resource)
}

func writetab(store store.Opener, list map[string]interface{}) error {
	var size int
	tabs, err := store.Open("/tabs")
	if err != nil {
		return err
	}

	tabs.Seek(0, io.SeekStart)
	b := bufio.NewWriter(tabs)
	for tab := range list {
		n, err := b.WriteString(tab + "\n")
		if err != nil {

			return err
		}

		size += n
	}

	if e := tabs.Truncate(int64(size)); e != nil {
		return e
	}

	return tabs.Close()
}

func fileLogger(msg fileMsg, args ...interface{}) {
	switch msg {
	case fileErr:
		l.Printf("error: %s", args[0])
	case fileBuffer:
		l.Printf("create: buffer %s", args[0])
	case fileDelete:
		l.Printf("delete: buffer %s", args[0])
	case fileNotify:
		l.Printf("notification: buff=\"%s\" from=\"%s\" msg=\"%s\"", args[0], args[1], args[2])
	case fileRemove:
		l.Printf("remove: buff=\"%s\"", args[0])
	}
}
