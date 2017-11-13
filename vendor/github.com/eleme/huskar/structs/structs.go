// Copyright 2016 Eleme Inc. All rights reserved.

package structs

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"path"
	"strconv"
	"time"
)

var (
	// ErrKeyNotExists is returned when the key not exists.
	ErrKeyNotExists = errors.New("key not exists")
	// ErrKeyEmpty is returned when the key is empty.
	ErrKeyEmpty = errors.New("key can not be empty")
	// ErrWaitTimeout is returned when wait timeout.
	ErrWaitTimeout = errors.New("wait timeout")
)

// MessageType represents the message type.
type MessageType string

const (
	// MessageTypePing indicates that the ping message.
	MessageTypePing = MessageType("ping")
	// MessageTypeAll indicates that the all message.
	MessageTypeAll = MessageType("all")
	// MessageTypeUpdate indicates that the update message.
	MessageTypeUpdate = MessageType("update")
	// MessageTypeDelete indicates that the delete message.
	MessageTypeDelete = MessageType("delete")

	// MIN is the comparison of precision.
	MIN = 0.00001

	// StatusSuccess represents ths status of success from API response.
	StatusSuccess = "SUCCESS"
	// StatusApplicationExistedError represents the status of application existed.
	StatusApplicationExistedError = "ApplicationExistedError"
)

// EventType represents the event type.
type EventType int

const (
	// EventUnknown represents unkonwn event.
	EventUnknown = iota
	// EventAll represents all event.
	EventAll
	// EventUpdate represents update event.
	EventUpdate
	// EventDelete represents delete event.
	EventDelete
	// EventError represents error event.
	EventError
)

var (
	events = [5]string{"unknown", "all", "update", "delete", "error"}
)

// String return a string described event type.
func (et EventType) String() string {
	return events[et]
}

// Event represents the event of configuration.
type Event struct {
	Key   string
	Value string
	node  *Node
	Type  EventType
	Err   error
}

// NewEvent create an event with node.
func NewEvent(node *Node) *Event {
	if node == nil {
		return nil
	}
	return &Event{
		Key:   node.Key,
		Value: node.Value,
		node:  node,
		Type:  node.EventType,
	}
}

// IsOn return the state of key. True is ON, false is OFF. Only for toggle.
func (e *Event) IsOn() (bool, error) {
	return e.node.IsOn()
}

// Rate return the rate of key. Only for toggle.
func (e *Event) Rate() (float32, error) {
	return e.node.Rate()
}

// Config describes the configuration for interaction with Huskar API.
type Config struct {
	// Endpoint is a Huskar API URL prefix.
	Endpoint string
	// Token is the authorized token string.
	Token string
	// Service is the name of your service(e.g. zeus.eos).
	Service string
	// Cluster is the cluster name of your servers.
	Cluster string
	// WaitTimeout is the timeout of first connect to Huskar API. Default: 5 seconds.
	WaitTimeout time.Duration
	// DialTimeout is the timeout of dialing to Huskar API. Default: 1 second.
	DialTimeout time.Duration
	// RetryDelay is retry delay to connect to Huskar API. Default: 1 second.
	RetryDelay time.Duration
	// SOAMode is the 'soa_mode' in /etc/eleme/env.yaml.
	SOAMode string
}

// WatchType represents the watch type.
type WatchType int

const (
	// WatchTypeConfig indicates that the config type.
	WatchTypeConfig = iota
	// WatchTypeToggle indicates that the switch type.
	WatchTypeToggle
	// WatchTypeService indicates that the service type.
	WatchTypeService
)

var (
	watchTypes = [3]string{"config", "switch", "service"}
)

