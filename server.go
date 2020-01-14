package main

import "context"

type server struct {
	ctx      context.Context
	cfg      *config
	events   chan *event
	inputs   chan interface{}
	clients  chan interface{}
	controls chan interface{}
	errors   []error
}

type message struct {
	service string
	data    string
	buff    string
}

type fileHandler struct {
	fn func(msg *message) (interface{}, error)
}

var handlers map[string]*fileHandler

func init() {
	handlers = make(map[string]*fileHandler)
}

func addFileHandler(path string, fh *fileHandler) {
	handlers[path] = fh
}

func newServer(ctx context.Context, cfg *config) (*server, error) {
	events, err := listenEvents(ctx, cfg)
	if err != nil {
		return nil, err
	}
	s := &server{
		events: events,
		inputs: make(chan interface{}),
		controls: make(chan interface{}),
		clients: make(chan interface{}),
		ctx: ctx,
		cfg: cfg,
	}
	return s, nil
}

func (s *server) listenAndServe() {
	// Loop through each service and listen. Use our fileHandlers
}
