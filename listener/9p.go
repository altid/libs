package listener

import (
	"github.com/halfwit/styx"
)

type NineP struct {
	session *styx.Session
}

// Listen9p returns a server that listens over 9p for incoming clients
type Listen9p *NineP

// Listen9p returns a server that listens over 9p.2000 for incoming clients
type Listen9p2000 *NineP

func (np *NineP) Listen() error {
	return nil
}

func (np *NineP) Connect() error {
	return nil
}

func (np *NineP) List() ([]*File, error) {
	return nil, nil
}

func (np *NineP) Control() error {
	return nil
}

//func (np *NineP) Auth(*auth.Protocol) error

type p9File struct {

}

func (f *p9File) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (f *p9File) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (f *p9File) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f *p9File) Stream() (chan []byte, error) {
	return nil, nil
}