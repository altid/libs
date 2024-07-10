//go:build !plan9
// +build !plan9

package control

import (
	"context"
)

func ConnectService(ctx context.Context, name string) (*Control, error) {
	return nil, nil
}
