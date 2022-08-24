package command

import (
	"fmt"
	"strings"
)

const (
	// Don't override our other values
	cmdName byte = iota + 2
	cmdFrom
	cmdArgs
	cmdErr
)

func parseCmd(cmd string) (string, string, []string, error) {
	var name, from string
	var args []string

	l := &lexer{
		src:   []byte(cmd),
		items: make(chan item, 2),
		state: parseCmdName,
	}

	for {
		i := l.next()
		switch i.itemType {
		case cmdErr:
			return "", "", nil, fmt.Errorf("%s", i.data)
		case cmdName:
			name = string(i.data)
		case cmdFrom:
			from = string(i.data)
		case cmdArgs:
			args = strings.Fields(string(i.data))
		case parserEOF:
			// Clean up, possible on /quit, etc
			if from != "" && len(args) == 0 {
				args = strings.Fields(from)
				from = ""
			}

			return name, from, args, nil
		}
	}
}

func parseCmdName(l *lexer) stateFn {
	for {
		c := l.peek()

		if strings.IndexByte(" \t", c) >= 0 {
			if l.pos > l.start {
				l.emit(cmdName)
			}
		}

		if c == parserEOF {
			l.emit(cmdName)
		}

		switch l.nextChar() {
		case parserEOF:
			l.emit(parserEOF)
			return nil
		case ' ', '\t':
			l.acceptRun("\t ")
			l.ignore()
			return parseCmdFrom
		}
	}
}

func parseCmdFrom(l *lexer) stateFn {
	var inCompound bool

	for {
		switch inCompound {
		case true:
			if l.peek() == '"' {
				l.emit(cmdFrom)
			}

			switch l.nextChar() {
			case parserEOF:
				l.emit(parserEOF)
				return nil
			case '"':
				// Careful we don't eat the next tag
				l.accept("\"")
				l.accept(" ")
				l.ignore()
				return parseCmdArgs
			}
		case false:
			c := l.peek()

			if strings.IndexByte(" \t\n", c) >= 0 {
				if l.pos > l.start {
					l.emit(cmdFrom)
				}
			}

			// Catch early EOF here
			if c == parserEOF {
				l.emit(cmdFrom)
			}

			switch l.nextChar() {
			case parserEOF, '\n':
				l.emit(parserEOF)
				return nil
			case '"':
				l.accept("\"")
				l.ignore()
				inCompound = true
			case ' ', '\t':
				l.acceptRun("\t ")
				l.ignore()
				return parseCmdArgs
			}
		}
	}
}

func parseCmdArgs(l *lexer) stateFn {
	var inCompound bool

	for {
		switch inCompound {
		case true:
			if l.peek() == '"' {
				l.emit(cmdArgs)
			}

			switch l.nextChar() {
			case parserEOF, '"':
				l.emit(parserEOF)
				return nil
			}
		case false:
			if l.peek() == '\n' {
				l.emit(cmdArgs)
			}

			// If we don't find an opening quotation, send the whole string
			switch l.nextChar() {
			case parserEOF, '\n':
				l.emit(parserEOF)
				return nil
			case '"':
				l.accept("\"")
				l.ignore()
				inCompound = true
			}
		}
	}
}
