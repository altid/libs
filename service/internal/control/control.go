package control

// We really gotta kill off all of the file-specific stuff before we do anything else

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"sort"

	"github.com/altid/libs/markup"
	"github.com/altid/libs/service/input"
	"github.com/altid/libs/service/internal/command"
	"github.com/altid/libs/store"
)

type Control struct {
	tabs    store.File
	errors  store.File
	cmdlist []*command.Command
	done    chan struct{}
	rundir  string
	logdir  string
	store	store.Filer
	tablist []*tab
	ctx     context.Context
}

type WriteCloser struct {
	store store.File
	path string
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
	// input  io.ReadCloser
}

func New(ctx context.Context, r, l, d string, t []string) *Control {
	var tablist []*tab

	data := store.NewRamStore()

	tf, _ := data.Open("tabs")
	ew, _ := data.Open("errors")

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
		rundir:  r,
		errors:  ew,
		logdir:  l,
		tabs:    tf,
		store:   data,
		tablist: tablist,
	}
}

func (c *Control) Input(handler input.Handler, buffer string, payload []byte) error {
	l := markup.NewLexer(payload)
	return handler.Handle(c.rundir, l)
}

func (c *Control) SetCommands(cmd ...*command.Command) error {
	c.cmdlist = append(c.cmdlist, cmd...)
	sort.Sort(command.CmdList(c.cmdlist))

	return nil
}

func (c *Control) BuildCommand(cmd string) (*command.Command, error) {
	return command.ParseCmd(cmd, c.cmdlist)
}

func (c *Control) Cleanup() {
	c.tabs.Close()
	c.errors.Close()
}

// These next two functions may end up needing more
// in the store side of things, but for now we just do the more simple thing
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
	doc := path.Join(c.rundir, buffer, filename)
	return c.store.Delete(doc)
}

// TODO: Use STORE for this
func (c *Control) Notification(buff, from, msg string) error {
	nfile := path.Join(c.rundir, buff, "notification")
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
	var sb bytes.Buffer

	for _, tab := range c.tablist {
		sb.WriteString(tab.name + "\n")
	}

	if _, e := c.tabs.Seek(0, io.SeekStart); e != nil {
		return e
	}

	if _, e := c.tabs.Write(sb.Bytes()); e != nil {
		return e
	}

	return c.tabs.Truncate(int64(sb.Len()))
}

func (c *Control) FileWriter(buffer, target string) (*WriteCloser, error) {
	ep := path.Join(buffer, target)
	mf, err := c.store.Open(buffer)
	if err != nil {
		return nil, err
	}

	wc := &WriteCloser{
		store: mf,
		path: ep,
	}
	
	return wc, nil
}

func (c *Control) Errorwriter() (*WriteCloser, error) {
	ew, err := c.store.Open("errors")
	if err != nil {
		return nil, err
	}

	wc := &WriteCloser{
		store: ew,
		path: "errors",
	}

	return wc, nil
}

func (c *Control) ImageWriter(buffer, resource string) (*WriteCloser, error) {
	ep := path.Join(buffer, "images")
	return c.FileWriter(ep, resource)
}

