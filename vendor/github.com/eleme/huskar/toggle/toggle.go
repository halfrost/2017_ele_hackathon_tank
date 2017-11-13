// Copyright 2016 Eleme Inc. All rights reserved.

package toggle

import (
	"github.com/eleme/huskar/config"
	"github.com/eleme/huskar/internal"
	"github.com/eleme/huskar/structs"
)

// Toggler is usually used to enable or disable the API, it also has the ability to limit the passing rate.
type Toggler interface {
	// EnableCache enable disk cache.
	EnableCache(fpath string) error
	// IsOn get the current state of toggle by key name. True is ON, false is OFF.
	IsOn(key string) (bool, error)
	// IsOnOr get the current state of toggle by key name, and will return toggleDefault either when the desired toggle doesn't exists or there's an error getting it.
	IsOnOr(key string, toggleDefault bool) bool
	// Rate get the current rate of toggle by key name.
	Rate(key string) (float32, error)
	// Watch used to watch specified switch by key, and return the stop watch function.
	Watch(key string) (eventC <-chan *structs.Event, stopWatch func(), err error)
	// WatchAll the value change of all key, and return the stop watch function.
	WatchAll() (nodeC <-chan *structs.Event, stopWatch func(), err error)
	// GetAll return all key-value pair.
	GetAll() (map[string]string, error)
}

// Toggle implements the Toggler interface.
type Toggle struct {
	config.Configuration
}

// New create a Toggler with config.
func New(cfg structs.Config) (Toggler, error) {
	client, err := internal.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewWithClient(client), nil
}

// NewWithClient create a Toggle with client.
func NewWithClient(client *internal.Client) *Toggle {
	configuration := config.NewConfiguration(client, structs.WatchTypeToggle)
	return &Toggle{*configuration}
}

// IsOn get the current state of toggle by key name. True is ON, false is OFF.
func (tg *Toggle) IsOn(key string) (bool, error) {
	node, err := tg.GetNode(key)
	if err != nil {
		return false, err
	}
	return node.IsOn()
}

// IsOnOr get the current state of toggle by key name, and will return toggleDefault
// either if the desired toggle doesn't exists or there's an error getting it.
func (tg *Toggle) IsOnOr(key string, toggleDefault bool) bool {
	isOn, err := tg.IsOn(key)
	if err != nil {
		return toggleDefault
	}
	return isOn
}

// Rate get the current rate of toggle by key name.
func (tg *Toggle) Rate(key string) (float32, error) {
	node, err := tg.GetNode(key)
	if err != nil {
		return 0.0, err
	}
	return node.Rate()
}
