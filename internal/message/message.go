package message

// Message contains information about which file is being requested
// The UUID should be set to the client which is requesting the file
type Message struct {
	Service string
	Buffer  string
	Target  string
	UUID    uint32
}
