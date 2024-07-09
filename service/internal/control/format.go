package control

import "fmt"

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
}

func newPrefix(c *Control, fmt string, args ...string) (*prefix, error) {
	// Do some sanity checking here and return error if we ever need to
	return &prefix{
		c: c,
		fmt: fmt,
		args: args,
	}, nil
}

func (p *prefix) Write(b []byte) (int, error) {
	if len(p.args) > 0 {
		return fmt.Fprintf(p.c.ctl, p.fmt, p.args, b)
	} 
	return fmt.Fprintf(p.c.ctl, p.fmt, b)
}

func (p *prefix) Close() error {
	return nil
}