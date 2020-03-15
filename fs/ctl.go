package fs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
)

// Controller is our main type for controlling a session
// Open is called when a control message starting with 'open' or 'join' is written to the ctl file
// Close is called when a control message starting with 'close or 'part' is written to the ctl file
// Link is called when a control message starting with 'link' is written to the ctl file
// Default is called when any other control message is written to the ctl file.
// If a client attempts to write an invalid control message, it will return a generic error
// When Open is called, a file will be created with a path of `mountpoint/msg/document (or feed)`, containing initially a file named what you've set doctype to.. Calls to open are expected to populate that file, as well as create any supplementary files needed, such as title, aside, status, input, etc
// When Link is called, the content of the current buffer is expected to change, and the name of the current tab will be removed, replaced with msg
// The main document or feed file is also symlinked into the given log directory, under service/msgs, so for example, an expensive parse would only have to be completed once for a given request, even across separate runs; or a chat log could have history from previous sessions accessible.
// The message provided to all three functions is all of the message, less 'open', 'join', 'close', or 'part'.
type Controller interface {
	Open(c *Control, msg string) error
	Close(c *Control, msg string) error
	Link(c *Control, from, msg string) error
	Default(c *Control, cmd *Command) error
}

// SigHandler - Optional interface to provide fine grained control for catching signals.
// It is expected that you will call c.Cleanup() within your SigHandle function
// If none is supplied, c.Cleanup() will be called on SIGINT and SIGKILL
type SigHandler interface {
	SigHandle(c *Control)
}

type runner interface {
	cleanup()
	setCommand(cmd ...*Command) error
	buildCommand(string) (*Command, error)
	event(string) error
	createBuffer(string, string) error
	deleteBuffer(string, string) error
	hasBuffer(string, string) bool
	listen() error
	remove(string, string) error
	start() (context.Context, error)
	notification(string, string, string) error
}

type writercloser interface {
	errorwriter() (*WriteCloser, error)
	fileWriter(string, string) (*WriteCloser, error)
	imageWriter(string, string) (*WriteCloser, error)
}

// Control type can be used to manage a running ctl file session
type Control struct {
	req   chan string
	done  chan struct{}
	ctl   Controller
	run   runner
	write writercloser
	watch watcher
	debug func(ctlMsg, ...interface{})
	sync.Mutex
}

type ctlMsg int

const (
	ctlError ctlMsg = iota
	ctlEvent
	ctlCleanup
	ctlCreate
	ctlDelete
	ctlRemove
	ctlStart
	ctlNotify
	ctlDefault
)

//TODO(halfiwt) i18n
var defaultCommands = []*Command{
	{
		Name:        "open",
		Args:        []string{"<buffer>"},
		Heading:     DefaultGroup,
		Description: "Open and change buffers to a given service",
	},
	{
		Name:        "close",
		Args:        []string{"<buffer>"},
		Heading:     DefaultGroup,
		Description: "Close a buffer and return to the last opened previously",
	},
	{
		Name:        "buffer",
		Args:        []string{"<buffer>"},
		Heading:     DefaultGroup,
		Description: "Change to the named buffer",
	},
	{
		Name:        "link",
		Args:        []string{"<to>", "<from>"},
		Heading:     DefaultGroup,
		Description: "Overwrite the current <to> buffer with <from>, switching to from after. This destroys <to>",
	},
	{
		Name:        "quit",
		Args:        []string{},
		Heading:     DefaultGroup,
		Description: "Exits the service",
	},
}

// MockCtlFile returns a type that can be used for testing services
// it will track in-memory and behave like a file-backed control
// It will wait for messages on reqs which act as ctl messages
// By default it writes to Stdout + Stderr with each WriteCloser
// If debug is true, all logs will be written to stdout
func MockCtlFile(ctl Controller, reqs chan string, debug bool) (*Control, error) {

	done := make(chan struct{})
	cmds := make(chan string)
	errs := make(chan error)
	t := &mockctl{
		err:  errs,
		reqs: reqs,
		cmds: cmds,
		done: done,
	}

	c := &Control{
		ctl:   ctl,
		req:   cmds,
		done:  done,
		run:   t,
		write: t,
		watch: watcher{},
		debug: func(ctlMsg, ...interface{}) {},
	}

	if debug {
		c.debug = ctlLogger
	}

	c.SetCommands(defaultCommands...)

	return c, nil
}

