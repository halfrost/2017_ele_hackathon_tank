package sdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/eleme/samaritan/sdk/url"
	"github.com/pkg/errors"
)

// version is SDK version
const version = "1.0.0"

// samClient is a client for Samaritan.
type samClient struct {
	*baseHTTPClient

	appID   string
	cluster string
	url     *url.Samaritan
}

// Assert samClient implements Client.
var _ = Client(&samClient{})

// NewClient creates a client with given appID, cluster and Samaritan's host and port.
func NewClient(appID, cluster, host string, port int) (Client, error) {
	appID, cluster = trimSpaces(appID, cluster)
	if len(appID) == 0 || len(cluster) == 0 {
		return nil, errors.New("both AppID and cluster must be specified")
	}

	c := &samClient{
		baseHTTPClient: newBaseClient(),
		appID:          appID,
		cluster:        cluster,
		url:            url.NewSamaritan(host, port),
	}
	c.setUserAgent(fmt.Sprintf(fmt.Sprintf("SamSDK/go/%s/%s/%s", version, appID, cluster)))
	return c, nil
}

// NewLocalClient creates a client with given appID and cluster, use local Samaritan.
func NewLocalClient(appID, cluster string) (Client, error) {
	return NewClient(appID, cluster, "127.0.0.1", 12345)
}

// DeclareUserApplication declares user application information to Samaritan.
func (c *samClient) DeclareUserApplication() error {
	if support, err := c.isAPISupported(apiApplication); err != nil || !support || c.isUserApplicationDeclared() {
		return err
	}

	if err := c.doSimple(http.MethodPost, c.url.DeclareUserApplication(c.appID, c.cluster)); err != nil {
		return errors.WithMessage(err, "fail to declare user application to Samaritan")
	}
	return nil
}

// RevokeUserApplicationDeclaration revokes the user application declaration from Samaritan.
func (c *samClient) RevokeUserApplicationDeclaration() error {
	if support, err := c.isAPISupported(apiApplication); err != nil || !support || !c.isUserApplicationDeclared() {
		return err
	}

	if err := c.doSimple(http.MethodDelete, c.url.DeclareUserApplication(c.appID, c.cluster)); err != nil {
		return errors.WithMessage(err, "fail to revoke user application declaration to Samaritan")
	}
	return nil
}

// GetHostPort returns the host and port of given appID and cluster.
// It will registers given appID and cluster if not registered.
func (c *samClient) GetHostPort(appID, cluster string) (string, int, error) {
	if err := c.RegisterDep(appID, cluster); err != nil {
		return "", 0, err
	}

	var host = c.url.Host()
	port, err := c.portOfAppID(appID)
	return host, port, err
}

// RegisterDep registers the given appID and cluster to Samaritan.
func (c *samClient) RegisterDep(appID, cluster string) error {
	appID, cluster = trimSpaces(appID, cluster)
	return c.RegisterDepTimeout(appID, cluster, 30*time.Second)
}

// RegisterDepTimeout registers the given appID and cluster to Samaritan, and wait until timeout or corresponding frontend to show up.
func (c *samClient) RegisterDepTimeout(appID, cluster string, timeout time.Duration) error {
	c.DeclareUserApplication()
	appID, cluster = trimSpaces(appID, cluster)
	if c.isDepRegistered(appID, cluster) {
		return nil
	}

	var waitSecond = int(timeout / time.Second)
	return c.doSimple(http.MethodPost, c.url.RegDep(appID, cluster, waitSecond > 0, waitSecond))
}

// DeregisterDep deregisters the given appID and cluster from Samaritan.
func (c *samClient) DeregisterDep(appID, cluster string) error {
	c.DeclareUserApplication()
	appID, cluster = trimSpaces(appID, cluster)
	if !c.isDepRegistered(appID, cluster) {
		return nil
	}

	return c.doSimple(http.MethodDelete, c.url.RegDep(appID, cluster, false, 0))
}

// isDepRegistered returns true if given appID and cluster already registered to Samaritan.
func (c *samClient) isDepRegistered(appID, cluster string) bool {
	deps, _ := c.getDep()
	dependentCluster, exist := deps[appID]
	return exist && dependentCluster == cluster
}

// isUserApplicationDeclared returns true if user application already declared to Samaritan.
func (c *samClient) isUserApplicationDeclared() bool {
	apps, _ := c.getApp()
	cluster, exist := apps[c.appID]
	return exist && cluster == c.cluster
}

// getDep returns all registered dependency.
func (c *samClient) getDep() (map[string]string, error) {
	u := c.url.GetRegistry()
	if supported, err := c.isAPISupported(apiApplication); err == nil && supported {
		u = c.url.GetDependency()
	}

	response, err := c.do(http.MethodGet, u)
	if err != nil {
		return nil, errors.WithMessage(err, "fail to get dependency")
	}

	defer response.Body.Close()
	rawDep, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "fail to read dependency")
	}

	var dep map[string]string
	if err := json.Unmarshal(rawDep, &dep); err != nil {
		return nil, errors.WithMessage(err, "invalid JSON")
	}
	return dep, nil
}

// getApp returns all declared user application.
func (c *samClient) getApp() (map[string]string, error) {
	u := c.url.GetApp("")
	response, err := c.do(http.MethodGet, u)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	rawApp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var apps map[string]string
	err = json.Unmarshal(rawApp, &apps)
	return apps, err
}

// portOfAppID gets given appID's port from Samaritan.
func (c *samClient) portOfAppID(appID string) (int, error) {
	response, err := c.do(http.MethodGet, c.url.GetServicePort(appID))
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	rawPort, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(rawPort))
}

// samStats returns the stats of Samaritan.
func (c *samClient) samStats() (*samStats, error) {
	response, err := c.do(http.MethodGet, c.url.GetStats())
	if err != nil {
		return nil, errors.WithMessage(err, "fail to get Samaritan stats")
	}
	defer response.Body.Close()

	rawStats, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "fail to get Samaritan stats")
	}

	var stats samStats
	err = json.Unmarshal(rawStats, &stats)
	return &stats, err
}

// samVersion returns the version of Samaritan.
func (c *samClient) samVersion() (samVersion, error) {
	stats, err := c.samStats()
	if err != nil {
		return nil, errors.WithMessage(err, "fail to get Samaritan version")
	}
	ver, _ := newVersionFromString(stats.Version)
	return ver, nil
}

// isAPISupported returns true, if Samaritan supported the given API.
func (c *samClient) isAPISupported(name api) (bool, error) {
	currentVersion, err := c.samVersion()
	if err != nil {
		return false, err
	}

	lowestVersion, found := lowestVersionOfAPI[name]
	if !found {
		return false, nil
	}
	return !currentVersion.Less(lowestVersion), nil
}
