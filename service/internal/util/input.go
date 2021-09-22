package util

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/altid/libs/fs/input"
	"github.com/altid/libs/markup"
)

func RunInput(ctx context.Context, name string, h input.Handler, rd io.Reader, ew io.Writer) error {
	inputs := make(chan []byte)
	errors := make(chan error)

	go func() {
		for msg := range inputs {
			l := markup.NewLexer(msg)
			if e := h.Handle(name, l); e != nil {
				fmt.Fprintf(ew, "%v", e)
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(rd)
		defer close(inputs)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case err := <-errors:
				fmt.Fprintf(ew, "input error on %s: %v", name, err)
			case inputs <- scanner.Bytes():
			}
		}

		if e := scanner.Err(); e != io.EOF && e != nil {
			fmt.Fprintf(ew, "input error: %s: %v", name, e)
		}
	}()

	return nil
}
