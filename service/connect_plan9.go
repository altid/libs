package service

import (
	"io"
	"os"
)

func connectService(name string) (io.ReadWriteCloser, error) {
	fd, err := os.OpenFile("/mnt/alt/clone", os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	// This creates /srv/$name, and returns our ctl file handle
	if _, err := fd.Write([]byte(name)); err != nil {
		return nil, err
	}
	return fd, nil
}
