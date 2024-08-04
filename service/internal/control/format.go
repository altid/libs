package control

import (
	"fmt"
	"os"
)

const (
	errorFmt = iota
	statusFmt
	sideFmt
	navFmt
	titleFmt
	feedFmt
	mainFmt
	imageFmt
	//notifyFmt
)

type prefix struct {
	c *Control
	fmt int
	nfd *os.File
	args []string
}

func newPrefix(c *Control, fmt int, args ...string) (*prefix, error) {
	// Do some sanity checking here and return error if we ever need to
	nfd, err := os.OpenFile(c.ctl.Name(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &prefix{
		nfd: nfd,
		c: c,
		fmt: fmt,
		args: args,
	}, nil
}

func (p *prefix) Write(b []byte) (int, error) {
	p.c.l.Lock()
	defer p.c.l.Unlock()
	switch p.fmt {
	case errorFmt: 
		return fmt.Fprintf(p.nfd, "error\n%s", b)
	case statusFmt:
		return fmt.Fprintf(p.nfd, "status %s\n\t%s", p.args[0], b)
	case navFmt:
		return fmt.Fprintf(p.nfd, "navi\n%s", b)
	case titleFmt:
		return fmt.Fprintf(p.nfd, "title %s\n\t%s", p.args[0], b)
	case feedFmt:
		return fmt.Fprintf(p.nfd, "feed %s\n\t%s", p.args[0], b)
	case imageFmt:
		return fmt.Fprintf(p.nfd, "image %s/%s\n\t%s", p.args[0], p.args[1], b)
	default:
		return 0, fmt.Errorf("unknown format specifier supplied\n")
	}
}

func (p *prefix) Close() error {
	p.c.l.Lock()
	defer p.c.l.Unlock()
	return p.nfd.Close()
}
