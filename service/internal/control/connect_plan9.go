package control

import (
	"os"
)

func ConnectService(name string) (*Control, error) {
	fd, err := os.OpenFile("/mnt/alt/clone", os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	// This creates /srv/$name, and returns our ctl file handle
	if _, err := fd.Write([]byte(name)); err != nil {
		return nil, err
	}

	return &Control{
		ctl: fd,
	}, nil
}
