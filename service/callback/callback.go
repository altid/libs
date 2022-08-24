package callback

type Connecter interface {
	Connect(*Client) error
}

type Controller interface {
	Control() error
}

type Callback interface {
	Connecter
	Controller
}

// A client is returned on Client connection
type Client struct {
	Username string
}