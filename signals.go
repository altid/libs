package main

import (
	"context"
	"fmt"
)

func handleSig(cancel context.CancelFunc, signal string) {
	switch signal {
	case "interrupt":
		cancel()
	case "suspended":
		cancel()
	default:
		fmt.Printf("Unhandled signal caught: %s", signal)
	}
}
