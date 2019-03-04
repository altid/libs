package cleanmark

import (
	"fmt"
	"io"
	"strings"
)

// Url represents a link markdown element, which is Stringable.
type Url struct {
	link []byte
	msg  []byte
}

// NewUrl returns a Url that will correctly stringify, or nil if the link element is empty.
// In the case of an empty link, it will return an error "No link provided for $msg"
// If message is empty, it will be set as the link.
func NewUrl(link, msg []byte) (*Url, error) {
	if len(link) == 0 {
		return nil, fmt.Errorf("No link provided for %s\n", msg)
	}
	if len(msg) == 0 {
		msg = link
	}
	url := &Url{
		link: link,
		msg:  msg,
	}
	return url, nil
}

// Implements Stringer interface
// Calling string will return the correctly markdown-formatted URL element
func (u *Url) String() string {
	return fmt.Sprintf("[%s](%s)", u.msg, u.link)
}

// Image represents an image markdown element, which is Stringable.
// All fields but path are optional, but it's strongly advised that any Image has alt text available.
type Image struct {
	alt  []byte
	path []byte
	msg  []byte
}

// NewImage returns an image that will correctly stringify, or nil and any errors encountered
func NewImage(path, msg, alt []byte) (*Image, error) {
	if len(alt) == 0 {
		alt = path
	}
	if len(path) == 0 {
		return nil, fmt.Errorf("No path provided for image")
	}
	img := &Image{
		alt:  alt,
		path: path,
		msg:  msg,
	}
	return img, nil
}

// Implements Stringer interface
// Calling string will return the correctly markdown-formatted image element
func (i *Image) String() string {
	return fmt.Sprintf("![%s](%s \"%s\")", i.alt, i.path, i.msg)
}

// Cleaner wraps an io.WriteCloser, allowing you to call various functions, sending formatted strings to the writer.
type Cleaner struct {
	w io.WriteCloser
}

func NewCleaner(w io.WriteCloser) *Cleaner {
	return &Cleaner{
		w: w,
	}
}

// Traditional Write method, this is used when you don't need to escape any text being written
// such as when the string is known at compile time.
// Calling Write([]bytes("this *should* be marked as bold")) would result in `this *should* be marked as bold` being written.
func (c *Cleaner) Write(msg []byte) (n int, err error) {
	return c.w.Write(msg)
}

// Variant of Write which accepts a string
func (c *Cleaner) WriteString(msg string) (n int, err error) {
	return io.WriteString(c.w, msg)
}

// WriteEscaped writes the properly escaped markdown to the underlying WriteCloser
// Any byte which may collide with a markdown element is converted as follows:
// `* -> \*`, `\ -> `\\`, etc.
// Calling WriteEscaped([]bytes("this *should* not be marked as bold")) would result in `this \*should\* not be marked as bold` being written.
func (c *Cleaner) WriteEscaped(msg []byte) (n int, err error) {
	return c.w.Write(escape(msg))
}

// Variant of WriteEscaped which accepts a string as input
func (c *Cleaner) WriteStringEscaped(msg string) (n int, err error) {
	return io.WriteString(c.w, escapeString(msg))
}

// Variant of Write which accepts a format specifier
func (c *Cleaner) Writef(format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.w, format, args...)
}

// Variant of WriteEscaped which accepts a format specifier
func (c *Cleaner) WritefEscaped(format string, args ...interface{}) (n int, err error) {
	return doWritef(c.w, format, args...)
}

// Variant of WriteEscaped which adds an nth-nested markdown list element to the underlying WriteCloser, using tabs to inset.
// Calling WriteList(2, []bytes("my list element")) would result in `		 - my list element` being written
func (c *Cleaner) WriteList(depth int, msg []byte) (n int, err error) {
	spaces := strings.Repeat("	", depth)
	if depth > 0 {
		spaces += "- "
	}
	return fmt.Fprintf(c.w, "%s%s", spaces, escape(msg))
}

// Variant of WriteList which accepts a format specifier
func (c *Cleaner) WritefList(depth int, format string, args ...interface{}) (n int, err error) {
	spaces := strings.Repeat("	", depth)
	if depth > 0 {
		spaces += "- "
	}
	return doWritef(c.w, spaces+format, args...)
}

// Variant of WriteEscaped which writes an nth degree markdown header element to the underlying WriteCloser
// Calling WriteHeader(2, []bytes("my header")) would result in `##my header` being written.
func (c *Cleaner) WriteHeader(degree int, msg []byte) (n int, err error) {
	hashes := strings.Repeat("#", degree)
	return fmt.Fprintf(c.w, "%s%s", hashes, escape(msg))
}

// Variant of WriteHeader which accepts a format specifier
func (c *Cleaner) WritefHeader(degree int, format string, args ...interface{}) (n int, err error) {
	hashes := strings.Repeat("#", degree)
	return doWritef(c.w, hashes+format, args...)
}

func (c *Cleaner) Close() {
	c.w.Close()
}

func doWritef(w io.WriteCloser, format string, args ...interface{}) (n int, err error) {
	for n := range args {
		switch f := args[n].(type) {
		case []byte:
			args[n] = escape(f)
		case string:
			args[n] = escapeString(f)
		}
	}
	return fmt.Fprintf(w, format, args...)
}

func escape(msg []byte) []byte {
	var offset int
	result := make([]byte, len(msg)*2)
	for i, c := range msg {
		switch c {
		case '*', '#', '_', '-', '\\', '/', '(', ')', '`', '[', ']':
			result[i+offset] = '\\'
			offset++
		}
		result[i+offset] = c
	}
	return result[:len(msg)+offset]
}

func escapeString(msg string) string {
	escaped := escape([]byte(msg))
	return string(escaped)
}
