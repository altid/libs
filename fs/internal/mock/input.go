package mock

import (
	"context"
	"fmt"
	"io"
)

type Input struct {
	reqs    chan string
	errlist []error
}

func NewInput(reqs chan string) *Input {
	return &Input{
		reqs: reqs,
	}
}

func (i *Input) Errs() []error {
	return i.errlist
}

func (i *Input) AddErr(format string, args ...interface{}) {
	e := fmt.Errorf(format, args...)
	i.errlist = append(i.errlist, e)
}

func (i *Input) Stop(ctx context.Context) {
	ctx.Done()
}

func (i *Input) Read(b []byte) (n int, err error) {
	for line := range i.reqs {
		n = copy(b, []byte(line+"\n"))
		return
	}

	return 0, io.EOF
}

func (i *Input) Close() error {
	return nil
}
