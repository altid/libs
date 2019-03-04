package cleanmark

import (
	"fmt"
	"io"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var empty struct{}

// HTMLCleaner wraps the underlying WriteCloser, and handles parsing HTML into ubqt-flavoured markdown, to the underlying writer.
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
// TODO: This parse is somewhat naive in how it handles certain elements, and may miss important imformation from a and img tags.
func (c *HTMLCleaner) Parse(r io.ReadCloser) error {
	z := html.NewTokenizer(r)
	for {
		if z.Next() == html.ErrorToken {
			return z.Err()
		}
		token := z.Token()
		// TODO: we will add a few more types here
		switch token.DataAtom {
		case atom.A:
			url, msg := parseUrl(z, token)
			fmt.Fprintf(c.w, "[%s](%s)", msg, url)
		case atom.B:
			fmt.Fprintf(c.w, "%c", '*')
		case atom.Br:
			fmt.Fprintf(c.w, "\n")
		case atom.Img:
			image, msg := parseImage(token)
			fmt.Fprintf(c.w, "![%s](%s)", msg, image)
		case atom.H1:
			switch token.Type {
			case html.StartTagToken:
				fmt.Fprintf(c.w, "%c", '#')
			case html.EndTagToken:
				fmt.Fprintf(c.w, "%c", '\n')
			}
		case atom.H2:
			switch token.Type {
			case html.StartTagToken:
				fmt.Fprintf(c.w, "%s", "##")
			case html.EndTagToken:
				fmt.Fprintf(c.w, "%c", '\n')
			}
		case atom.H3:
			switch token.Type {
			case html.StartTagToken:
				fmt.Fprintf(c.w, "%s", "###")
			case html.EndTagToken:
				fmt.Fprintf(c.w, "%c", '\n')
			}
		case atom.H4:
			switch token.Type {
			case html.StartTagToken:
				fmt.Fprintf(c.w, "%s", "####")
			case html.EndTagToken:
				fmt.Fprintf(c.w, "%c", '\n')
			}
		case atom.H5:
			switch token.Type {
			case html.StartTagToken:
				fmt.Fprintf(c.w, "%s", "#####")
			case html.EndTagToken:
				fmt.Fprintf(c.w, "%c", '\n')
			}
		case atom.H6:
			switch token.Type {
			case html.StartTagToken:
				fmt.Fprintf(c.w, "%s", "######")
			case html.EndTagToken:
				fmt.Fprintf(c.w, "%c", '\n')
			}
		//case atom.Ul
		case atom.P:
			if token.Type == html.StartTagToken {
				fmt.Fprintf(c.w, "\n	")
			}
		case atom.S:
			fmt.Fprintf(c.w, "-")
		case atom.U:
			fmt.Fprintf(c.w, "%c", '_')
		}
		if token.Type == html.TextToken {
			fmt.Fprintf(c.w, "%s", escapeString(token.Data))
		}
	}
}

// Write interface, simply writes the bytes to the underlying WriteCloser unmodified.
func (c *HTMLCleaner) Write(msg []byte) (n int, err error) {
	return c.w.Write(msg)
}

// This is the same as Write, except it accepts a string
func (c *HTMLCleaner) WriteString(msg string) (n int, err error) {
	return io.WriteString(c.w, msg)
}

// Close interface for the WriteCloser, this closes the underlying WriteCloser
func (c *HTMLCleaner) Close() {
	c.w.Close()
}

func parseUrl(z *html.Tokenizer, token html.Token) (link, url string) {
	for _, attr := range token.Attr {
		if attr.Key == "href" {
			url = attr.Val
		}
	}
	for {
		tt := z.Next()
		switch tt {
		case html.TextToken:
			link = string(z.Text())
		case html.EndTagToken:
			z.Next()
			return
		case html.ErrorToken:
			return
		}
	}
	return
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
