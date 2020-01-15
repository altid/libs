package main

import (
	"fmt"
	"log"
)

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

type client struct {
	target string
	uuid   int64
}

// Clients can be connected tospecial buffer "none"
// Client message means incoming new client
func handleClient(msg interface{}) {
	cl, ok := msg.(*client)
	if !ok {
		log.Fatal("Received non-client struct in handleClient")
	}
	fmt.Println(cl)
}
