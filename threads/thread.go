//go:build !plan9
// +build !plan9

package threads

func Start(func runmain(), fg bool){
	return
}
