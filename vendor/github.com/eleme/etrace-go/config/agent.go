package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ConfigAgent periodically fetch client config from etrace server.
type configAgent struct {
	endpoint string
	appid    string
	hostip   string
	client   *http.Client
}

// NewConfigAgent returns an etrace config agent.
func newConfigAgent(endpoint, appid, hostip string, timeout time.Duration) *configAgent {
	c := &configAgent{
		endpoint: endpoint,
		appid:    appid,
		hostip:   hostip,
	}
	c.client = &http.Client{
		Timeout: timeout,
	}
	return c
}

func (c *configAgent) fetchServer() ([]Collector, error) {
	url := fmt.Sprintf(`http://%s/collector?appId=%s&host=%s`, c.endpoint, c.appid, c.hostip)
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var collectors []Collector
	err = json.NewDecoder(resp.Body).Decode(&collectors)
	return collectors, err
}

func (c *configAgent) fetchConfig() (RemoteConfig, error) {
	config := defaultRemoteConfig()
	url := fmt.Sprintf(`http://%s/agent-config?appId=%s&host=%s`, c.endpoint, c.appid, c.hostip)
	resp, err := c.client.Get(url)
	if err != nil {
		return config, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&config)
	return config, err
}
