package main

// Every normal file
type normal struct {
}

func init() {
	s := &fileHandler{
		fn: getNormal,
	}
	addFileHandler("default", s)
}

// Heavy lifting here with fields function and join should be rewritten eventually
func getNormal(msg *message) (interface{}, error) {
	return &normal{}, nil
}
