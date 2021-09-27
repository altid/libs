package control 

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/altid/libs/markdown"
	"github.com/altid/libs/service/input"
	"github.com/altid/libs/service/internal/command"
)

type Control struct {
	tabs    *memfs.File
	errors  *memfs.File
	cmdlist []*command.Command
	done    chan struct{}
	rundir  string
	logdir  string
	store	*memfs.FS
	tablist []*tab
	ctx     context.Context
}

type WriteCloser struct {
	store *memfs.FS
	file *memfs.FS
	path string
}

func (w *WriteCloser) Write(b []byte) (int, error) {
	return len(b), w.store.Write(path, b)
}

func (w *WriteCloser) Close() error {
	return nil
}

type tab struct {
	ctx    context.Context
	cancel context.CancelFunc
	name   string
	input  io.ReadCloser
}

func New(ctx context.Context, r, l, d string, t []string) *Control {
	var tablist []*tab

	data := memfs.New()

	// We have to write data to our files before start
	if e := data.WriteFile("/errors", []byte(""), 0755); e != nil {
		log.Fatal(e)
	}

	if e := data.WriteFile("/tabs", []byte(""), 0755); e != nil {
		log.Fatal(e)
	}

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

func (c *Control) Input(handler input.Handler, buffer string, payload byte[]) error {
	l := markup.Lexer(payload)
	return handler.Handle(c, l)
}

func (c *Control) SetCommands(cmd ...*command.Command) error {
	for _, comm := range cmd {
		c.cmdlist = append(c.cmdlist, comm)
	}

	sort.Sort(command.CmdList(c.cmdlist))

	return nil
}

func (c *Control) BuildCommand(cmd string) (*command.Command, error) {
	return command.ParseCmd(cmd, c.cmdlist)
}

func (c *Control) Cleanup() {
	if runtime.GOOS == "plan9" {
		glob := path.Join(c.rundir, "*")

		files, err := filepath.Glob(glob)
		if err != nil {
			log.Print(err)
		}

		for _, f := range files {
			command := exec.Command("/bin/unmount", f)
			log.Print(command.Run())
		}
	}

	c.tabs.Close()
	c.errors.Close()
	os.RemoveAll(c.rundir)
}

func (c *Control) CreateBuffer(name string) error {
	c.pushTab(name)
	return c.store.MkdirAll(name,  0777)
}
 
// TODO: Remove from store
func (c *Control) DeleteBuffer(name string) error {
	c.popTab(name)
	c.store.Remove(name)
}

func (c *Control) HasBuffer(name string) bool {
	return c.store.HasBuffer(name)
}

func (c *Control) Remove(buffer, filename string) error {
	doc := path.Join(c.rundir, buffer, filename)
	return c.store.Remove(doc)
}

func (c *Control) Notification(buff, from, msg string) error {
	nfile := path.Join(c.rundir, buff, "notification")
	if _, e := os.Stat(path.Dir(nfile)); os.IsNotExist(e) {
		os.MkdirAll(path.Dir(nfile), 0755)
	}

	f, err := os.OpenFile(nfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
	var sb strings.Builder

	for _, tab := range c.tablist {
		sb.WriteString(tab.name + "\n")
	}

	if _, e := c.tabs.Seek(0, io.SeekStart); e != nil {
		return e
	}

	if _, e := c.tabs.WriteString(sb.String()); e != nil {
		return e
	}

	return c.tabs.Truncate(int64(sb.Len()))
}

func (c *Control) Errorwriter() (*WriteCloser, error) {
	ep := path.Join(c.rundir, "errors")
	w := c.store.New(c.Event, ep, "errors")

	return w, nil
}

func (c *Control) FileWriter(buffer(*WriteCloser, error) {
	w := c.store.New(c.Event, buffer)
	return w, nil
}

func (c *Control) ImageWriter(buffer, resource string) (*WriteCloser, error) {
	os.MkdirAll(path.Dir(path.Join(c.rundir, buffer, "images", resource)), 0755)
	return c.FileWriter(buffer, path.Join("images", resource))
}
