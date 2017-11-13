package url

import (
	"net"
	"strconv"
)

type url struct {
	host string
	port string
}

// SetHost sets inner host.
func (u *url) SetHost(host string) {
	u.host = host
}

// SetPort sets inner port.
func (u *url) SetPort(port int) {
	u.port = strconv.Itoa(port)
}

// Host returns inner host.
func (u *url) Host() string {
	return u.host
}

// Port returns inner port.
func (u *url) Port() string {
	return u.port
}

// HTTPAddress returns the HTTP address.
func (u *url) HTTPAddress() string {
	return "http://" + u.Address()
}

// Address returns the address as "host:port".
func (u *url) Address() string {
	return net.JoinHostPort(u.Host(), u.Port())
}
