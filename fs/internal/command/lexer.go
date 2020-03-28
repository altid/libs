package command

import (
	"strings"
	"unicode/utf8"
)

type stateFn func(*lexer) stateFn

// Lexer allows tokenizing of altid-flavored markdown for client-side parsers
type lexer struct {
	src     []byte
	start   int
	width   int
	pos     int
	items   chan item
	state   stateFn
	heading ComGroup
}

type item struct {
	itemType byte
	data     []byte
}

func (l *lexer) next() item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			l.state = l.state(l)
		}
	}
}

func (l *lexer) nextChar() byte {
	if l.pos >= len(l.src) {
		l.width = 0
		return parserEOF
	}

	rune, width := utf8.DecodeRune(l.src[l.pos:])
	l.width = width
	l.pos += l.width

	return byte(rune)
}

func (l *lexer) emit(t byte) {
	l.items <- item{
		t,
		l.src[l.start:l.pos],
	}

	l.start = l.pos
}

func (l *lexer) peek() byte {
	rune := l.nextChar()
	l.backup()

	return rune
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexByte(valid, l.nextChar()) >= 0 {
		return true
	}

	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for {
		if strings.IndexByte(valid, l.nextChar()) < 0 {
			l.backup()
			return
		}
	}
}
