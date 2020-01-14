package main

import (
	"context"
	"log"
	"os"
	"path"
)

type input struct {
	service string
	buff    string
	data    string
}

var in chan interface{}

func init() {
	in = make(chan interface{})
	s := &fileHandler{
		fn: newInput,
		ch: in,
	}
	addFileHandler("input", s)
}

// From the Styx server we get input data, send it from there to the underlying input files here
func listenInputs(ctx context.Context, cfg *config) (chan interface{}, error) {
	return in, nil
}

func newInput(msg *message) interface{} {
	return &input{
		service: msg.service,
		data:    msg.data,
		buff:    msg.buff,
	}
}

// Just append the message to the underlying file
func handleInput(msg interface{}) {
	input, ok := msg.(*input)
	if !ok {
		return
	}

	file := path.Join(*inpath, input.service, input.buff, "input")

	fp, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer fp.Close()
	fp.WriteString(input.data)
}
