package listener

type Listener interface {
	//Auth(*auth.Protocol) error
	Connect() error
	Control() error
	Listen() error
	List() ([]*File, error)
}

type File interface {
	Read(b []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
	Stream() (chan []byte, error)
}
