package defaults

import (
	"context"
	"fmt"
	"os"
)

type Input struct {
	errlist []error
	fp      string
	ew      *os.File
}

func NewInput(ew *os.File, fp string) *Input {
	return &Input{
		fp: fp,
		ew: ew,
	}
}

func (i *Input) Errs() []error {
	return i.errlist
}

func (i *Input) AddErr(format string, args ...interface{}) {
	e := fmt.Errorf(format, args...)
	i.errlist = append(i.errlist, e)

	fmt.Fprintf(i.ew, format, args...)
}

func (i *Input) Stop(ctx context.Context) {
	ctx.Done()
	os.Remove(i.fp)
}
