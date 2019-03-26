package cleanmark

import (
	"strings"
	"unicode/utf8"
)

const (
	NormalText byte = iota
	ColorCode
	ColorText
	UrlLink
	UrlText
	ImagePath
	ImageText
	ImageLink
	EOF
)

type stateFn func(*Lexer) stateFn

// Lexer allows tokenizing of ubqt-flavored markdown for client-side parsers
type Lexer struct {
	src []byte
	start int
	width int
	pos   int
	items chan Item
	state stateFn
}

func NewLexer(src []byte) *Lexer {
	return &Lexer{
		src: src,
		items: make(chan Item, 2),
		state: lexText,
	}
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
		if strings.IndexByte("\\%", l.peek()) >= 0 {
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
		}
	}
}

func lexBack(l *Lexer) stateFn {
	l.ignore()
	l.accept("\\([])*_-~`")
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
		if strings.IndexByte("]\\", l.peek()) >= 0 {
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
  		case '\\':
			l.backup()
			l.emit(ColorText)
			l.accept("\\")
			l.ignore()
			l.accept("\\")
		}
	}
}

func lexColorCode(l *Lexer) stateFn {
	l.acceptRun("](")
	l.ignore()
	// Hex code
	l.acceptRun("#1234567890")
	// All valid chars from color code const
	l.acceptRun("abcdeghiklnoprtuwy")
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
		if  l.peek() == ')' {
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
	for strings.IndexByte(valid, l.nextChar()) >= 0 {
	}
	l.backup()
}
