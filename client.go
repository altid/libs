package main

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

func handleClient(msg interface{}) {
	// Add client to our list and update all requisite data sets
}
