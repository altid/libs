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
}

// NewInput takes a Handler and the name of a buffer.
// This function returns an Input, or nil as well as any errors encountered
func NewInput(h Handler, dir, buffer string) (*Input, error) {
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
	inputMsg := make(chan []byte)
	errorMsg := make(chan error)

	go func() {
		for msg := range inputMsg {
			l := markup.NewLexer(msg)
			if e := i.h.Handle(i.fname, l); e != nil {
				errorMsg <- e
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(i.r)
		defer i.ew.Close()
		defer close(inputMsg)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case err := <-errorMsg:
				fmt.Fprintf(i.ew, "input error on %s: %v", i.fname, err)
			case inputMsg <- scanner.Bytes():
			}
		}

		if e := scanner.Err(); e != io.EOF && e != nil {
			fmt.Fprintf(i.ew, "input error on %s: %v", i.fname, e)
		}
	}()
}
