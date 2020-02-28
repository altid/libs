// escape and parse markdown elements
package markup

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

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
// Valid values for code are any [markup constants], or color strings in hexidecimal form.
// #000000 to #FFFFFF, as well as #000 to #FFF. No alpha channel support currently exists.
type Color struct {
	code string
	msg  []byte
}

// NewColor returns a Color
// Returns error if color code is invalid
func NewColor(code string, msg []byte) (*Color, error) {
	if !validateColorCode(code) {
		return nil, fmt.Errorf("Invalid color code %s\n", code)
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

// Url represents a link markdown element
type Url struct {
	Link []byte
	Msg  []byte
}

// NewUrl returns a Url
// If `msg` is empty, the contents of `link` will be used
// If `link` is empty, an error will be returned
// There are no assumptions about what link points to
// But before a beta release there may be a schema imposed
func NewUrl(link, msg []byte) (*Url, error) {
	if len(link) == 0 {
		return nil, fmt.Errorf("No link provided for %s\n", msg)
	}
	if len(msg) == 0 {
		msg = link
	}
	url := &Url{
		Link: link,
		Msg:  msg,
	}
	return url, nil
}

// The form will be "[msg](link)"
func (u *Url) String() string {
	return fmt.Sprintf("[%s](%s)", u.Msg, u.Link)
}

// Image represents an image markdown element
type Image struct {
	alt  []byte
	path []byte
	msg  []byte
}

// NewImage returns an Image
// If `path` is empty, an error will be returned
// If both `img` and `alt` are empty, an error will be returned
// If either `img` or `alt` are empty, one will be substituted for the other
func NewImage(path, msg, alt []byte) (*Image, error) {
	if len(alt) == 0 && len(msg) == 0 {
		return nil, fmt.Errorf("No img or alt provided for path %s\n", path)
	}
	if len(path) == 0 {
		return nil, fmt.Errorf("No path provided for image")
	}
	if len(alt) == 0 {
		alt = msg
	}
	if len(msg) == 0 {
		msg = alt
	}
	img := &Image{
		alt:  alt,
		path: path,
		msg:  msg,
	}
	return img, nil
}

func (i *Image) String() string {
	return fmt.Sprintf("![%s](%s \"%s\")", i.alt, i.path, i.msg)
}

// Cleaner represents a WriteCloser used to escape any altid markdown elements from a reader
type Cleaner struct {
	w io.WriteCloser
}

// Returns a Cleaner
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

// Variant of WriteEscaped which accepts a string as input
func (c *Cleaner) WriteStringEscaped(msg string) (n int, err error) {
	return io.WriteString(c.w, EscapeString(msg))
}

// Variant of Write which accepts a format specifier
func (c *Cleaner) Writef(format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(c.w, format, args...)
}

// Variant of WriteEscaped which accepts a format specifier
func (c *Cleaner) WritefEscaped(format string, args ...interface{}) (n int, err error) {
	return doWritef(c.w, format, args...)
}

// Variant of WriteEscaped which adds an nth-nested markdown list element to the underlying WriteCloser
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
func (c *Cleaner) WriteHeader(degree int, msg []byte) (n int, err error) {
	hashes := strings.Repeat("#", degree)
	return fmt.Fprintf(c.w, "%s%s", hashes, escape(msg))
}

// Variant of WriteHeader which accepts a format specifier
func (c *Cleaner) WritefHeader(degree int, format string, args ...interface{}) (n int, err error) {
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

func doWritef(w io.WriteCloser, format string, args ...interface{}) (n int, err error) {
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
	var offset int
	result := make([]byte, len(msg)*2)
	for i, c := range msg {
		switch c {
		case '*', '#', '_', '-', '~', '\\', '/', '(', ')', '`', '[', ']', '!':
			result[i+offset] = '\\'
			offset++
		}
		result[i+offset] = c
	}
	return result[:len(msg)+offset]
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
