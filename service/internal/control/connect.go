//go:build !plan9
// +build !plan9

package control

func ConnectService(name string) (*Control, error) {
	return nil, nil
}
