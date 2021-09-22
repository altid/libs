package util

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/altid/libs/service/input"
	"github.com/altid/libs/markup"
)

func RunInput(ctx context.Context, name string, h input.Handler, rd io.Reader, ew io.Writer) error {
	scanner := bufio.NewScanner(rd)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case err := <-errors:
			fmt.Fprintf(ew, "input error on %s: %v", name, err)
		case input := <- scanner.Bytes():
                        l := markup.NewLexer(input)
			if e := h.Handle(name, l); e != nil {
                             fmt.Fprintf(ew, "%v", e)
			}
		}
	}

	if e := scanner.Err(); e != io.EOF && e != nil {
		fmt.Fprintf(ew, "input error: %s: %v", name, e)
	}
}
