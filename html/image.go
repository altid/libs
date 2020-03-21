package html

import "fmt"

// Image is an image tag <img src="foo" alt="bar">
type Image struct {
	Src []byte
	Alt []byte
}

func (i *Image) String() string {
	return fmt.Sprintf("![%s](%s)\n", i.Alt, i.Src)
}
