//go:build !plan9
// +build !plan9

package service

import (
	"io"
)

func connectService(name string) (io.ReadWriteCloser, error) {
	return nil, nil
}
