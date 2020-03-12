package fs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/altid/libs/markup"
)

// Handler is called whenever content is written to the associated `input` file
type Handler interface {
	Handle(path string, c *markup.Lexer) error
}

type Input struct {
	h      Handler
	r      *reader
	ew     *os.File
	rundir string
	fname  string
	debug  func(inputMsg, ...interface{})
}

type inputMsg int

const (
	inputnorm inputMsg = iota
	inputerr
)

// NewInput takes a Handler and the name of a buffer.
// This function returns an Input, or nil as well as any errors encountered
// If debug is true, it will write to stdout for all messages/errors received
func NewInput(h Handler, dir, buffer string, debug bool) (*Input, error) {
	// make directory for input on path
	if e := os.MkdirAll(dir, 0755); e != nil {
		return nil, e
	}

	inpath := path.Join(dir, buffer)

	if _, e := os.Stat(path.Join(inpath, "input")); os.IsNotExist(e) {
		r, err := newReader(path.Join(inpath, "input"))
		if err != nil {
			return nil, err
		}

		ep := path.Join(dir, "errors")

		ew, err := os.OpenFile(ep, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		i := &Input{
			ew:    ew,
			h:     h,
			r:     r,
			fname: inpath,
			debug: func(inputMsg, ...interface{}) {},
		}

		if debug {
			i.debug = inputLogging
		}

		return i, nil
	}

	return nil, fmt.Errorf("Input file already exist at %s", inpath)
}

// Start will watch for reads on Input's path, and send messages to the callers Handle function
// Errors will be logged to the errors file
func (i *Input) Start() {
	i.StartContext(context.TODO())
}

// StartContext is a variant of Start which takes a context for cancellation
func (i *Input) StartContext(ctx context.Context) {
	inputs := make(chan []byte)
	errors := make(chan error)

	i.debug(inputnorm, i.fname, "starting input")

	go func() {
		for msg := range inputs {
			l := markup.NewLexer(msg)
			if e := i.h.Handle(i.fname, l); e != nil {
				errors <- e
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(i.r)
		defer i.ew.Close()
		defer close(inputs)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				i.debug(inputnorm, i.fname, "closing input")
				return
			case err := <-errors:
				i.debug(inputerr, i.fname, err)
				fmt.Fprintf(i.ew, "input error on %s: %v", i.fname, err)
			case inputs <- scanner.Bytes():
				i.debug(inputnorm, i.fname, scanner.Bytes())
			}
		}

		if e := scanner.Err(); e != io.EOF && e != nil {
			fmt.Fprintf(i.ew, "input error on %s: %v", i.fname, e)
		}
	}()
}

func inputLogging(msg inputMsg, args ...interface{}) {
	switch msg {
	case inputnorm:
		fmt.Printf("input: chan=\"%s\" msg=\"%s\"\n", args[0], args[1])
	case inputerr:
		fmt.Printf("input: chan=\"%s\", err=\"%v\"\n", args[0], args[1])
	}
}