func (wt WatchType) String() string {
	return watchTypes[wt]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Entry represents the returned stream messages called by Long Poll API.
//
/* Such as:
```
{
	"body": {
		"switch": {},
		"config": {},
		"service": {
			"test_service": {
				"test_cluster": {
					"test_key": {
						"runtime": "{\"state\": \"up\"}",
						"value": "{\"ip\": \"8.8.8.8\", \"port\": {\"main\": 88}}"
					},
					"test_key2": {
						"runtime": "{}",
						"value": "{\"ip\":\"8.8.8.8\",\"port\":{\"main\":89},\"state\":\"up\",\"meta\":{},\"name\":\"test_service\"}"
					}
				}
			}
		}
	},
	"message": "all"
}
```
*/
type Entry struct {
	WatchType WatchType `json:"-"`

	Body struct {
		Config  ServiceMap `json:"config,omitempty"`
		Switch  ServiceMap `json:"switch,omitempty"`
		Service ServiceMap `json:"service,omitempty"`
	} `json:"body"`

	Message MessageType `json:"message"`
}

// ServiceMap represents a map of service to cluster map.
type ServiceMap map[string]ClusterMap

// ClusterMap represents a map of cluster to node map.
type ClusterMap map[string]NodeMap

// NodeMap represents a map of node name to node.
type NodeMap map[string]*Node

// Node represents the single node in Entry.
type Node struct {
	Key     string `json:"key,omitempty"`
	Value   string `json:"value,omitempty"`
	Runtime string `json:"runtime,omitempty"`

	Cluster   string    `json:"cluster,omitempty"`
	Service   string    `json:"-"`
	EventType EventType `json:"-"`
}

// IsEqual compare to another Node.
func (node *Node) IsEqual(another *Node) bool {
	if node.Key != another.Key {
		return false
	}
	if node.Value != another.Value {
		return false
	}
	if node.Runtime != another.Runtime {
		return false
	}
	return true
}

// IsOn return the state of node by key name. True is ON, false is OFF. Only for toggle.
func (node *Node) IsOn() (bool, error) {
	v, err := node.Rate()
	if err != nil {
		return false, err
	}
	if IsEqual(100.0, float64(v)) {
		return true, nil
	} else if IsEqual(float64(v), 0.0) {
		return false, nil
	} else {
		return float32(rand.Int31n(10000))/100.0 < v, nil
	}
}

// IsEqual compare two float numbers are equal to each other.
func IsEqual(f1, f2 float64) bool {
	return math.Dim(f1, f2) < MIN
}

// Rate return the rate of node by key name. Only for toggle.
func (node *Node) Rate() (float32, error) {
	v, err := strconv.ParseFloat(node.Value, 32)
	return float32(v), err
}

// SetEventType set the event type.
func (node *Node) SetEventType(et EventType) {
	node.EventType = et
}

// GetServiceMap return ServiceMap from indicated watch type.
func (e Entry) GetServiceMap() ServiceMap {
	switch e.WatchType {
	case WatchTypeConfig:
		return e.Body.Config
	case WatchTypeToggle:
		return e.Body.Switch
	case WatchTypeService:
		return e.Body.Service
	}
	return nil
}

// NodesMap return a nodes map from entry. Every node with the service and cluster name.
func (e Entry) NodesMap() map[string]*Node {
	nodesMap := make(map[string]*Node)
	if serviceMap := e.GetServiceMap(); serviceMap != nil {
		for service := range serviceMap {
			for cluster, nodeMap := range serviceMap[service] {
				for key, node := range nodeMap {
					node.Service = service
					node.Cluster = cluster
					node.Key = key
					nodesMap[path.Join(service, cluster, key)] = node
				}
			}
		}
	}
	return nodesMap
}

// Result is the base result of API response.
type Result struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// ResultString represents the result with string.
type ResultString struct {
	Result
	Data string `json:"data"`
}

// ResultNode represents the result with node info.
type ResultNode struct {
	Result
	Data *Node `json:"data"`
}

// ResultNodes contains multi nodes info.
type ResultNodes struct {
	Result
	Data []*Node `json:"data"`
}

// Cluster represends a service cluster
type Cluster struct {
	Name string `json:"name"`
}

// ResultClusters contains multi clusters in API result.
type ResultClusters struct {
	Data    []*Cluster `json:"data"`
	Message string     `json:"message"`
	Status  string     `json:"status"`
}

// NewResultNode creates ResultNode with content bytes.
func NewResultNode(content []byte) (*ResultNode, error) {
	ret := &ResultNode{}
	if err := json.Unmarshal(content, ret); err != nil {
		return nil, err
	}
	if ret.Status != StatusSuccess {
		return nil, fmt.Errorf("status: %s, message: %s", ret.Status, ret.Message)
	}
	return ret, nil
}

// NewResultNodes creates ResultNodes with content bytes.
func NewResultNodes(content []byte) (*ResultNodes, error) {
	ret := &ResultNodes{}
	if err := json.Unmarshal(content, ret); err != nil {
		return nil, err
	}
	if ret.Status != StatusSuccess {
		return nil, fmt.Errorf("status: %s, message: %s", ret.Status, ret.Message)
	}
	return ret, nil
}

// NewResultClusters creates ResultClusters with content bytes.
func NewResultClusters(content []byte) (*ResultClusters, error) {
	ret := new(ResultClusters)
	if err := json.Unmarshal(content, ret); err != nil {
		return nil, err
	}
	if ret.Status != StatusSuccess {
		return nil, fmt.Errorf("status: %s, message: %s", ret.Status, ret.Message)
	}
	return ret, nil
}

// NewResultString creates ResultString with content bytes.
func NewResultString(content []byte) (*ResultString, error) {
	ret := new(ResultString)
	if err := json.Unmarshal(content, ret); err != nil {
		return nil, err
	}
	if ret.Status != StatusSuccess {
		return nil, fmt.Errorf("status: %s, message: %s", ret.Status, ret.Message)
	}
	return ret, nil
}
