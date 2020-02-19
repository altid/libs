 package markup

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

const (
	NormalText byte = iota
	ColorCode
	ColorText
	ColorTextBold
	ColorTextUnderline
	ColorTextStrike
	ColorTextEmphasis
	UrlLink
	UrlText
	ImagePath
	ImageText
	ImageLink
	BoldText
	StrikeText
	EmphasisText
	UnderlineText
	ErrorText
	EOF
)

type stateFn func(*Lexer) stateFn

// Lexer allows tokenizing of altid-flavored markdown for client-side parsers
type Lexer struct {
	src   []byte
	start int
	width int
	pos   int
	items chan Item
	state stateFn
}

func NewLexer(src []byte) *Lexer {
	return &Lexer{
		src:   src,
		items: make(chan Item, 2),
		state: lexText,
	}
}

func NewStringLexer(src string) *Lexer {
	return &Lexer{
		src: []byte(src),
		items: make(chan Item, 2),
		state: lexText,
	}
}

// Bytes wil return a parsed byte array from the input with markdown elements cleaned
// Any URL will be turned from `[some text](someurl)` to `some text (some url)`
// IMG will be turned from `![some text](someimage)` to `some text (some image)`
// color tags will be removed and the raw text will be output
func (l *Lexer) Bytes() []byte {
	var dst bytes.Buffer
	for {
		i := l.Next()
		switch i.ItemType {
		case EOF:
			return dst.Bytes()
		case ColorCode, ImagePath:
			continue
		case UrlLink, ImageLink:
			dst.WriteString(" (")
			dst.Write(i.Data)
			dst.WriteString(") ")
		default:
			dst.Write(i.Data)
		}
	}
	return dst.Bytes()
}

// String is the same as Bytes, but returns a string
func (l *Lexer) String() string {
	b := l.Bytes()
	return string(b)
}

// Item is returned from a call to Next()
// ItemType will be an ItemType
type Item struct {
	ItemType byte
	Data     []byte
}

// Next() returns the next Item from the tokenizer
// If ItemType is EOF, any subsequent calls to Next() will panic
func (l *Lexer) Next() Item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
}

func (l *Lexer) nextChar() byte {
	if l.pos >= len(l.src) {
		l.width = 0
		return EOF
	}
	rune, width := utf8.DecodeRune(l.src[l.pos:])
	l.width = width
	l.pos += l.width
	return byte(rune)
}

func lexText(l *Lexer) stateFn {
	for {
		if strings.IndexByte("\\%[!*-~_", l.peek()) >= 0 {
			if l.pos > l.start {
				l.emit(NormalText)
			}
		}
		switch l.nextChar() {
		case EOF:
			if l.pos > l.start {
				l.emit(NormalText)
			}
			l.emit(EOF)
			return nil
		case '\\':
			return lexBack
		case '%':
			return lexMaybeColor
		case '[':
			return lexMaybeUrl
		case '!':
			return lexMaybeImage
		case '*':
			return lexBold
		case '-':
			return lexEmphasis
		case '~':
			return lexStrike
		case '_':
			return lexUnderline
		}
	}
}

func lexBack(l *Lexer) stateFn {
	l.ignore()
	l.accept("\\!([])*_-~`")
	return lexText
}

