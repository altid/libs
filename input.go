package main

import "context"

type input struct {
	from string
	buff string
	data string
}

var in chan interface{}

func init() {
	// Register a handler for styx messages. We use input and ctl the same way
	in = make(chan interface{})
	s := &fileHandler{
		fn: newInput,
		ch: in,
	}
	addFileHandler("input", s)
}

// From the Styx server we get input data, send it from there to the underlying input files here
func listenInput(ctx context.Context, cfg *config) (chan interface{}, error) {
	return in, nil
}

func newInput(msg *message) interface{} {
	return &input{
		from: msg.service,
		data: msg.data,
		buff: msg.buff,
	}
}

func handleInput(in interface{}) {
	// Verify we have an *input
	// Check that in.from is a service that exists
	// Validate that in.buff has an input file present
	// write data
}
