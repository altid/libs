package html

import "fmt"

// URL is a url tag <a href="something">Msg</a>
type URL struct {
	Href []byte
	Msg  []byte
}

func (u *URL) String() string { return fmt.Sprintf(" * [%s](%s)\n", u.Msg, u.Href) }
