package files

import (
	"fmt"
	"path"

	"github.com/altid/libs/markup"
	"github.com/altid/libs/service/controller"
	"github.com/altid/libs/service/input"
	"github.com/altid/libs/store"
)

type Files struct {
	store   store.Filer
	tabs    store.File
	errors  store.File
	tablist []*tab
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

type tab struct {
	name string
}

func New(store store.Filer) *Files {
	var tablist []*tab

	tf, _ := store.Open("/tabs")
	ew, _ := store.Open("/errors")

	return &Files{
		errors:  ew,
		tabs:    tf,
		store:   store,
		tablist: tablist,
	}
}

func (c *Files) Input(handler input.Handler, buffer string, payload []byte) error {
	ep := path.Join("/", buffer)
	l := markup.NewLexer(payload)

	return handler.Handle(ep, l)
}

func (c *Files) Cleanup() {
	c.tabs.Close()
	c.errors.Close()
}

func (c *Files) CreateBuffer(name string) error {
	return c.pushTab(name)
}

func (c *Files) DeleteBuffer(name string) error {
	return c.popTab(name)
}

func (c *Files) HasBuffer(name string) bool {
	for _, item := range c.tablist {
		if item.name == name {
			return true
		}
	}
	return false
}

func (c *Files) Remove(buffer, filename string) error {
	doc := path.Join("/", buffer, filename)
	return c.store.Delete(doc)
}

func (c *Files) Notification(buff, from, msg string) error {
	nfile := path.Join("/", buff, "notification")
	f, err := c.store.Open(nfile)
	if err != nil {
		return err
	}

	defer f.Close()
	fmt.Fprintf(f, "%s\n%s\n", from, msg)

	return nil
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

func (c *Files) pushTab(tabname string) error {
	for n := range c.tablist {
		if c.tablist[n].name == tabname {
			return fmt.Errorf("entry already exists: %s", tabname)
		}
	}

	t := &tab{
		name: tabname,
	}

	c.tablist = append(c.tablist, t)

	return writetabs(c)
}

func (c *Files) popTab(tabname string) error {
	for n := range c.tablist {
		if c.tablist[n].name == tabname {
			c.tablist = append(c.tablist[:n], c.tablist[n+1:]...)

			return writetabs(c)
		}
	}

	return fmt.Errorf("entry not found: %s", tabname)
}

func writetabs(c *Files) error {
	var size int
	for _, tab := range c.tablist {
		n, _ := fmt.Fprintf(c.tabs, "%s\n", tab.name)
		size += n
	}

	return c.tabs.Truncate(int64(size))
}
