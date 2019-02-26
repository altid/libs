package cleanmark

import (
	"fmt"
	"io"
	"strings"
)

type Url struct {
	link []byte
	msg []byte
}

func NewUrl(link, msg []byte) *Url {
	return &Url{
		link: link,
		msg: msg,
	}
}

func (u *Url) String() string {
	return fmt.Sprintf("[%s](%s)", u.msg, u.link)
}

type Image struct {
	alt []byte
	path []byte
	msg []byte
}

func NewImage(path, msg, alt []byte) *Image {
	return &Image{
		alt: alt,
		path: path,
		msg: msg,
	}
}

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
func (c *Cleaner) Write(msg []byte) (n int, err error) {
	return c.w.Write(msg)
}

// Simple WriteString wrapper, this is used when you don't need to escape any text being written
func (c *Cleaner) WriteString(msg string) (n int, err error) {
	return io.WriteString(c.w, msg)
}

// WriteEscaped takes a string as input, and writes the properly escaped markdown to the underlying WriteCloser
func (c *Cleaner) WriteEscaped(msg []byte) (n int, err error) {
	return c.w.Write(escape(msg))
}

// Escaped variant of WriteString
func (c *Cleaner) WriteStringEscaped(msg string) (n int, err error) {
	return io.WriteString(c.w, escapeString(msg))
}

// Writef - variant of Write which accepts a format string as the first argument
func (c *Cleaner) Writef(format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.w, format, args...)
}

// WritefEscaped - variant of WriteEscaped which accepts a format string as the first argument
func (c *Cleaner) WritefEscaped(format string, args ...interface{}) (n int, err error) {
	return doWritef(c.w, format, args...)
}

// WriteList - Variant of Write which adds an nth-nested list to the underlying WriteCloser
// This will also escape any markdown elements
func (c *Cleaner) WriteList(depth int, msg []byte) (n int, err error) {
	spaces := strings.Repeat("	", depth)
	if depth > 0 {
		spaces += "- "
	}
	return fmt.Fprintf(c.w, "%s%s", spaces, escape(msg))
}

// WritefList - variant of WriteList which accepts a format string 
func (c *Cleaner) WritefList(depth int, format string, args ...interface{}) (n int, err error) {
	spaces := strings.Repeat("	", depth)
	if depth > 0 {
		spaces += "- "
	}
	return doWritef(c.w, spaces + format, args...)
}

// WriteHeader - Variant of Write which writes an nth degree header to the underlying WriteCloser
// This will also escape any markdown elements
func (c *Cleaner) WriteHeader(degree int, msg []byte) (n int, err error) {
	hashes := strings.Repeat("#", degree)
	return fmt.Fprintf(c.w, "%s%s", hashes, escape(msg))
}

// WritefHeader - variant of WriteHeader which accepts a format string
func (c *Cleaner) WritefHeader(degree int, format string, args ...interface{}) (n int, err error) {
	hashes := strings.Repeat("#", degree)
	return doWritef(c.w, hashes + format, args...)
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
	result := make([]byte, len(msg) * 2)
	for i, c := range msg {
		switch c {
		case '*', '#', '_', '-', '\\', '/', '(', ')', '`', '[', ']':
			result[i+offset] = '\\'
			offset++
		}
		result[i + offset] = c
	}
	return result[:len(msg)+offset]
}

func escapeString(msg string) string {
	escaped := escape([]byte(msg))
	return string(escaped)
}
