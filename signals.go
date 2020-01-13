// build !plan9

package main

import (
	"context"
	"fmt"
)

func handleSig(ctx context.Context, signal string) {
	switch signal {
	case "interrupt":
		ctx.Done()
		return
	default:
		fmt.Printf("Unhandled signal caught: %s", signal)
	}
}
