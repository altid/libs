package fs

import (
	"context"
	"fmt"
	"os"
)

type input struct {
	errlist []error
	fp      string
	ew      *os.File
}

func (i *input) errs() []error {
	return i.errlist
}

func (i *input) addErr(format string, args ...interface{}) {
	e := fmt.Errorf(format, args...)
	i.errlist = append(i.errlist, e)

	fmt.Fprintf(i.ew, format, args...)
}

func (i *input) stop(ctx context.Context) {
	ctx.Done()
	os.Remove(i.fp)
}