// CreateCtlFile sets up a ready-to-listen ctl file
// logdir is the directory to store copies of the contents of files created; specifically doctype. Logging any other type of data is left to implementation details, but is considered poor form for Altid's design.
// mtpt is the directory to create the file system in
// service is the subdirectory inside mtpt for the runtime fs
// This will return an error if a ctl file exists at the given directory, or if doctype is invalid.
func CreateCtlFile(ctl Controller, logdir, mtpt, service, doctype string, debug bool) (*Control, error) {
	if doctype != "document" && doctype != "feed" {
		return nil, fmt.Errorf("unknown doctype: %s", doctype)
	}

	rundir := path.Join(mtpt, service)

	if _, e := os.Stat(path.Join(rundir, "ctl")); os.IsNotExist(e) {
		var tab []string
		req := make(chan string)
		done := make(chan struct{})
		rtc := &control{
			rundir:  rundir,
			logdir:  logdir,
			doctype: doctype,
			tabs:    tab,
			req:     req,
			done:    done,
		}

		// TODO(halfwit) Go back and re-add signal watching

		c := &Control{
			req:   req,
			done:  done,
			run:   rtc,
			ctl:   ctl,
			write: rtc,
			watch: watcher{},
			debug: func(ctlMsg, ...interface{}) {},
		}

		if debug {
			c.debug = ctlLogger
		}

		c.SetCommands(defaultCommands...)

		return c, nil
	}

	return nil, fmt.Errorf("Control file already exist at %s", rundir)
}

// Event appends the given string to the events file of Control's working directory.
// Strings cannot contain newlines, tabs, spaces, or control characters.
// Returns "$service: invalid event $eventmsg" or nil.
func (c *Control) Event(eventmsg string) error {
	c.debug(ctlEvent, eventmsg)
	return c.run.event(eventmsg)
}

// Cleanup removes created symlinks and removes the main dir
// On plan9, it unbinds any file named 	"document" or "feed", prior to removing the directory itself.
func (c *Control) Cleanup() {
	c.debug(ctlCleanup)
	c.run.cleanup()

}

// CreateBuffer creates a buffer of given name, as well as symlinking your file as follows:
// `os.Symlink(path.Join(logdir, name), path.Join(rundir, name, doctype))`
// This logged file will persist across reboots
// Calling CreateBuffer on a directory that already exists will return nil
func (c *Control) CreateBuffer(name, doctype string) error {
	c.debug(ctlCreate, name, doctype)
	return c.run.createBuffer(name, doctype)
}

// DeleteBuffer unlinks a document/buffer, and cleanly removes the directory
// Will return an error if it's unable to unlink on plan9, or if the remove fails.
func (c *Control) DeleteBuffer(name, doctype string) error {
	c.debug(ctlDelete, name)
	return c.run.deleteBuffer(name, doctype)
}

// HasBuffer returns whether or not a buffer is present in the current control session
func (c *Control) HasBuffer(name, doctype string) bool {
	return c.run.hasBuffer(name, doctype)
}

// Remove removes a buffer from the runtime dir. If the buffer doesn't exist, this is a no-op
func (c *Control) Remove(buffer, filename string) error {
	c.debug(ctlRemove, buffer, filename)
	return c.run.remove(buffer, filename)
}

// Listen creates a file named "ctl" inside RunDirectory, after making sure the directory exists
// Any text written to the ctl file will be parsed, line by line.
// Messages handled internally are as follows: open (or join), close (or part), and quit, which causes Listen() to return.
// This will return an error if we're unable to create the ctlfile itself, and will log any error relating to control messages.
func (c *Control) Listen() error {
	go sigwatch(c)
	go dispatch(c)

	c.debug(ctlStart, "listen")
	return c.run.listen()
}

// Start is like listen, but occurs in a separate go routine, returning flow to the calling process once the ctl file is instantiated.
// This provides a context.Context that can be used for cancellations
func (c *Control) Start() (context.Context, error) {
	go sigwatch(c)
	go dispatch(c)

	c.debug(ctlStart, "start")
	return c.run.start()
}

