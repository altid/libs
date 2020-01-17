package main

import (
	"log"
	"os"
	"path"
)

type input struct {
	service string
	buff    string
	data    string
}

func init() {
	s := &fileHandler{
		fn:   getInput,
		stat: getInputStat,
	}
	addFileHandler("/input", s)
}

func getInput(msg *message) (interface{}, error) {
	fp := path.Join(*inpath, msg.service, msg.buff, msg.file)
	return os.OpenFile(fp, os.O_RDWR, 0644)
}

func getInputStat(msg *message) (os.FileInfo, error) {
	fp := path.Join(*inpath, msg.service, msg.buff, msg.file)
	return os.Stat(fp)
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
