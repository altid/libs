package fs

import (
	"os"
	"path"
)

// WriteCloser is a type that implements io.WriteCloser
type WriteCloser struct {
	c      *Control
	fp     *os.File
	buffer string
}

func (w *WriteCloser) Write(b []byte) (n int, err error) {
	return w.fp.Write(b)
}

func (w *WriteCloser) Close() error {
	w.c.Event(w.buffer)
	return w.fp.Close()
}

// ErrorWriter returns a WriteCloser attached to a services' errors file
func (c *Control) ErrorWriter() (*WriteCloser, error) {
	ep := path.Join(c.rundir, "errors")

	fp, err := os.OpenFile(ep, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {

		return nil, err
	}

	w := &WriteCloser{
		c:      c,
		fp:     fp,
		buffer: "errors",
	}

	return w, nil
}

// StatusWriter returns a WriteCloser attached to a buffers status file, which will as well send the correct event to the events file
func (c *Control) StatusWriter(buffer string) (*WriteCloser, error) {
	return newWriteCloser(c, buffer, "status")
}

// SideWriter returns a WriteCloser attached to a buffers `aside` file, which will as well send the correct event to the events file
func (c *Control) SideWriter(buffer string) (*WriteCloser, error) {
	return newWriteCloser(c, buffer, "aside")
}

// NavWriter returns a WriteCloser attached to a buffers nav file, which will as well send the correct event to the events file
func (c *Control) NavWriter(buffer string) (*WriteCloser, error) {
	return newWriteCloser(c, buffer, "navi")
}

// TitleWriter returns a WriteCloser attached to a buffers title file, which will as well send the correct event to the events file
func (c *Control) TitleWriter(buffer string) (*WriteCloser, error) {
	return newWriteCloser(c, buffer, "title")
}

// ImageWriter returns a WriteCloser attached to a named file in the buffers' image directory
func (c *Control) ImageWriter(buffer, resource string) (*WriteCloser, error) {
	os.MkdirAll(path.Dir(path.Join(c.rundir, buffer, "images", resource)), 0755)
	return newWriteCloser(c, buffer, path.Join("images", resource))
}

// MainWriter returns a WriteCloser attached to a buffers feed/document function to set the contents of a given buffers' document or feed file, which will as well send the correct event to the events file
func (c *Control) MainWriter(buffer, doctype string) (*WriteCloser, error) {
	return newWriteCloser(c, buffer, doctype)
}

// Remove removes a buffer from the runtime dir. If the buffer doesn't exist, this is a no-op
func (c *Control) Remove(buffer, filename string) error {
	doc := path.Join(c.rundir, buffer, filename)
	// Don't try to delete that which isn't there
	if _, err := os.Stat(doc); os.IsNotExist(err) {
		return nil
	}
	event(c, doc)
	return os.Remove(doc)
}

func newWriteCloser(c *Control, buffer, doctype string) (*WriteCloser, error) {
	doc := path.Join(c.rundir, buffer, doctype)
	if doctype == "feed" {
		fp, err := os.OpenFile(doc, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {

			return nil, err
		}
		w := &WriteCloser{
			fp:     fp,
			c:      c,
			buffer: doc,
		}
		return w, nil
	}
	// Abuse truncation semantics of Create so we clear any old data
	fp, err := os.Create(doc)
	if err != nil {
		return nil, err
	}
	w := &WriteCloser{
		fp:     fp,
		c:      c,
		buffer: doc,
	}

	return w, nil
}
