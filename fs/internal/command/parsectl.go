package command

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

/* Files are simple
mything:
	name <arg> #comment
	name|othername <arg1> <arg2> #comment
*/

const (
	parserError byte = iota
	parserEOF
	parserHeading
	parserNewEntry
	parserEntryName
	parserEntryArgs
	parserEntryAlias
	parserEntryDesc
)

// Parse returns any commands found within the byte array
func ParseCtlFile(b []byte) ([]*Command, error) {
	var cmdlist []*Command

	l := &lexer{
		src:     b,
		items:   make(chan item, 2),
		state:   parseHeading,
		heading: 9001,
	}

	for {
		c, err := parseCtlFile(l)
		switch err {
		case io.EOF:
			if c.Name != "" {
				cmdlist = append(cmdlist, c)
			}

			return cmdlist, nil
		case nil:
			if c.Name != "" {
				cmdlist = append(cmdlist, c)
			}

			continue
		default:
			return nil, err
		}
	}
}

func parseCtlFile(l *lexer) (*Command, error) {
	cmd := &Command{}

	for {
		i := l.next()

		switch i.itemType {
		case parserEOF:
			cmd.Heading = l.heading
			return cmd, io.EOF
		case parserError:
			return nil, fmt.Errorf("%s", i.data)
		case parserHeading:
			heading, err := headingFromString(i.data)
			if err != nil {
				return nil, err
			}

			l.heading = heading
			return cmd, nil
		case parserNewEntry:
			cmd.Heading = l.heading
			return cmd, nil
		case parserEntryName:
			cmd.Name = string(i.data)
		case parserEntryAlias:
			cmd.Alias = append(cmd.Alias, string(i.data))
		case parserEntryArgs:
			cmd.Args = append(cmd.Args, string(i.data))
		case parserEntryDesc:
			cmd.Description = string(i.data)
		}
	}
}

func parseHeading(l *lexer) stateFn {
	for {
		if l.peek() == ':' {
			l.emit(parserHeading)
		}

		switch l.nextChar() {
		case parserEOF:
			l.src = []byte("found heading with no body")
			l.start = 0
			l.pos = len(l.src)

			l.emit(parserError)
			return nil
		case '\n':
			l.src = []byte("malformed header: no ending colon")
			l.start = 0
			l.pos = len(l.src)

			l.emit(parserError)
			return nil
		case ':':
			l.acceptRun(":\n\t ")
			l.ignore()
			return parseEntryName
		}
	}
}

func parseMaybeHeading(l *lexer) stateFn {
	for {
		switch l.nextChar() {
		case parserEOF:
			l.emit(parserEOF)
			return nil
		case ' ', '\t':
			l.acceptRun("\t ")
			l.ignore()
			l.emit(parserNewEntry)
			return parseEntryName
		default:
			l.backup()
			return parseHeading
		}

	}
}

// Possible chars: " ", "<", "#", "\t", or a heading
func parseEntryAmbiguous(l *lexer) stateFn {
	for {
		switch l.nextChar() {
		case parserEOF:
			l.emit(parserEOF)
			return nil
		case '\n':
			// Whitespace means heading
			if l.peek() == ' ' || l.peek() == '\t' {
				l.acceptRun(" \t")
				l.ignore()
				l.emit(parserNewEntry)
				return parseEntryName
			}
			l.accept("\n")
			l.ignore()
			return parseHeading
		case '<':
			l.accept("<")
			l.ignore()
			return parseEntryArg
		case '#':
			l.acceptRun("# \t")
			l.ignore()
			return parseEntryDesc
		case ' ', '\t':
			l.acceptRun(" \t")
			l.ignore()

		}
	}
}

// Possible chars: " ", "|", "\n", entry name chars
func parseEntryName(l *lexer) stateFn {
	for {
		if strings.IndexByte("| \t", l.peek()) >= 0 {
			if l.pos > l.start {
				l.emit(parserEntryName)
			}
		}

		switch l.nextChar() {
		case parserEOF:
			if l.pos > l.start {
				l.emit(parserEntryName)
			}

			l.emit(parserEOF)
			return nil
		case ' ', '\t':
			l.acceptRun(" \t")
			l.ignore()
			return parseEntryAmbiguous

		case '|':
			l.accept("|")
			l.ignore()
			return parseEntryAlias
		}
	}
}

func parseEntryAlias(l *lexer) stateFn {
	for {
		if strings.IndexByte("| \t", l.peek()) >= 0 {
			if l.pos > l.start {
				l.emit(parserEntryAlias)
			}
		}

		switch l.nextChar() {
		case parserEOF:
			if l.pos > l.start {
				l.emit(parserEntryAlias)
			}

			l.emit(parserEOF)
			return nil
		case '\n':
			l.accept("\n")
			l.ignore()
			return parseMaybeHeading
		case ' ', '\t':
			l.acceptRun("\t ")
			l.ignore()
			return parseEntryAmbiguous
		case '|':
			// We found another alias, stay in function
			l.accept("|")
			l.ignore()
			return parseEntryAlias
		}
	}
}

func parseEntryArg(l *lexer) stateFn {
	for {
		if l.peek() == '>' {
			if l.pos > l.start {
				l.emit(parserEntryArgs)
			}
		}

		switch l.nextChar() {
		case parserEOF:
			l.emit(parserError)
			return nil
		case '\n':
			l.accept("\n")
			l.ignore()
			return parseMaybeHeading
		case '>':
			l.acceptRun("> ")
			l.ignore()
			return parseEntryAmbiguous
		}
	}
}

func parseEntryDesc(l *lexer) stateFn {
	for {
		if l.peek() == '\n' {
			if l.pos > l.start {
				l.emit(parserEntryDesc)
			}

		}
		switch l.nextChar() {
		case parserEOF:
			l.emit(parserEntryDesc)
			return nil
		case '\n':
			l.accept("\n")
			l.ignore()
			return parseMaybeHeading
		}

	}
}

func headingFromString(b []byte) (ComGroup, error) {
	switch string(b) {
	case "general":
		return DefaultGroup, nil
	case "media":
		return MediaGroup, nil
	case "emotes":
		return ActionGroup, nil
	case "service":
		return ServiceGroup, nil
	default:
		return 0, errors.New("unknown heading")
	}
}
