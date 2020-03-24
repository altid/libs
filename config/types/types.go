package types

// Auth matches an auth= tuple in a config
// If the value matches factotum, it will use the factotum to return a password
// If the value matches password, it will return the value of a password= tuple
// If the value matches none, it will return an empty string
type Auth string

// Logdir is the directory that an Altid service can optionally store logs to
// If this is unset in the config, it will be filled with "none"
type Logdir string

// ListenAddress is the listen_address tuple in a config
// If this is unset in the config, it will be filled with "localhost"
type ListenAddress string
