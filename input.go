package fslib

import (
	"bufio"
	"context"
	"os"
	"path"
)

type Handler interface {
	Handle(path, msg string) error
}

type Input struct {
	h     Handler
	r     *reader
	fname string
}

// NewInput takes a Handler and the name of a buffer.
// This function returns an Input, or nil as well as any errors encountered
// NewInput does _not_ send an event at this time. This is to allow someone to use either Input or Ctrl, without requiring the other.
func NewInput(h Handler, dir, buffer string) (*Input, error) {
	// make directory for input on path
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	inpath := path.Join(dir, buffer)
	r, err := newReader(path.Join(inpath, "input"))
	if err != nil {
		return nil, err
	}
	i := &Input{
		h:     h,
		r:     r,
		fname: inpath,
	}
	return i, nil
}

// Start will watch for reads on Input's path, and send messages to the callers Handle function
// Any errors on from the Handler will cause this function to return, with the error message
func (i *Input) Start() error {
	return i.StartContext(context.TODO())
}

// StartContext is a variant of Start which takes a context for cancellation
func (i *Input) StartContext(ctx context.Context) error {
	inputMsg := make(chan string)
	errorMsg := make(chan error)
	defer close(inputMsg)
	go func() {
		// TODO(halfwit): Handle will be passed a type with access to a tokenizer
		// It needs a call to String() in case someone wants a pretty printed version of
		// Markdown input, such as removing %[some text](some color)
		// or ![Some image](/path/to/image), leaving in `*some text*`
		// and any particular thing like `_some text_` which are benign text elements
		// https://github.com/altid/fslib/issues/2
		for msg := range inputMsg {
			err := i.h.Handle(i.fname, msg)
			if err != nil {
				errorMsg <- err
				return
			}
		}
	}()
	scanner := bufio.NewScanner(i.r)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errorMsg:
			return err
		case inputMsg <- scanner.Text():
		}
	}
	return nil
}
