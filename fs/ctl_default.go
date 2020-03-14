package fs

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type control struct {
	cmdlist []*Command
	rundir  string
	logdir  string
	doctype string
	tabs    []string
	req     chan string
	done    chan struct{}
}

type controlrunner struct {
	ctx     context.Context
	scanner *bufio.Scanner
	done    chan struct{}
	req     chan string
	cancel  context.CancelFunc
}

func (c *control) event(eventmsg string) error {
	file := path.Join(c.rundir, "event")
	if _, err := os.Stat(path.Dir(file)); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(file), 0755)
	}

	if e := validateString(file); e != nil {
		sname := path.Base(c.rundir)
		return fmt.Errorf("%s: invalid event %s", sname, eventmsg)
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	defer f.Close()
	f.WriteString(eventmsg + "\n")

	return nil
}

func (c *control) setCommand(cmd ...*Command) error {
	for _, comm := range cmd {
		c.cmdlist = append(c.cmdlist, comm)
	}

	sort.Sort(cmdList(c.cmdlist))

	return nil
}

func (c *control) buildCommand(cmd string) (*Command, error) {
	return buildCommand(cmd, c.cmdlist)
}

func (c *control) cleanup() {
	if runtime.GOOS == "plan9" {
		glob := path.Join(c.rundir, "*", c.doctype)

		files, err := filepath.Glob(glob)
		if err != nil {
			log.Print(err)
		}

		for _, f := range files {
			command := exec.Command("/bin/unmount", f)
			log.Print(command.Run())
		}
	}

	os.RemoveAll(c.rundir)
}

func (c *control) createBuffer(name, doctype string) error {
	if name == "" {
		return fmt.Errorf("no buffer name given")
	}

	fp := path.Join(c.rundir, name)
	d := path.Join(fp, doctype)

	if _, e := os.Stat(fp); e != nil && !os.IsNotExist(e) {
		return e
	}

	if e := os.MkdirAll(fp, 0755); e != nil {
		return e
	}

	if e := ioutil.WriteFile(d, []byte("Welcome!\n"), 0644); e != nil {
		return e
	}

	if e := c.pushTab(name); e != nil {
		return e
	}

	// If there is no log, we're done otherwise create symlink
	if c.logdir == "none" {
		return nil
	}

	logfile := path.Join(c.logdir, name)

	return symlink(logfile, d)
}

func (c *control) deleteBuffer(name, doctype string) error {
	if c.logdir != "none" {
		d := path.Join(c.rundir, name, doctype)
		if e := unlink(d); e != nil {
			return e
		}
	}

	defer os.RemoveAll(path.Join(c.rundir, name))

	return c.popTab(name)
}

func (c *control) hasBuffer(name, doctype string) bool {
	d := path.Join(c.rundir, name, doctype)

	if _, e := os.Stat(d); os.IsNotExist(e) {
		return false
	}

	return true
}

func (c *control) listen() error {
	ctx, cancel := context.WithCancel(context.Background())

	cr, err := newControlRunner(ctx, cancel, c)
	if err != nil {
		return err
	}

	cr.listen()

	return nil
}

func (c *control) start() (context.Context, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cr, err := newControlRunner(ctx, cancel, c)
	if err != nil {
		return nil, err
	}

	go cr.listen()

	return ctx, nil
}

func (c *control) remove(buffer, filename string) error {
	doc := path.Join(c.rundir, buffer, filename)
	// Don't try to delete that which isn't there
	if _, e := os.Stat(doc); os.IsNotExist(e) {
		return nil
	}

	c.event(doc)
	return os.Remove(doc)
}

func (c *control) notification(buff, from, msg string) error {
	nfile := path.Join(c.rundir, buff, "notification")
	if _, e := os.Stat(path.Dir(nfile)); os.IsNotExist(e) {
		os.MkdirAll(path.Dir(nfile), 0755)
	}

	f, err := os.OpenFile(nfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	c.event(nfile)
	fmt.Fprintf(f, "%s\n%s\n", from, msg)

	return nil
}

func (c *control) popTab(tabname string) error {
	for n := range c.tabs {
		if c.tabs[n] == tabname {
			c.tabs = append(c.tabs[:n], c.tabs[n+1:]...)
			return tabs(c)
		}
	}
	return fmt.Errorf("entry not found: %s", tabname)
}

func (c *control) pushTab(tabname string) error {
	for n := range c.tabs {
		if c.tabs[n] == tabname {
			return fmt.Errorf("entry already exists: %s", tabname)
		}
	}
	c.tabs = append(c.tabs, tabname)

	return tabs(c)
}

func tabs(c *control) error {
	// Create truncates and opens file in a single step, utilize this.
	file := path.Join(c.rundir, "tabs")

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	f.WriteString(strings.Join(c.tabs, "\n") + "\n")
	c.event(file)

	return nil
}

func (c *control) errorwriter() (*WriteCloser, error) {
	ep := path.Join(c.rundir, "errors")

	fp, err := os.OpenFile(ep, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {

		return nil, err
	}

	w := &WriteCloser{
		c:      c,
		fp:     fp,
		buffer: "errors",
	}

	return w, nil
}

func (c *control) fileWriter(buffer, doctype string) (*WriteCloser, error) {
	doc := path.Join(c.rundir, buffer, doctype)
	if doctype == "feed" {
		fp, err := os.OpenFile(doc, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {

			return nil, err
		}
		w := &WriteCloser{
			fp:     fp,
			c:      c,
			buffer: doc,
		}

		return w, nil
	}

	// Abuse truncation semantics of Create so we clear any old data
	fp, err := os.Create(doc)
	if err != nil {
		return nil, err
	}

	w := &WriteCloser{
		fp:     fp,
		c:      c,
		buffer: doc,
	}

	return w, nil
}

func (c *control) imageWriter(buffer, resource string) (*WriteCloser, error) {
	os.MkdirAll(path.Dir(path.Join(c.rundir, buffer, "images", resource)), 0755)
	return c.fileWriter(buffer, path.Join("images", resource))
}

func newControlRunner(ctx context.Context, cancel context.CancelFunc, c *control) (*controlrunner, error) {
	if e := os.MkdirAll(c.rundir, 0755); e != nil {
		return nil, e
	}

	cfile := path.Join(c.rundir, "ctl")
	c.event(cfile)

	ctl, err := os.Create(cfile)
	if err != nil {
		return nil, err
	}

	if e := printCtlFile(c.cmdlist, ctl); e != nil {
		return nil, e
	}

	ctl.Close()

	r, err := newReader(cfile)
	if err != nil {
		return nil, err
	}

	cr := &controlrunner{
		scanner: bufio.NewScanner(r),
		req:     c.req,
		ctx:     ctx,
		cancel:  cancel,
		done:    c.done,
	}

	return cr, nil
}

func (c *controlrunner) listen() {
	defer close(c.req)

	for c.scanner.Scan() {
		select {
		case <-c.ctx.Done():
			break
		default:
			line := c.scanner.Text()
			if line == "quit" {
				c.cancel()
				close(c.done)
				break
			}

			c.req <- line
		}
	}
}
