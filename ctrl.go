package fslib

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
)

var valid *regexp.Regexp = regexp.MustCompile("[^ -~]+")

// Controller
// Open is called when a control message starting with 'open' or 'join' is written to the ctrl file
// Close is called when a control message starting with 'close or 'part' is written to the ctrl file
// Default is called when any other control message is written to the ctrl file.
// When Open is called, a file will be created with a path of `mountpoint/msg/document (or feed)`, containing initially a file named what you've set doctype to.. Calls to open are expected to populate that file, as well as create any supplementary files needed, such as title, sidebar, status, input, etc
// The main document or feed file is also symlinked into the given log directory, under service/msgs, so for example, an expensive parse would only have to be completed once for a given request, even across seperate runs; or a chat log could have history from previous sessions accessible.
// The message provided to all three functions is all of the message, less 'open', 'join', 'close', or 'part'.
type Controller interface {
	Open(c *Control, msg string) error
	Close(c *Control, msg string) error
	Default(c *Control, msg string) error
}

type Control struct {
	rundir  string
	logdir  string
	doctype string
	tabs    []string
	req     chan string
	done    chan struct{}
	ctrl    Controller
}

// CreateCtrlFile sets up a ready-to-listen ctrl file
// logdir is the directory to store copies of the contents of files created; specifically doctype. Logging any other type of data is left to implementation details, but is considered poor form for ubqt's design.
// mtpt is the directory to create the file system in
// service is the subdirectory inside mtpt for the runtime fs
// This will return an error if a ctrl file exists at the given directory, or if doctype is invalid.
func CreateCtrlFile(ctrl Controller, logdir, mtpt, service, doctype string) (*Control, error) {
	if doctype != "document" && doctype != "feed" {
		return nil, fmt.Errorf("Unknown doctype: %s", doctype)
	}
	rundir := path.Join(mtpt, service)
	_, err := os.Stat(path.Join(rundir, "ctrl"))
	if os.IsNotExist(err) {
		var tab []string
		req := make(chan string)
		done := make(chan struct{})
		c := &Control{
			rundir:  rundir,
			logdir:  logdir,
			doctype: doctype,
			tabs:    tab,
			req:     req,
			done:    done,
			ctrl:    ctrl,
		}
		return c, nil
	}
	return nil, fmt.Errorf("Control file already exist at %s", rundir)
}

// Event appends the given string to the events file of Control's working directory.
// Strings cannot contain newlines, tabs, spaces, or control characters.
// Returns "error - invalid string" or nil.
func (c *Control) Event(eventmsg string) error {
	return event(c, eventmsg)
}

// Cleanup removes created symlinks and removes the main dir
// On plan9, it unbinds any file named 	"document" or "feed", prior to removing the directory itself.
func (c *Control) Cleanup() {
	if runtime.GOOS == "plan9" {
		glob := path.Join(c.rundir, "*", c.doctype)
		files, err := filepath.Glob(glob)
		if err != nil {
			log.Print(err)
		}
		for _, f := range files {
			command := exec.Command("/bin/unmount", f)
			command.Run()
		}
	}
	os.RemoveAll(c.rundir)
}

// CreateBuffer creates a buffer of given name, as well as symlinking your file as follows:
// `os.Symlink(path.Join(logdir, name), path.Join(rundir, name, doctype))`
// This logged file will persist across reboots
// Calling CreateBuffer on a directory that already exists will return nil
func (c *Control) CreateBuffer(name, doctype string) error {
	d := path.Join(c.rundir, name, doctype)
	_, err := os.Stat(path.Join(c.rundir, name))
	if os.IsNotExist(err) {
		os.MkdirAll(path.Join(c.rundir, name), 0755)	
	}
	if err == nil {
		return nil
	}
	ioutil.WriteFile(d, []byte("Welcome!\n"), 0644)
	if c.logdir == "none" {
		return nil
	}
	logfile := path.Join(c.logdir, name)
	c.pushTab(name)
	return symlink(logfile, d)
}

// DeleteBuffer unlinks a document/buffer, and cleanly removes the directory
// Will return an error if it's unable to unlink on plan9, or if the remove fails.
func (c *Control) DeleteBuffer(name, doctype string) error {
	d := path.Join(c.rundir, name, doctype)
	if c.logdir != "none" {
		err := unlink(d)
		if err != nil {
			return err
		}
	}
	return os.RemoveAll(path.Join(c.rundir, name))
}

