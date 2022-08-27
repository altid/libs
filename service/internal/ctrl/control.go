package ctrl

import (
	"context"
	"fmt"
	"path"
	"sort"

	"github.com/altid/libs/markup"
	"github.com/altid/libs/service/command"
	"github.com/altid/libs/service/input"
	"github.com/altid/libs/service/internal/cmd"
	"github.com/altid/libs/store"
)

type Control struct {
	ctx     context.Context
	store   store.Filer
	tabs    store.File
	errors  store.File
	cmdlist []*cmd.Command
	tablist []*tab
	done    chan struct{}
	rundir  string
	logdir  string
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
	ctx    context.Context
	cancel context.CancelFunc
	name   string
}

func New(ctx context.Context, store store.Filer, r, l, d string, t []string) *Control {
	var tablist []*tab

	tf, _ := store.Open("/tabs")
	ew, _ := store.Open("/errors")

	for _, name := range t {
		tctx, cancel := context.WithCancel(ctx)
		tablist = append(tablist, &tab{
			name:   name,
			ctx:    tctx,
			cancel: cancel,
		})
	}

	return &Control{
		ctx:     ctx,
		done:    make(chan struct{}),
		errors:  ew,
		logdir:  l,
		tabs:    tf,
		store:   store,
		tablist: tablist,
	}
}

func (c *Control) Input(handler input.Handler, buffer string, payload []byte) error {
	ep := path.Join("/", c.rundir, buffer)
	l := markup.NewLexer(payload)

	return handler.Handle(ep, l)
}

func (c *Control) SetCommands(cmd ...*cmd.Command) error {
	c.cmdlist = append(c.cmdlist, cmd...)
	sort.Sort(command.CmdList(c.cmdlist))

	return nil
}

func (c *Control) BuildCommand(comm string) (*command.Command, error) {
	return command.ParseCmd(comm, c.cmdlist)
}

func (c *Control) Cleanup() {
	c.tabs.Close()
	c.errors.Close()
}

func (c *Control) CreateBuffer(name string) error {
	return c.pushTab(name)
}

func (c *Control) DeleteBuffer(name string) error {
	return c.popTab(name)
}

func (c *Control) HasBuffer(name string) bool {
	for _, item := range c.tablist {
		if item.name == name {
			return true
		}
	}
	return false
}

func (c *Control) Remove(buffer, filename string) error {
	doc := path.Join("/", buffer, filename)
	return c.store.Delete(doc)
}

func (c *Control) Notification(buff, from, msg string) error {
	nfile := path.Join("/", buff, "notification")
	f, err := c.store.Open(nfile)
	if err != nil {
		return err
	}

	defer f.Close()
	fmt.Fprintf(f, "%s\n%s\n", from, msg)

	return nil
}

func (c *Control) popTab(tabname string) error {
	for n := range c.tablist {
		if c.tablist[n].name == tabname {
			c.tablist[n].cancel()
			c.tablist = append(c.tablist[:n], c.tablist[n+1:]...)

			return writetabs(c)
		}
	}

	return fmt.Errorf("entry not found: %s", tabname)
}

func (c *Control) pushTab(tabname string) error {
	for n := range c.tablist {
		if c.tablist[n].name == tabname {
			return fmt.Errorf("entry already exists: %s", tabname)
		}
	}

	ctx, cancel := context.WithCancel(c.ctx)
	t := &tab{
		name:   tabname,
		ctx:    ctx,
		cancel: cancel,
	}

	c.tablist = append(c.tablist, t)

	return writetabs(c)
}

func writetabs(c *Control) error {
	var size int
	for _, tab := range c.tablist {
		n, _ := fmt.Fprintf(c.tabs, "%s\n", tab.name)
		size += n
	}

	return c.tabs.Truncate(int64(size))
}

func (c *Control) FileWriter(buffer, target string) (*WriteCloser, error) {
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

func (c *Control) Errorwriter() (*WriteCloser, error) {
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

func (c *Control) ImageWriter(buffer, resource string) (*WriteCloser, error) {
	ep := path.Join("/", buffer, "images")
	return c.FileWriter(ep, resource)
}
