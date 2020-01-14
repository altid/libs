package main

// feed files are special in that they're blocking
type feed struct{}

func init() {
	s := &fileHandler{
		fn: getFeed,
	}
	addFileHandler("/feed", s)
}

// Heavy lifting here with fields function and join should be rewritten eventually
func getFeed(msg *message) (interface{}, error) {
	return &feed{}, nil
}
