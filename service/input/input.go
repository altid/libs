package input

import (
	"github.com/altid/libs/markup"
)

// Handler is called when data is written to an `input` file
// The path will be the buffer in which the data was written
type Handler interface {
	Handle(path string, c *markup.Lexer) error
}
