// Copyright 2016 Eleme Inc. All rights reserved.

package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eleme/huskar/structs"
)

const (
	// EnvEndpoint is the environment name of huskar Endpoint.
	EnvEndpoint = "HUSKAR_ENDPOINT"
	// EnvToken is the environment name of huskar Token.
	EnvToken = "HUSKAR_TOKEN"

	// DefaultCluster indicates that the overall cluster name.
	DefaultCluster = "overall"

	// HeaderAuth is the authorization head in HTTP request.
	HeaderAuth = "Authorization"
)

var (
	// ErrInvalidEndpoint is returned when the endpoint is not a valid HTTP URL.
	ErrInvalidEndpoint = errors.New("huskar: invalid endpoint")
	// ErrInvalidToken is returned when the token is not a valid token string.
	ErrInvalidToken = errors.New("huskar: invalid token")
	// ErrConnectionRefused is returned when the HTTP Connection refuesed.
	ErrConnectionRefused = errors.New("huskar: connection refuesed")
	// ErrInvalidService is returned when the service is not a valid service name.
	ErrInvalidService = errors.New("huskar: invalid service")
	// ErrInvalidSOAMode is returned when the SOAMode is not valid.
	ErrInvalidSOAMode = errors.New("huskar: invalid soa mode")
)
var (
	soaModeHeader    = "X-SOA-Mode"
	soaClusterHeader = "X-Cluster-Name"
	soaModes         = map[string]bool{
		"":       true,
		"orig":   true,
		"prefix": true,
		"route":  true,
	}
)

// Client provides methods for interaction with the Huskar API.
type Client struct {
	config      structs.Config
	httpClient  *http.Client
	endpointURL *url.URL
}

// NewClient creates a new client from a given configuration.
func NewClient(config structs.Config) (*Client, error) {
	if endpoint := os.Getenv(EnvEndpoint); endpoint != "" {
		config.Endpoint = endpoint
	}
	if token := os.Getenv(EnvToken); token != "" {
		config.Token = token
	}

	u, err := parseEndpoint(config.Endpoint)
	if err != nil {
		return nil, err
	}
	if config.Token == "" {
		return nil, ErrInvalidToken
	}
	if config.Service == "" {
		return nil, ErrInvalidService
	}

	if config.Cluster == "" {
		config.Cluster = DefaultCluster
	}

	if config.WaitTimeout <= 0 {
		config.WaitTimeout = time.Second * 5
	}
	if config.DialTimeout <= 0 {
		config.DialTimeout = time.Second * 1
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = time.Second * 1
	}
	if !soaModes[config.SOAMode] {
		return nil, ErrInvalidSOAMode
	}

	dial := func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, config.DialTimeout)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: dial,
		},
	}

	c := &Client{
		config:      config,
		httpClient:  httpClient,
		endpointURL: u,
	}
	return c, nil
}

func parseEndpoint(endpoint string) (*url.URL, error) {
	if endpoint != "" && !strings.Contains(endpoint, "://") {
		endpoint = "http://" + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, ErrInvalidEndpoint
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		if e, ok := err.(*net.AddrError); ok {
			if e.Err == "missing port in address" {
				return u, nil
			}
		}
		return nil, ErrInvalidEndpoint
	}
	number, err := strconv.ParseInt(port, 10, 64)
	if err == nil && number > 0 && number < 65535 {
		return u, nil
	}
	return nil, ErrInvalidEndpoint
}

func (c *Client) getURL(path string) string {
	urlStr := strings.TrimRight(c.endpointURL.String(), "/")
	return fmt.Sprintf("%s%s", urlStr, path)
}

// Config return the config on client.
func (c *Client) Config() structs.Config {
	return c.config
}

// DoOptions represents the required options for short connection request.
type DoOptions struct {
	Headers  map[string]string
	FormData map[string]string
}

func (c *Client) addHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	if c.config.SOAMode != "" {
		req.Header.Add(soaModeHeader, c.config.SOAMode)
		req.Header.Add(soaClusterHeader, c.config.Cluster)
	}
}

// Do is used to do short connection request.
func (c *Client) Do(method, path string, doOptions DoOptions) (*http.Response, error) {
	var params bytes.Buffer
	w := multipart.NewWriter(&params)
	for k, v := range doOptions.FormData {
		if err := w.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	w.Close()

	u := c.getURL(path)
	req, err := http.NewRequest(method, u, &params)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	c.addHeaders(req, doOptions.Headers)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return nil, ErrConnectionRefused
		}
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return nil, structs.NewError(resp)
	}
	return resp, nil
}

// StreamOptions represents the required options for long connection request.
type StreamOptions struct {
	Headers map[string]string
	In      io.Reader
	Stdout  io.Writer
}

// Stream is used to do long connection request.
func (c *Client) Stream(method, path string, streamOptions StreamOptions) error {
	u := c.getURL(path)
	req, err := http.NewRequest(method, u, streamOptions.In)
	if err != nil {
		return err
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	c.addHeaders(req, streamOptions.Headers)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return ErrConnectionRefused
		}
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return structs.NewError(resp)
	}
	defer resp.Body.Close()
	_, err = io.Copy(streamOptions.Stdout, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
