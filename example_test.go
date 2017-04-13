package cleanmark

import (
	"fmt"
)

func Example() {
	foo := []byte("*test* all _of_ /this\\")
	bar := CleanBytes(foo)
	fmt.Println(string(bar))
	baz := CleanString("*test* all of _this_ /too\\")
	fmt.Println(baz)
}
