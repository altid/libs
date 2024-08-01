package control

import (
	"bytes"
	"fmt"
)

const (
	errorFmt = "error\n%s"
	statusFmt = "status %s\n%s"
	sideFmt = "aside %s\n%s"
	navFmt = "navi\n%s"
	titleFmt = "title %s\n%s"
	feedFmt = "feed %s\n%s"
	mainFmt = "main %s\n%s"
	imageFmt = "image %s/%s\n%s"
	//notifyFmt
)

type prefix struct {
	c *Control
	fmt string
	args []string
	data *bytes.Buffer
}

func newPrefix(c *Control, fmt string, args ...string) (*prefix, error) {
	// Do some sanity checking here and return error if we ever need to
	var b bytes.Buffer
	return &prefix{
		c: c,
		fmt: fmt,
		args: args,
		data: &b,
	}, nil
}

// We don't put bytes to the ctl until the very end (close) as all implementations add one or two lines at a time
func (p *prefix) Write(b []byte) (int, error) {
	p.c.l.Lock()
	defer p.c.l.Unlock()
	if len(p.args) > 0 {
		return fmt.Fprintf(p.data, p.fmt + "\n", p.args, b)
	} 
	return fmt.Fprintf(p.data, p.fmt + "\n", b)
}

func (p *prefix) Close() error {
	p.c.l.Lock()
	defer p.c.l.Unlock()
	if p.data.Len() == 0 {
		return nil
	}

	if _, e := p.c.ctl.Write(p.data.Bytes()); e != nil {
		return e
	}
	return p.c.ctl.Close()
}
