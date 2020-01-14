package main

import "context"

type filetype int

const (
	pageTitle filetype = iota
	pageStatus
	pageAside
	pageNavi
	pageBody
	pageTabs
	pageClosed
)

func listenClients(ctx context.Context) (chan interface{}, error) {
	cl := make(chan interface{})
	return cl, nil
}

func handleClient(msg interface{}) {
	// Add client to list
}
