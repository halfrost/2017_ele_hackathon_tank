// Copyright 2016 Eleme Inc. All rights reserved.

package internal

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/eleme/huskar/structs"
)

const (
	fakeToken = "test_token"
)

// FakeRoundTripper represents a fake Transport.
type FakeRoundTripper struct {
	Message  string
	Status   int
	Header   map[string]string
	Requests []*http.Request
}

// RoundTrip implements the RoundTripper interface.
func (rt *FakeRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	body := strings.NewReader(rt.Message)
	rt.Requests = append(rt.Requests, r)
	res := &http.Response{
		StatusCode: rt.Status,
		Body:       ioutil.NopCloser(body),
		Header:     make(http.Header),
	}
	for k, v := range rt.Header {
		res.Header.Set(k, v)
	}
	return res, nil
}

// Reset set the Requests to nil.
func (rt *FakeRoundTripper) Reset() {
	rt.Requests = nil
}

// TestConfig just for test.
var TestConfig = structs.Config{
	Endpoint: "http://localhost:8020",
	Token:    "test_token",
	Service:  "test_service",
	Cluster:  "test_cluster",
}

// TestConfigAlpha for alpha environment tests..
var TestConfigAlpha = structs.Config{
	Endpoint: "http://soa-zk.alpha.elenet.me:8020",
	Token:    "",
	Service:  "test_service",
	Cluster:  "test_cluster",
}

// NewTestClient return a client with config and fake roundTripper.
func NewTestClient(config structs.Config, rt *FakeRoundTripper) (*Client, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	if rt != nil {
		client.httpClient.Transport = rt
	}
	// fake a token if not in alpha env.
	if config.Endpoint != TestConfigAlpha.Endpoint {
		client.config.Token = fakeToken
	}
	return client, nil
}
