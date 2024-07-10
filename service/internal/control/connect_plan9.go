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
	if _, e := ctl.Read(b); e != nil {
		return nil, e
	}

	path :=	fmt.Sprintf("/mnt/alt/%s", b)
	fd, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	// This creates /srv/$name, and returns our ctl file handle
	fmt.Fprint(fd, name)
	if _, err := fmt.Fprint(fd, name); err != nil {
		return nil, err
	}

	return &Control{
		ctx: ctx,
		ctl: fd,
	}, nil
}
