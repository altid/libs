package main

func init() {
	s := &fileHandler{
		fn: getDir,
	}
	addFileHandler("/", s)
}

func getDir(srv *service, msg *message) (interface{}, error) {
	//get current buffer and find all associated underlying files
	//Interface will be a list of fileInfo in this case.
	return nil, nil
}
