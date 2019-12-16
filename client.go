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
