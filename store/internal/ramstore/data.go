package ramstore

import (
	"sync"
)

// Internal data store
type data struct {
	bytes []byte
	sync.RWMutex
}

// The following three functions are adapted from the bytes.Buffer package
func growSlice(b []byte, n int) []byte {
	c := len(b) + n // ensure enough space for n elements
	if c < 2*cap(b) {
		c = 2 * cap(b)
	}
	// Double our buffer
	b2 := append([]byte(nil), make([]byte, c)...)
	copy(b2, b)
	return b2[:len(b)]
}

func (s *data) tryGrowByReslice(n int) (int, bool) {
	if l := len(s.bytes); n <= cap(s.bytes)-l {
		s.bytes = s.bytes[:l+n]
		return l, true
	}
	return 0, false
}

func (s *data) grow(off, n int) int {
	m := len(s.bytes)
	// If our buffer is empty, reset the slice
	if i, ok := s.tryGrowByReslice(n); ok {
		return i
	}
	if m == 0 {
		s.bytes = s.bytes[:0]
	}
	// Make a larger initial array so allocations don't race
	if s.bytes == nil && n <= 4096 {
		s.bytes = make([]byte, n, 4096)
		return 0
	}
	c := cap(s.bytes)
	if n <= c/2-m {
		copy(s.bytes, s.bytes[off:])
	} else {
		s.bytes = growSlice(s.bytes[off:], off+n)
	}
	s.bytes = s.bytes[:m+n]
	return m
}