// SetCommands allows services to add additional commands
// Any client command encountered which matches will send
// The resulting command down to RunCommand
// Commands must include at least a name and a heading
// Running SetCommands after calling Start or Listen will have no effect
func (c *Control) SetCommands(cmd ...*Command) error {
	for _, comm := range cmd {
		if comm.Name == "" {
			return errors.New("command requires Name")
		}

		switch comm.Heading {
		case DefaultGroup, MediaGroup, ActionGroup:
			continue
		default:
			return errors.New("Unsupported or nil Heading set")
		}
	}

	if e := c.run.setCommand(cmd...); e != nil {
		return e
	}

	return nil
}

// Notification appends the content of msg to a buffers notification file
// Any errors encountered during file opening/creation will be returned
// The canonical form of notification can be found in the markup libs' Notification type,
// And the output of the Parse() method can be used directly here
// For example
//     ntfy, err := markup.NewNotifier(buff, from, msg)
//     if err != nil {
//         log.Fatal(err)
//     }
//     fs.Notification(ntfy.Parse())
func (c *Control) Notification(buff, from, msg string) error {
	c.debug(ctlNotify, buff, from, msg)
	return c.run.notification(buff, from, msg)
}

// ErrorWriter returns a WriteCloser attached to a services' errors file
func (c *Control) ErrorWriter() (*WriteCloser, error) {
	return c.write.errorwriter()
}

// StatusWriter returns a WriteCloser attached to a buffers status file, which will as well send the correct event to the events file
func (c *Control) StatusWriter(buffer string) (*WriteCloser, error) {
	return c.write.fileWriter(buffer, "status")
}

// SideWriter returns a WriteCloser attached to a buffers `aside` file, which will as well send the correct event to the events file
func (c *Control) SideWriter(buffer string) (*WriteCloser, error) {
	return c.write.fileWriter(buffer, "aside")
}

// NavWriter returns a WriteCloser attached to a buffers nav file, which will as well send the correct event to the events file
func (c *Control) NavWriter(buffer string) (*WriteCloser, error) {
	return c.write.fileWriter(buffer, "navi")
}

// TitleWriter returns a WriteCloser attached to a buffers title file, which will as well send the correct event to the events file
func (c *Control) TitleWriter(buffer string) (*WriteCloser, error) {
	return c.write.fileWriter(buffer, "title")
}

// ImageWriter returns a WriteCloser attached to a named file in the buffers' image directory
func (c *Control) ImageWriter(buffer, resource string) (*WriteCloser, error) {
	return c.write.imageWriter(buffer, resource)

}

// MainWriter returns a WriteCloser attached to a buffers feed/document function to set the contents of a given buffers' document or feed file, which will as well send the correct event to the events file
func (c *Control) MainWriter(buffer, doctype string) (*WriteCloser, error) {
	return c.write.fileWriter(buffer, doctype)
}

func ctlLogger(msg ctlMsg, args ...interface{}) {
	l := log.New(os.Stdout, "ctl ", 0)

	switch msg {
	case ctlError:
		l.Printf("error: buffer=\"%s\" err=\"%v\"\n", args[0], args[1])
	case ctlEvent:
		l.Printf("event: msg=\"%s\"\n", args[0])
	case ctlCleanup:
		l.Println("cleanup: ending")
	case ctlCreate:
		l.Printf("create: buffer=\"%s\" doctype=%s\n", args[0], args[1])
	case ctlDelete:
		l.Printf("delete: buffer=\"%s\"\n", args[0])
	case ctlRemove:
		l.Printf("remove: buffer=\"%s\", filename=\"%s\"\n", args[0], args[1])
	case ctlStart:
		l.Printf("%s: starting\n", args[0])
	case ctlNotify:
		l.Printf("notify: buffer=\"%s\" from=\"%s\" msg=\"%s\"\n", args[0], args[1], args[2])
	case ctlDefault:
		cmd := args[0].(*Command)
		switch cmd.Heading {
		case ActionGroup:
			l.Printf("%s group=\"action\" arguments=\"%s\"\n", cmd.Name, cmd.Args)
		case MediaGroup:
			l.Printf("%s group=\"media\" arguments=\"%s\"\n", cmd.Name, cmd.Args)
		}
	}
}
