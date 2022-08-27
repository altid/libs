package controller

import "github.com/altid/libs/service/input"

type Controller interface {
	Input(input.Handler, string, []byte) error
	CreateBuffer(string) error
	DeleteBuffer(string) error
	Remove(string, string) error
	Notification(string, string, string) error
	ErrorWriter() (WriteCloser, error)
	StatusWriter(string) (WriteCloser, error)
	SideWriter(string) (WriteCloser, error)
	NavWriter(string) (WriteCloser, error)
	TitleWriter(string) (WriteCloser, error)
	ImageWriter(string, string) (WriteCloser, error)
	MainWriter(string) (WriteCloser, error)
	HasBuffer(string) bool
}

type WriteCloser interface {
	Write(b []byte) (int, error)
	Close() error
}
