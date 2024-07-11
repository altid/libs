package control

import (
	"context"
	"fmt"
	"os"
)

func ConnectService(ctx context.Context, name string) (*Control, error) {
	b := make([]byte, 0)
	ctl, err := os.Open("/mnt/alt/clone")
	if err != nil {
		return nil, err
	}
	
	n, err := ctl.Read(b)
	if err != nil {
		return nil, err
	}

	path :=	fmt.Sprintf("/mnt/alt/%c", b[:n-1])
	fd, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	// This creates /srv/$name, and returns our ctl file handle
	if _, err := fmt.Fprintf(fd, "%s\n", name); err != nil {
		return nil, err
	}

	return &Control{
		ctx: ctx,
		ctl: fd,
	}, nil
}
