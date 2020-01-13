package main

type server struct {
	msg message
}

type message struct {
	service string
	data    string
	buff    string
}

type fileHandler struct {
	fn func(msg *message) interface{}
	ch chan interface{}
}

var handlers map[string]*fileHandler

func init() {
	handlers = make(map[string]*fileHandler)
}

func addFileHandler(path string, fh *fileHandler) {
	handlers[path] = fh
}
