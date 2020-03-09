package client

import "io"

type feed struct {
	data chan []byte
	done chan struct{}
}

func (f *feed) Read(b []byte) (n int, err error) {
	select {
	case in := <-f.data:
		n = copy(b, in)
		return
	case <-f.done:
		return 0, io.EOF
	}
}

func (f *feed) Close() error {
	f.done <- struct{}{}
	return nil
}
