// helper library for parsing HTML for Altid markup
package html

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/altid/libs/markup"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var empty struct{}

// HTMLCleaner wraps the underlying WriteCloser, and handles parsing HTML into Altid-flavoured markdown, to the underlying writer.
type HTMLCleaner struct {
	w io.WriteCloser
	p Handler
}

type Handler interface {
	NavHandler
	ImgHandler
}

type NavHandler interface {
	Nav(*markup.Url) error
}

type ImgHandler interface {
	Img(string) error
}

// NewHTMLCleaner returns a usable HTMLCleaner struct
// if either w or p are nil it will return an error
func NewHTMLCleaner(w io.WriteCloser, p Handler) (*HTMLCleaner, error) {
	if w == nil {
		return nil, errors.New("nil WriteCloser")
	}

	if p == nil {
		return nil, errors.New("nil Handler")
	}

	h := &HTMLCleaner{
		w: w,
		p: p,
	}

	return h, nil
}

// Parse - This assumes properly formatted html, and will return an error from the underlying html tokenizer if encountered
// Parse writes properly formatted Altid markup to the underlying writer, translating many elements into their markdown form. This will be considered lossy, as the token metadata is ignored in all cases.
// This will return any errors encountered, and EOF on success
func (c *HTMLCleaner) Parse(r io.ReadCloser) error {
	z := html.NewTokenizer(r)
	tags := make(map[atom.Atom]bool)
	for {
		switch z.Next() {
		case html.ErrorToken:
			return z.Err()
		case html.StartTagToken:
			t := z.Token()
			// NOTE(halfwit): It's likely that this will grow much larger
			// due to how the tokenizer works, this is being done out of band
			if t.DataAtom == atom.A {
				switch {
				case tags[atom.Li]:
					fmt.Fprintf(c.w, " - ")
				case tags[atom.H1]:
					fmt.Fprintf(c.w, "# ")
				case tags[atom.H2]:
					fmt.Fprintf(c.w, "## ")
				case tags[atom.H3]:
					fmt.Fprintf(c.w, "### ")
				case tags[atom.H4]:
					fmt.Fprintf(c.w, "#### ")
				case tags[atom.H5]:
					fmt.Fprintf(c.w, "##### ")
				case tags[atom.H6]:
					fmt.Fprintf(c.w, "###### ")
				}
				url, msg := parseURL(z, t)
				fmt.Fprintf(c.w, "[%s](%s)", url, msg)
				continue
			}
			if t.DataAtom == atom.Img {
				image, msg := parseImage(t)
				fmt.Fprintf(c.w, "![%s](%s)", msg, image)
				if i, ok := c.p.(ImgHandler); ok {
					go i.Img(image)
				}
				continue
			}
			if t.DataAtom == atom.Nav {
				if i, ok := c.p.(NavHandler); ok {
					for n := range parseNav(z, t) {
						i.Nav(n)
					}
				}
				continue
			}
			tags[t.DataAtom] = true
		case html.EndTagToken:
			t := z.Token().DataAtom
			fmt.Fprintf(c.w, "%s", endToken(t))
			tags[t] = false
		case html.TextToken:
			data := parseToken(z.Token(), tags)
			fmt.Fprintf(c.w, "%s", data)
		case html.SelfClosingTagToken:
			t := z.Token()
			if t.DataAtom == atom.Img {
				image, msg := parseImage(t)
				fmt.Fprintf(c.w, "![%s](%s)", msg, image)
				if i, ok := c.p.(ImgHandler); ok {
					go i.Img(image)
				}
				continue
			}
			data := parseToken(t, tags)
			fmt.Fprintf(c.w, "%s", data)
		}
	}
}

// Write calls the underlying WriteCloser's Write method. It does not modify the contents of `msg`
func (c *HTMLCleaner) Write(msg []byte) (n int, err error) {
	return c.w.Write(msg)
}

// WriteString is the same as Write, except it accepts a string instead of bytes.
func (c *HTMLCleaner) WriteString(msg string) (n int, err error) {
	return io.WriteString(c.w, msg)
}

// Close calls the underlying WriteCloser's Close method.
func (c *HTMLCleaner) Close() {
	c.w.Close()
}

func endToken(t atom.Atom) string {
	// insert any newlines, etc before we finish up
	switch t {
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6, atom.P:
		return "\n\n"
	case atom.Li, atom.Ul:
		return "\n"
	case atom.B:
		return "*"
	case atom.Em:
		return "-"
	case atom.Strike:
		return "~~"
	case atom.U:
		return "_"
	}
	return ""
}

func parseToken(t html.Token, m map[atom.Atom]bool) string {
	// NOTE(halfwit): This is rather messy right now, and will need a revisit
	var dst bytes.Buffer

	if m[atom.Script] || m[atom.Head] {
		return ""
	}

	switch {
	case m[atom.H1]:
		dst.WriteString("# ")
	case m[atom.H2]:
		dst.WriteString("## ")
	case m[atom.H3]:
		dst.WriteString("### ")
	case m[atom.H4]:
		dst.WriteString("#### ")
	case m[atom.H5]:
		dst.WriteString("##### ")
	case m[atom.H6]:
		dst.WriteString("###### ")
	case m[atom.Strike]:
		dst.WriteString("~~")
	case m[atom.Em]:
		dst.WriteString("-")
	case m[atom.B]:
		dst.WriteString("*")
	case m[atom.U]:
		dst.WriteString("_")
	case m[atom.P]:
		dst.WriteString("  ")
	// TODO: Get ordered list numberings
	case m[atom.Li]:
		dst.WriteString(" - ")
	}

	d := t.Data

	// If all we had is whitespace, don't return anything
	if strings.TrimSpace(d) == "" {
		return ""
	}

	dst.WriteString(markup.EscapeString(d))
	return dst.String()
}

// TODO: Give back triple, containing link, url, image
// Switch on image == ""
func parseURL(z *html.Tokenizer, t html.Token) (link, url string) {
	for _, attr := range t.Attr {
		if attr.Key == "href" {
			url = attr.Val
		}
	}
	for {
		tt := z.Next()
		switch tt {
		case html.StartTagToken:
			// We'll have to revisit the interface for occasions such as what follows:
			// <a href="somesite"></img someimage></a>
			// ![[somesite](someimage)](linktosite)
			// Additionally, scrub out any newlines
			// <a href="http://pressbooks.com>
			//   <img src="assets/pressbooks-promo.png" alt="pressbooks.com"/>
			// </a>
			if z.Token().DataAtom == atom.Img {

			}
		case html.SelfClosingTagToken:
			link = string(z.Text())
		case html.TextToken:
			link = string(z.Text())
		case html.EndTagToken:
			return
		case html.ErrorToken:
			return
		}
	}
}

func parseImage(token html.Token) (image, alt string) {
	for _, attr := range token.Attr {
		switch attr.Key {
		case "src":
			image = attr.Val
		case "alt":
			alt = attr.Val
		}
	}
	return
}

func parseNav(z *html.Tokenizer, t html.Token) chan *markup.Url {
	m := make(chan *markup.Url)
	go func() {
		defer close(m)
		for {
			switch z.Next() {
			case html.StartTagToken:
				t := z.Token()
				if t.DataAtom == atom.Nav {
					return
				}
				if t.DataAtom != atom.A {
					continue
				}
				link, url := parseURL(z, t)
				m <- &markup.Url{
					Link: []byte(link),
					Msg:  []byte(url),
				}
			case html.EndTagToken:
				if z.Token().DataAtom == atom.Nav {
					return
				}
			case html.ErrorToken:
				return
			}
		}
	}()
	return m
}
