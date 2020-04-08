package settings

import "crypto/tls"

type Settings struct {
	Path     string
	Factotum bool
	Cert     tls.Certificate
}

// NewSettings returns a Settings for use in a Server's Config step
// Path is the path to Altid service
// If factotum is true, authentication will occur via a factotum
// https://9fans.github.io/plan9port/man/man4/factotum.html
func NewSettings(path string, factotum bool, cert tls.Certificate) *Settings {
	return &Settings{
		Path:     path,
		Factotum: factotum,
		Cert:     cert,
	}
}
