package control

import (
	"context"
	"fmt"
	"os"
)

func ConnectService(ctx context.Context, name string) (*Control, error) {
	fd, err := os.OpenFile("/mnt/alt/clone", os.O_RDWR, 0644)
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
