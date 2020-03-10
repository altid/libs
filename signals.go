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
	case "terminated":
		cancel()
	case "urgent I/O condition":
		return
	default:
		fmt.Printf("Unhandled signal caught: %s", signal)
	}
}
