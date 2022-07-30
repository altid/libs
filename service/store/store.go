package store

import "io"

// Lister returns an array of paths available on the storage
// These can be accessed with an Open command, given the same path
type Lister interface {
	List() ([]string)
}

// Opener returns a single File item from the storage by name
// If the file does not exist, it is created and returned
// Files returned by an Opener should be closed with Close() after 
type Opener interface {
	Open(string) (File, error)
}

// Deleter removes a single File item from the storage by name
// If the file does not exist, has active Streams, or has not been correctly closed, an error is returned
type Deleter interface {
	Delete(string) error
}

// Filer is an interface that is required for the Listeners to have access to data
type Filer interface {
	Lister
	Opener
	Deleter
}

// File is an interface which represents data for a single file
type File interface {
	// Read reads up to len(b) bytes from the File and stores them in b. It returns the number of bytes read and any error encountered. At end of file, Read returns 0, io.EOF.
	Read(b []byte) (n int, err error)
	//Write writes len(b) bytes from b to the File. It returns the number of bytes written and an error, if any. Write returns a non-nil error when n != len(b).
	Write(p []byte) (n int, err error)
	// Seek sets the offset for the next Read or Write on file to offset, interpreted according to whence: 0 means relative to the origin of the file, 1 means relative to the current offset, and 2 means relative to the end. It returns the new offset and an error, if any.
	Seek(offset int64, whence int) (int64, error)
	// Close closes the File, rendering it unusable for I/O. On files that support SetDeadline, any pending I/O operations will be canceled and return immediately with an ErrClosed error. Close will return an error if it has already been called.
	Close() error
	// Stream ReadCloser that can be used to read bytes in a continuous manner
	// Each call to stream will get a copy of what has been written to the file
	// All further reads will block until there is new data, or Close() is called
	Stream() (io.ReadCloser, error)
	// Path returns the internal pathname of the File
	Path() string
}
