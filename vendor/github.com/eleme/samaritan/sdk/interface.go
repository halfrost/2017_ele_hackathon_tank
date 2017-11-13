package sdk

import "time"

// Client includes the fundamental of a Samaritan client.
type Client interface {
	GetHostPort(appID, cluster string) (host string, port int, err error)
	RegisterDep(appID, cluster string) error
	RegisterDepTimeout(appID, cluster string, timeout time.Duration) error
	DeregisterDep(appID, cluster string) error
	DeclareUserApplication() error
	RevokeUserApplicationDeclaration() error
}