func lexStrike(l *Lexer) stateFn {
	l.backup()
	l.emit(NormalText)
	l.accept("~")
	l.ignore()
	for {
		if l.peek() == '~' {
			l.emit(StrikeText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case '~':
			l.accept("~")
			l.ignore()
			return lexText
		}
	}
}

func lexUnderline(l *Lexer) stateFn {
	l.backup()
	l.emit(NormalText)
	l.accept("_")
	l.ignore()
	for {
		if l.peek() == '_' {
			l.emit(UnderlineText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case '_':
			l.accept("_")
			l.ignore()
			return lexText
		}
	}
}

func lexEmphasis(l *Lexer) stateFn {
	l.backup()
	l.emit(NormalText)
	l.accept("-")
	l.ignore()
	for {
		if l.peek() == '-' {
			l.emit(EmphasisText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case '-':
			l.accept("-")
			l.ignore()
			return lexText
		}
	}
}

// NOTE(halfwit): We would want to check for any possible token to switch state here in theory
// For the time being we'll hope everything is escaped
func lexBold(l *Lexer) stateFn {
	l.backup()
	l.emit(NormalText)
	l.accept("*")
	l.ignore()
	for {
		if l.peek() == '*' {
			l.emit(BoldText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case '*':
			l.accept("*")
			l.ignore()
			return lexText
		}
	}
	return lexText
}

func lexMaybeColor(l *Lexer) stateFn {
	switch l.nextChar() {
	case EOF:
		l.emit(EOF)
		return nil
	case '[':
		return lexColorText
	default:
		return lexText
	}
}

func lexColorText(l *Lexer) stateFn {
	l.accept("[")
	l.ignore()
	for {
		if strings.IndexByte("*~-_]\\", l.peek()) >= 0 {
			l.emit(ColorText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case ']':
			l.accept("]")
			l.ignore()
			return lexColorCode
		case '\\': // eat a single slash
			l.accept("\\")
			l.ignore()
			l.accept("\\!([])*_-~`")
		case '*':
			l.accept("*")
			l.ignore()
			return lexColorBold
		case '_':
			l.accept("_")
			l.ignore()
			return lexColorUnderline
		case '~':
			l.accept("~")
			l.ignore()
			return lexColorStrikeout
		case '-':
			l.accept("-")
			l.ignore()
			return lexColorEmphasis
		}
	}
}

func lexColorStrikeout(l *Lexer) stateFn {
	for {
		switch l.peek() {
		case '~', '\\':
			l.emit(ColorTextStrike)
		case ']':
			l.emit(ColorText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case '\\':
			l.accept("\\")
			l.ignore()
			l.accept("\\!([])*_-~`")
		case '~':
			l.accept("~")
			l.ignore()
			return lexColorText
		case ']':
			l.accept("]")
			l.ignore()
			return lexColorCode
		}
	}
}

func lexColorEmphasis(l *Lexer) stateFn {
	for {
		switch l.peek() {
		case '-', '\\':
			l.emit(ColorTextEmphasis)
		case ']':
			l.emit(ColorText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case '\\':
			l.accept("\\")
			l.ignore()
			l.accept("\\!([])*_-~`")
		case '-':
			l.accept("-")
			l.ignore()
			return lexColorText
		case ']':
			l.accept("]")
			l.ignore()
			return lexColorCode
		}
	}
}

func lexColorUnderline(l *Lexer) stateFn {
	for {
		switch l.peek() {
		case '_', '\\':
			l.emit(ColorTextUnderline)
		case ']':
			l.emit(ColorText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case '\\':
			l.accept("\\")
			l.ignore()
			l.accept("\\!([])*_-~`")
		case '_':
			l.accept("_")
			l.ignore()
			return lexColorText
		case ']':
			l.accept("]")
			l.ignore()
			return lexColorCode
		}
	}
}

func lexColorBold(l *Lexer) stateFn {
	for {
		switch l.peek() {
		case '*', '\\':
			l.emit(ColorTextBold)
		case ']':
			l.emit(ColorText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case '\\':
			l.accept("\\")
			l.ignore()
			l.accept("\\!([])*_-~`")
		case '*':
			l.accept("*")
			l.ignore()
			return lexColorText
		case ']':
			l.accept("]")
			l.ignore()
			return lexColorCode
		}
	}
}

func lexColorCode(l *Lexer) stateFn {
	l.acceptRun("](")
	l.ignore()
	// Hex code
	l.acceptRun("#1234567890,")
	l.acceptRun("abcdefghijklmnopqrstuvwxyz,")
	l.emit(ColorCode)
	l.accept(")")
	l.ignore()
	return lexText
}

func lexMaybeUrl(l *Lexer) stateFn {
	l.ignore()
	switch l.nextChar() {
	case EOF:
		l.emit(EOF)
		return nil
	case '!':
		return lexImageLinkText
	default:
		return lexUrlText
	}
}

func lexUrlText(l *Lexer) stateFn {
	for {
		if l.peek() == ']' {
			l.emit(ImageText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case ']':
			return lexUrlLink
		}
	}
}

func lexUrlLink(l *Lexer) stateFn {
	l.acceptRun("](")
	l.ignore()
	for {
		if l.peek() == ')' {
			l.emit(UrlLink)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case ')':
			l.accept(")")
			l.ignore()
			return lexText
		}
	}
}

// [![alt text](/path/to/image)](link)
func lexImageLinkText(l *Lexer) stateFn {
	l.acceptRun("[!")
	l.ignore()
	for {
		if l.peek() == ']' {
			l.emit(ImageText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case ']':
			return lexImageLinkPath
		}
	}
}

func lexImageLinkPath(l *Lexer) stateFn {
	l.acceptRun("](")
	l.ignore()
	for {
		if l.peek() == ')' {
			l.emit(ImagePath)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case ')':
			return lexImageLink
		}
	}
}

func lexImageLink(l *Lexer) stateFn {
	l.acceptRun(")](")
	l.ignore()
	for {
		if l.peek() == ')' {
			l.emit(ImageLink)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case ')':
			l.accept(")")
			l.ignore()
			return lexText
		}
	}
}

func lexMaybeImage(l *Lexer) stateFn {
	switch l.nextChar() {
	case EOF:
		l.emit(EOF)
		return nil
	case '[':
		return lexImageText
	default:
		return lexText
	}
}

func lexImageText(l *Lexer) stateFn {
	l.accept("[")
	l.ignore()
	for {
		if l.peek() == ']' {
			l.emit(ImageText)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case ']':
			return lexImagePath
		}
	}
}

func lexImagePath(l *Lexer) stateFn {
	l.acceptRun("](")
	l.ignore()
	for {
		if l.peek() == ')' {
			l.emit(ImagePath)
		}
		switch l.nextChar() {
		case EOF:
			l.emit(EOF)
			return nil
		case ')':
			l.accept(")")
			l.ignore()
			return lexText
		}
	}
}

func (l *Lexer) emit(t byte) {
	l.items <- Item{
		t,
		l.src[l.start:l.pos],
	}
	l.start = l.pos
}

func (l *Lexer) ignore() {
	l.start = l.pos
}

func (l *Lexer) backup() {
	l.pos -= l.width
}

func (l *Lexer) peek() byte {
	rune := l.nextChar()
	l.backup()
	return rune
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexByte(valid, l.nextChar()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for {
		if strings.IndexByte(valid, l.nextChar()) < 0 {
			l.backup()
			return
		}	
	}
}
