package cleanmark

import (
	"io"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var empty struct {}
type HTMLCleaner struct {
	w io.WriteCloser
}

// Return a cleaner ready to go for HTML	
func NewHTMLCleaner(w io.WriteCloser) *HTMLCleaner {
	return &HTMLCleaner{
		w: w,
	}
}

// Parse - This assumes properly formatted html
// This will write properly formatted ubqt markup to the underlying writer
// This returns any errors that it encounters, or EOF once it's exhausted the reader.
// TODO: This parse is somewhat naive in how it handles a and img tags
func (c *HTMLCleaner) Parse(r io.ReadCloser) error {

	//var url, img, imgAlttext []byte
	flags := make(map[atom.Atom]struct{})
	z := html.NewTokenizer(r)
	for {
		ttype := z.Next()
		token := z.Token()

		// We'll handle images and links seperately, as they require special attention
		switch token.DataAtom {
		case atom.A:	
			continue
		case atom.Img:
			continue
		}

		// All other types simply augment the writer state
		switch ttype {
		case html.ErrorToken:
			return z.Err()	
		case html.StartTagToken:
			flags[token.DataAtom] = empty
		case html.EndTagToken:
			delete(flags, z.Token().DataAtom)
		case html.SelfClosingTagToken:
			// TODO: so far we care about img, br, col,
			// see what else we need to handle
		case html.TextToken: // This is where the meat happens
			c.w.Write(z.Text())
		}
	}
}

// Write - Write normal strings to the underlying stream, unmodified
func (c *HTMLCleaner) Write(msg []byte) (n int, err error) {
	return c.w.Write(msg)
}

func (c *HTMLCleaner) WriteString(msg string) (n int, err error) {
	return io.WriteString(c.w, msg)
}

func (c *HTMLCleaner) Close() {
	c.w.Close()
}
