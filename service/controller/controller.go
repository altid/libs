package controller

import "io"

type Controller interface {
	CreateBuffer(string) error
	DeleteBuffer(string) error
	Remove(string, string) error
	Notification(string, string, string) error
	WriteError(string) error
	WriteStatus(string, string) error
	WriteAside(string, string) error
	WriteNav(string, string) error
	WriteTitle(string, string) error
	WriteImage(string, string, io.ReadCloser) error
	WriteMain(string, io.ReadCloser) error
	WriteFeed(string, io.ReadCloser) error
	HasBuffer(string) bool
}

type WriteCloser interface {
	Write(b []byte) (int, error)
	Close() error
}
