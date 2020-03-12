package fs

import (
	"context"
	"fmt"
	"io"
)

type mockinput struct {
	reqs    chan string
	errlist []error
}

func (i *mockinput) errs() []error {
	return i.errlist
}

func (i *mockinput) addErr(format string, args ...interface{}) {
	e := fmt.Errorf(format, args...)
	i.errlist = append(i.errlist, e)
}

func (i *mockinput) stop(ctx context.Context) {
	ctx.Done()
}

func (i *mockinput) Read(b []byte) (n int, err error) {
	for line := range i.reqs {
		n = copy(b, []byte(line+"\n"))
		return
	}

	return 0, io.EOF
}

func (i *mockinput) Close() error {
	return nil
}
