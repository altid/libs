# server
Altid server library and 9p client

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/altid/server) ![Tests](https://github.com/altid/server/workflows/Tests/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/altid/server)](https://goreportcard.com/report/github.com/altid/server) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org)

# 9pd

`go get github.com/altid/server/cmd/9pd`

9pd will watch (by default, /tmp/altid) a path containing Altid services, and orchestrate access to them for a 9p-based Altid client.
