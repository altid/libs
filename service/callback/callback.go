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

// Sender interface is used by the listeners to handle control messages
// SendCommand can be intercepted, but finally should call your service.SendCommand from your controller with the payload
type Sender interface {
	SendCommand(string) error
}

// A client is returned on Client connection
type Client struct {
	Username string
}
