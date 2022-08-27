package runner

import "context"

type Starter interface {
	//Start(*control.Control) error
	Start() error
	StartContext(context.Context)
}

type Listener interface {
	//Listen(*control.Control)
	Listen()
	ListenContext(context.Context)
}

type Runner interface {
	Listener
	Starter
}