// Listen creates a file named "ctrl" inside RunDirectory, after making sure the directory exists
// Any text written to the ctrl file will be parsed, line by line.
// Messages handled internally are as follows: open (or join), close (or part), and quit, which causes Listen() to return.
// This will return an error if we're unable to create the ctrlfile itself, and will log any error relating to control messages.
func (c *Control) Listen() error {
	err := os.MkdirAll(c.rundir, 0755)
	if err != nil {
		return err
	}
	cfile := path.Join(c.rundir, "ctrl")
	go sigwatch(c)
	go dispatch(c)
	r, err := newReader(cfile)
	if err != nil {
		return err
	}
	event(c, cfile)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "quit" {
			close(c.done)
			break
		}
		c.req <- line
	}
	close(c.req)
	return nil
}

// Start is like listen, but occurs in a seperate go routine, returning flow to the calling process once the ctrl file is instantiated.
// This provides a context.Context that can be used for cancellations
func (c *Control) Start() (context.Context, error) {
	err := os.MkdirAll(c.rundir, 0755)
	if err != nil {
		return nil, err
	}
	cfile := path.Join(c.rundir, "ctrl")
	go sigwatch(c)
	go dispatch(c)
	event(c, cfile)
	r, err := newReader(cfile)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(c.req)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "quit" {
				cancel()
				close(c.done)
				break
			}
			c.req <- line
		}
	}()
	return ctx, nil
}

// Notification appends the content of msg to a buffers notification file
// Any errors encountered during file opening/creation will be returned
// The canonical form of notification can be found in cleanmark's Notification type,
// And the output of the Parse() method can be used directly here
// For example
//     ntfy, err := cleanmark.NewNotifier(buff, from, msg)
//     if err != nil {
//         log.Fatal(err)
//     }
//     fslib.Notification(ntfy.Parse())
func (c *Control) Notification(buff, from, msg string) error {
	nfile := path.Join(c.rundir, buff, "notification")
	if _, err := os.Stat(path.Dir(nfile)); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(nfile), 0755)
	}
	f, err := os.OpenFile(nfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "%s\n%s\n", from, msg)
	return nil
}

func (c *Control) pushTab(tabname string) error {
	err := validateString(tabname)
	if err != nil {
		return err
	}
	for n := range c.tabs {
		if c.tabs[n] == tabname {
			return fmt.Errorf("entry exists: %s", tabname)
		}
	}
	c.tabs = append(c.tabs, tabname)
	tabs(c)
	return nil
}

func (c *Control) popTab(tabname string) error {
	for n := range c.tabs {
		if c.tabs[n] == tabname {
			c.tabs = append(c.tabs[:n], c.tabs[n+1:]...)
			tabs(c)
			return nil
		}
	}
	return fmt.Errorf("Entry not found: %s", tabname)
}

func sigwatch(c *Control) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGKILL, syscall.SIGINT)
	for sig := range sigs {
		switch sig {
		case syscall.SIGKILL, syscall.SIGINT:
			c.Cleanup()
			//case syscall.SIGUSR
		}
	}
}

func tabs(c *Control) {
	// Create truncates and opens file in a single step, utilize this.
	file := path.Join(c.rundir, "tabs")
	f, err := os.Create(file)
	defer f.Close()
	if err != nil {
		log.Print(err)
		return
	}
	f.WriteString(strings.Join(c.tabs, "\n") + "\n")
	c.Event(file)
}

func dispatch(c *Control) {
	// TODO: wrap with waitgroups
	// If close is requested on a file which is currently being opened, cancel open request
	// If open is requested on file which already exists, no-op
	for {
		select {
		case line := <-c.req:
			token := strings.Fields(line)
			if len(token) < 1 {
				continue
			}
			switch token[0] {
			case "open":
				if len(token) < 2 {
					continue
				}
				err := c.ctrl.Open(c, token[1])
				if err != nil {
					log.Print(err)
					continue
				}
				err = c.pushTab(token[1])
				if err != nil {
					log.Print(err)
					continue
				}
			case "close":
				if len(token) < 2 {
					continue
				}
				err := c.ctrl.Close(c, token[1])
				if err != nil {
					log.Print(err)
					continue
				}
				err = c.popTab(token[1])
				if err != nil {
					log.Print(err)
					continue
				}
			default:
				err := c.ctrl.Default(c, line)
				if err != nil {
					log.Print(err)
					continue
				}
			}
		case <-c.done:
			return
		}
	}
}
