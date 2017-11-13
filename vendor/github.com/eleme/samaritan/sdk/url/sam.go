package url

import "fmt"

// Samaritan generates URLs about Samaritan APIs.
type Samaritan struct {
	url
}

// Ping returns the URL used to ping Samaritan.
func (s *Samaritan) Ping() string {
	return s.HTTPAddress()
}

// GetStats returns the URL used to get Samaritan's stats.
func (s *Samaritan) GetStats() string {
	return join(s.HTTPAddress(), "stats")
}

// GetServicePort returns the URL used to get service's port.
func (s *Samaritan) GetServicePort(appID string) string {
	return join(s.HTTPAddress(), "ports", appID)
}

// GetRegistry returns the URL used to get all registry information.
// Note: better use GetDependency.
func (s *Samaritan) GetRegistry() string {
	return join(s.HTTPAddress(), "registry")
}

// GetDependency returns the URL used to get all dependency information.
func (s *Samaritan) GetDependency() string {
	return join(s.HTTPAddress(), "dependency")
}

// RegDep returns the URL used to register an appID with cluster to Samaritan.
// The third Boolean indicates wait until corresponding frontend to show up.
// Timeout unit is second.
func (s *Samaritan) RegDep(appID, cluster string, wait bool, timeout int) string {
	url := join(s.GetRegistry(), appID)

	if cluster != "" {
		url = join(url, cluster)
	}

	if wait {
		url += "?wait=1"
		if timeout > 0 {
			url += fmt.Sprintf("&timeout=%d", timeout)
		}
	}

	return url
}

// GetApp returns the URL that used to get Samaritan's user.
func (s *Samaritan) GetApp(appID string) string {
	url := join(s.HTTPAddress(), "application")
	if len(appID) > 0 {
		url = join(url, appID)
	}
	return url
}

// DeclareUserApplication returns the URL that used to register self to Samaritan.
func (s *Samaritan) DeclareUserApplication(appID, cluster string) string {
	return join(s.GetApp(appID), cluster)
}

// NewSamaritan creates a URL generator for Samaritan APIs.
func NewSamaritan(host string, port int) *Samaritan {
	sam := new(Samaritan)
	sam.SetHost(host)
	sam.SetPort(port)
	return sam
}
