package markup

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Available markup colour codes, feel free to PR your favourite
const (
	White      = "white"
	Black      = "black"
	Blue       = "blue"
	Green      = "green"
	Red        = "red"
	Brown      = "brown"
	Purple     = "purple"
	Orange     = "orange"
	Yellow     = "yellow"
	LightGreen = "lightgreen"
	Cyan       = "cyan"
	LightCyan  = "lightcyan"
	LightBlue  = "lightblue"
	Pink       = "pink"
	Grey       = "grey"
	LightGrey  = "lightgrey"
)

var (
	hex3 = regexp.MustCompile("#[A-F]{3}")
	hex6 = regexp.MustCompile("#[A-F]{6}")
	code = regexp.MustCompile("white|black|blue|green|red|brown|purple|orange|yellow|lightgreen|cyan|lightcyan|lightblue|pink|grey|lightgrey")
)

// Color represents a color markdown element
// Valid values for code are any [markup constants], or color strings in hexadecimal form.
// #000000 to #FFFFFF, as well as #000 to #FFF. No alpha channel support currently exists.
type Color struct {
	code string
	msg  []byte
}

// NewColor returns a Color
// Returns error if color code is invalid
func NewColor(code string, msg []byte) (*Color, error) {
	if !validateColorCode(code) {
		return nil, fmt.Errorf("invalid color code %s", code)
	}
	color := &Color{
		code: code,
		msg:  msg,
	}
	return color, nil
}

func (c *Color) String() string {
	return fmt.Sprintf("%%[%s](%s)", escape(c.msg), c.code)
}

// URL represents a link markdown element
type URL struct {
	Link []byte
	Msg  []byte
}

// NewURL returns a URL, or an error if encountered
// If `msg` is empty, the contents of `link` will be used
// If `link` is empty, an error will be returned
// There are no assumptions about what link points to
// But before a beta release there may be a schema imposed
func NewURL(link, msg []byte) (*URL, error) {
	if len(link) == 0 {
		return nil, fmt.Errorf("no link provided for %s", msg)
	}
	if len(msg) == 0 {
		msg = link
	}
	url := &URL{
		Link: link,
		Msg:  msg,
	}
	return url, nil
}

// The form will be "[msg](link)"
func (u *URL) String() string { return fmt.Sprintf("[%s](%s)", u.Msg, u.Link) }

// Image represents an image markdown element
type Image struct {
	Src string
	Alt string
}

func (i *Image) String() string { return fmt.Sprintf("![%s](%s)", i.Alt, i.Src) }

// Cleaner represents a WriteCloser used to escape any altid markdown elements from a reader
type Cleaner struct {
	w io.WriteCloser
}

// NewCleaner returns a Cleaner
func NewCleaner(w io.WriteCloser) *Cleaner {
	return &Cleaner{
		w: w,
	}
}

// Write call the underlying WriteCloser's Write method
// Write does not modify the contents of msg
func (c *Cleaner) Write(msg []byte) (n int, err error) {
	return c.w.Write(msg)
}

// WriteString is a variant of Write which accepts a string as input
func (c *Cleaner) WriteString(msg string) (n int, err error) {
	return io.WriteString(c.w, msg)
}

// WriteEscaped writes the properly escaped markdown to the underlying WriteCloser
func (c *Cleaner) WriteEscaped(msg []byte) (n int, err error) {
	return c.w.Write(escape(msg))
}

// WriteStringEscaped is a variant of WriteEscaped which accepts a string as input
func (c *Cleaner) WriteStringEscaped(msg string) (n int, err error) {
	return io.WriteString(c.w, EscapeString(msg))
}

// Writef is a variant of Write which accepts a format specifier
func (c *Cleaner) Writef(format string, args ...any) (n int, err error) {
	return fmt.Fprintf(c.w, format, args...)
}

// WritefEscaped is a variant of WriteEscaped which accepts a format specifier
func (c *Cleaner) WritefEscaped(format string, args ...any) (n int, err error) {
	return doWritef(c.w, format, args...)
}

// WriteList is a variant of WriteEscaped which adds an nth-nested markdown list element to the underlying WriteCloser
func (c *Cleaner) WriteList(depth int, msg []byte) (n int, err error) {
	spaces := strings.Repeat("	", depth)
	if depth > 0 {
		spaces += "- "
	}
	return fmt.Fprintf(c.w, "%s%s", spaces, escape(msg))
}

// WritefList is a variant of WriteList which accepts a format specifier
func (c *Cleaner) WritefList(depth int, format string, args ...any) (n int, err error) {
	spaces := strings.Repeat("	", depth)
	if depth > 0 {
		spaces += "- "
	}
	return doWritef(c.w, spaces+format, args...)
}

// WriteHeader is a variant of WriteEscaped which writes an nth degree markdown header element to the underlying WriteCloser
func (c *Cleaner) WriteHeader(degree int, msg []byte) (n int, err error) {
	hashes := strings.Repeat("#", degree)
	return fmt.Fprintf(c.w, "%s%s", hashes, escape(msg))
}

// WritefHeader is a variant of WriteHeader which accepts a format specifier
func (c *Cleaner) WritefHeader(degree int, format string, args ...any) (n int, err error) {
	hashes := strings.Repeat("#", degree)
	return doWritef(c.w, hashes+format, args...)
}

// Close wraps the underlying WriteCloser's Close method
func (c *Cleaner) Close() {
	c.w.Close()
}

// Notifier is a type provided for use with fslib's Notification function
type Notifier struct {
	buff string
	from string
	msg  string
}

// NewNotifier returns a notifier ready for parsing
func NewNotifier(path, from, msg string) *Notifier {
	return &Notifier{
		buff: path,
		from: from,
		msg:  msg,
	}
}

// Parse will properly clean the markdown for the `from` and `msg` elements
// As well as format the lines to fit the notification idioms expected by clients
func (n *Notifier) Parse() (string, string, string) {
	from := "# " + EscapeString(n.from)
	msg := EscapeString(n.msg)
	return n.buff, from, msg
}

func doWritef(w io.WriteCloser, format string, args ...any) (n int, err error) {
	for n := range args {
		switch f := args[n].(type) {
		case []byte:
			args[n] = escape(f)
		case string:
			args[n] = EscapeString(f)
		}
	}
	return fmt.Fprintf(w, format, args...)
}

func escape(msg []byte) []byte {
	var b bytes.Buffer
	for _, c := range msg {
		switch c {
		case '*', '#', '_', '-', '~', '\\', '/', '(', ')', '`', '[', ']', '!':
			b.WriteRune('\'')
		}
		b.WriteByte(c)
	}
	return b.Bytes()
}

// EscapeString returns a properly escaped Altid markup string
func EscapeString(msg string) string {
	var result strings.Builder
	for _, c := range msg {
		switch c {
		case '*', '#', '_', '-', '~', '\\', '/', '(', ')', '`', '[', ']', '!':
			result.WriteRune('\\')
		}
		result.WriteRune(c)
	}
	return result.String()
}

func validateColorCode(s string) bool {
	switch {
	case code.MatchString(s):
		return true
	case hex3.MatchString(s):
		return true
	case hex6.MatchString(s):
		return true
	}
	return false
}
