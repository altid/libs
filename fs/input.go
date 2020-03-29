package fs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/altid/libs/fs/internal/defaults"
	"github.com/altid/libs/fs/internal/mock"
	"github.com/altid/libs/fs/internal/reader"
	"github.com/altid/libs/markup"
)

// Handler is called whenever content is written to the associated `input` file
type Handler interface {
	Handle(path string, c *markup.Lexer) error
}

type inputter interface {
	Stop(context.Context)
}

// Input allows an Altid service to listen for writes to a file named input for a given buffer
type Input struct {
	h     Handler
	r     io.ReadCloser
	run   inputter
	err   io.Writer
	fname string
	buff  string
	ctx   context.Context
	debug func(inputMsg, ...interface{})
}

type inputMsg int

const (
	inputnorm inputMsg = iota
	inputerr
)

// NewInput takes a Handler and the name of a buffer.
// This function returns an Input, or nil as well as any errors encountered
// If debug is true, it will write to stdout for all messages/errors received
func NewInput(h Handler, dir, buffer string, ew io.Writer, debug bool) (*Input, error) {
	// make directory for input on path
	if e := os.MkdirAll(dir, 0755); e != nil {
		return nil, e
	}

	inpath := path.Join(dir, buffer)
	fp := path.Join(inpath, "input")

	if _, e := os.Stat(fp); os.IsNotExist(e) {

		r, err := reader.New(fp)
		if err != nil {
			return nil, err
		}

		ew, err := os.OpenFile(path.Join(dir, "errors"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		i := &Input{
			run:   defaults.NewInput(ew, fp),
			h:     h,
			r:     r,
			err:   ew,
			fname: inpath,
			buff:  buffer,
			debug: func(inputMsg, ...interface{}) {},
		}

		if debug {
			i.debug = inputLogging
		}

		return i, nil
	}

	return nil, fmt.Errorf("Input file already exist at %s", inpath)
}

// NewMockInput returns a faked input file for testing
// All writes to `reqs` will trigger the Handler internally
func NewMockInput(h Handler, buffer string, ew io.Writer, debug bool, reqs chan string) (*Input, error) {
	mci := mock.NewInput(reqs)

	i := &Input{
		h:     h,
		r:     mci,
		run:   mci,
		err:   ew,
		fname: buffer,
		buff:  buffer,
		debug: func(inputMsg, ...interface{}) {},
	}

	if debug {
		i.debug = inputLogging
	}

	return i, nil
}

// Start will watch for reads on Input's path, and send messages to the callers Handle function
// Errors will be logged to the errors file
func (i *Input) Start() {
	i.StartContext(context.Background())
}

// StartContext is a variant of Start which takes a context for cancellation
func (i *Input) StartContext(ctx context.Context) {
	inputs := make(chan []byte)
	errors := make(chan error)

	i.debug(inputnorm, i.fname, "starting input")
	i.ctx = ctx

	go func() {
		for msg := range inputs {
			l := markup.NewLexer(msg)
			if e := i.h.Handle(i.buff, l); e != nil {
				fmt.Fprintf(i.err, "%v", e)
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(i.r)
		defer close(inputs)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				i.debug(inputnorm, i.fname, "closing input")
				return
			case err := <-errors:
				i.debug(inputerr, i.fname, err)
				fmt.Fprintf(i.err, "input error on %s: %v", i.fname, err)
			case inputs <- scanner.Bytes():
				i.debug(inputnorm, i.fname, scanner.Bytes())
			}
		}

		if e := scanner.Err(); e != io.EOF && e != nil {
			fmt.Fprintf(i.err, "input error: %s: %v", i.fname, e)
		}
	}()
}

// Stop ends the Input session, cleaning up after itself
func (i *Input) Stop() {
	i.run.Stop(i.ctx)
}

func inputLogging(msg inputMsg, args ...interface{}) {
	l := log.New(os.Stdout, "input: ", 0)

	switch msg {
	case inputnorm:
		l.Printf("input: chan=\"%s\" msg=\"%s\"\n", args[0], args[1])
	case inputerr:
		l.Printf("input: chan=\"%s\", err=\"%v\"\n", args[0], args[1])
	}
}
