// Copyright 2016 Eleme Inc. All rights reserved.

package service

import (
	"encoding/json"
	"errors"
	"net"
	"strconv"

	"fmt"

	"github.com/eleme/huskar/structs"
)

var (
	// ErrMainPortRequired is returned when the main port not be indicated.
	ErrMainPortRequired = errors.New("main port must be specified")
)

// State decribes the state of service.
type State string

const (
	// StateUp indicates that the service state is UP.
	StateUp = State("up")
	// StateDown indicates that the service state is DOWN.
	StateDown = State("down")
)

// StaticInfo represents the static info on a service.
type StaticInfo struct {
	IP    string                 `json:"ip"`
	Port  map[string]int         `json:"port"`
	State State                  `json:"state,omitempty"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
	Name  string                 `json:"name,omitempty"`
}

// RuntimeInfo contain state only.
type RuntimeInfo struct {
	State State `json:"state"`
}

// Instance represents a service instance.
type Instance struct {
	Key     string       `json:"key"`
	Value   *StaticInfo  `json:"value"`
	Runtime *RuntimeInfo `json:"runtime,omitempty"`
}

// NewInstance create an instance with ip, port and state.
func NewInstance(ip string, port int, state State) *Instance {
	return &Instance{
		Key: fmt.Sprintf("%s_%d", ip, port),
		Value: &StaticInfo{
			IP:    ip,
			Port:  map[string]int{"main": port},
			State: state,
		},
		Runtime: &RuntimeInfo{State: state},
	}
}

func newInstanceFromNode(node *structs.Node) (*Instance, error) {
	instance := &Instance{}
	if node != nil {
		instance.Key = node.Key
		if node.Value != "" {
			staticInfo := &StaticInfo{}
			if err := json.Unmarshal([]byte(node.Value), staticInfo); err != nil {
				return nil, err
			}
			instance.Value = staticInfo
		}
		if node.Runtime != "" {
			runtimeInfo := &RuntimeInfo{}
			if err := json.Unmarshal([]byte(node.Runtime), runtimeInfo); err != nil {
				return nil, err
			}
			instance.Runtime = runtimeInfo
		}
	}
	return instance, nil
}

func (i *Instance) validate() error {
	if i.Key == "" {
		return structs.ErrKeyEmpty
	}

	if i.Value != nil {
		if _, err := net.ResolveIPAddr("ip4", i.Value.IP); err != nil {
			return err
		}
		if i.Value.Port == nil || len(i.Value.Port) == 0 {
			return ErrMainPortRequired
		}
		if mainPort, ok := i.Value.Port["main"]; !ok {
			return ErrMainPortRequired
		} else if _, err := net.LookupPort("tcp4", strconv.Itoa(mainPort)); err != nil {
			return err
		}
	}
	return nil
}

func (i *Instance) toMap() (map[string]string, error) {
	formData := map[string]string{
		"key": i.Key,
	}

	staticValue, err := json.Marshal(i.Value)
	if err != nil {
		return nil, err
	}
	formData["value"] = string(staticValue)

	if i.Runtime != nil {
		runtimeValue, err := json.Marshal(i.Runtime)
		if err != nil {
			return nil, err
		}
		formData["runtime"] = string(runtimeValue)
	}
	return formData, nil
}
