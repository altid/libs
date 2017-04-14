package cleanmark

import (
	"bytes"
	"strings"
)

// CleanBytes - Escape all elements in byte array so that they will be rendered properly in markdown
func CleanBytes(b []byte) []byte {
	// Quick escape if no work needed
	if !bytes.ContainsAny(b, "*#_/\\()~`[]!?") {
		return b
	}

	var offset int
	// We can never be more than 2x our initial size
	cap := len(b) * 2
	result := make([]byte, cap)

	for i, c := range b {
		switch c {
		case '*', '#', '_', '\\', '/', '(', ')', '`', '~', '[', ']', '!', '?':
			result[i+offset] = '\\'
			offset++
		}
		result[i+offset] = c
	}
	return result[:len(b)+offset]
}

// CleanString - Escape all elements in string so they will be rendered properly in markdown
func CleanString(s string) string {
	// Quick escape if no work needed
	if !strings.ContainsAny(s, "*#_/\\()~`[]!?") {
		return s
	}

	var offset int
	// We can never be more than 2x our initial size
	cap := len(s) * 2
	result := make([]rune, cap)

	for i, c := range s {
		switch c {
		case '*', '#', '_', '\\', '/', '(', ')', '`', '~', '[', ']', '!', '?':
			result[i+offset] = '\\'
			offset++
		}
		result[i+offset] = c
	}
	return string(result[:len(s)+offset])
}
