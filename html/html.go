// Package html helper library for parsing HTML for Altid markup
package html

import (
	"errors"
	"io"
)

// Cleaner wraps the underlying WriteCloser, and handles parsing HTML into Altid-flavoured markdown, to the underlying writer.
type Cleaner struct {
	w io.WriteCloser
	p Handler
}

// Handler will be called to satisfy both <nav> and <img> elements
type Handler interface {
	NavHandler
	ImgHandler
}

// NavHandler is called when the parser encounters a <nav> element
type NavHandler interface {
	Nav(*URL) error
}

// ImgHandler is called when the parser encounters an <img> element
type ImgHandler interface {
	Img(*Image) error
}

// NewCleaner returns a usable Cleaner struct
// if either w or p are nil it will return an error
func NewCleaner(w io.WriteCloser, p Handler) (*Cleaner, error) {
	if w == nil {
		return nil, errors.New("nil WriteCloser")
	}
	if p == nil {
		return nil, errors.New("nil Handler")
	}
	h := &Cleaner{
		w: w,
		p: p,
	}
	return h, nil
}

// Parse - This assumes properly formatted html, and will return an error from the underlying html tokenizer if encountered
// Parse writes properly formatted Altid markup to the underlying writer, translating many elements into their markdown form. This will be considered lossy, as the token metadata is ignored in all cases.
// This will return any errors encountered, and EOF on success
func (c *Cleaner) Parse(r io.ReadCloser) error { return parse(c, r) }

// Write calls the underlying WriteCloser's Write method. It does not modify the contents of `msg`
func (c *Cleaner) Write(msg []byte) (n int, err error) { return c.w.Write(msg) }

// WriteString is the same as Write, except it accepts a string instead of bytes.
func (c *Cleaner) WriteString(msg string) (n int, err error) { return io.WriteString(c.w, msg) }

// Close calls the underlying WriteCloser's Close method.
func (c *Cleaner) Close() { c.w.Close() }
