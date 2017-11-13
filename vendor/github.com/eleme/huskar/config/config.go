// Copyright 2016 Eleme Inc. All rights reserved.

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eleme/huskar/cache"
	"github.com/eleme/huskar/internal"
	"github.com/eleme/huskar/structs"
)

// Configer used to retrieving configurations.
type Configer interface {
	// EnableCache enable disk cache.
	EnableCache(fpath string) error
	// Get config value by key.
	Get(key string) (string, error)
	// GetAll return all key-value pair.
	GetAll() (map[string]string, error)
	// UnmarshalAll is used to unmarshal all config elements.
	UnmarshalAll(i interface{}) error
	// Watch the value change of specified key, and return the stop watch function.
	Watch(key string) (nodeC <-chan *structs.Event, stopWatch func(), err error)
	// WatchAll the value change of all key, and return the stop watch function.
	WatchAll() (<-chan *structs.Event, func(), error)
	// Set used to add a new config.
	Set(key string, value []byte, comment []byte) error
	// Update used to update a config. If comment given nil, the comment of the key will not be updated.
	Update(key string, value []byte, comment []byte) error
	// Delete used to delete a config with key.
	Delete(key string) error
	// SetLogger sets the logger to be used for printing errors.
	SetLogger(l structs.Logger)
}

// Configuration implements the Configer interface.
type Configuration struct {
	logger   structs.Logger
	client   *internal.Client
	cluster  string
	watcher  *internal.Watcher
	registry *internal.Registry
	codec    Codecer
	services internal.Services

	rwmu  *sync.RWMutex                       // lock for store
	store map[string]map[string]*structs.Node // key -> cluster -> node
	cache *cache.Cache

	mu      *sync.Mutex // lock for started
	started bool        // flag for start loop

	waitOnce *sync.Once
	waitCh   chan struct{}
	done     chan struct{}
}

// New create a Configer with config.
func New(config structs.Config) (Configer, error) {
	client, err := internal.NewClient(config)
	if err != nil {
		return nil, err
	}
	return NewWithClient(client), nil
}

// NewWithClient create a Configer with client.
func NewWithClient(client *internal.Client) *Configuration {
	return NewConfiguration(client, structs.WatchTypeConfig)
}

// NewConfiguration create a Configuration with client and watch type.
func NewConfiguration(client *internal.Client, watchType structs.WatchType) *Configuration {
	if client == nil {
		panic("client cannot be nil")
	}
	cluster := client.Config().Cluster
	if cluster == "" {
		cluster = internal.DefaultCluster
	}
	return &Configuration{
		logger:   structs.DefaultLogger{},
		client:   client,
		cluster:  cluster,
		watcher:  internal.NewWatcher(client, watchType),
		registry: internal.NewRegistry(5),
		codec:    new(Codec),
		services: buildServices(client.Config()),
		rwmu:     new(sync.RWMutex),
		store:    make(map[string]map[string]*structs.Node),
		mu:       new(sync.Mutex),
		started:  false,
		waitOnce: new(sync.Once),
		waitCh:   make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func buildServices(config structs.Config) internal.Services {
	services := internal.NewServices()
	services.AddClusters(config.Service, internal.DefaultCluster)
	if config.Cluster != "" && config.Cluster != internal.DefaultCluster {
		services.AddClusters(config.Service, config.Cluster)
	}
	return services
}

// EnableCache enable disk cache, call it ASAP.
func (c *Configuration) EnableCache(cachePath string) error {
	c.rwmu.Lock()
	defer c.rwmu.Unlock()

	cache, err := cache.New(cachePath)
	if err != nil {
		return err
	}

	if len(c.store) > 0 {
		err = cache.Save(&c.store)
	} else {
		err = cache.Load(&c.store)
		if err != nil { // compatible with old cache format
			c.logger.Printf("load cache failed, try load by old format: %v", err)
			c.store = map[string]map[string]*structs.Node{}
			var oldStore map[string]*structs.Node
			if err = cache.Load(&oldStore); err == nil {
				c.initFromOldCacheFormat(oldStore)
				err = cache.Save(&c.store) // reformat
			} else {
				c.store = map[string]map[string]*structs.Node{}
			}
		}
	}
	if err != nil {
		return err
	}
	c.cache = cache
	return nil
}

func (c *Configuration) initFromOldCacheFormat(nodesMap map[string]*structs.Node) {
	for key, node := range nodesMap {
		if node.Cluster != internal.DefaultCluster && node.Cluster != c.cluster {
			continue
		}
		clusterNodes, ok := c.store[key]
		if !ok {
			clusterNodes = map[string]*structs.Node{}
			c.store[key] = clusterNodes
		}
		clusterNodes[node.Cluster] = node
	}
}

func (c *Configuration) tryStart() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		return nil
	}

	go c.start()
	c.started = true

	select {
	// wait for first time start succeed.
	case <-c.waitCh:
		break
	case <-time.After(c.client.Config().WaitTimeout): // wait timeout.
		return structs.ErrWaitTimeout
	}
	return nil
}

func (c *Configuration) start() {
	select {
	case <-c.done:
		return
	default:
	}

	entryC := make(chan *structs.Entry)
	done := make(chan bool)
	errCh := c.watcher.Watch(c.services, internal.WatchOptions{EntryC: entryC, Done: done})
Loop:
	for {
		select {
		case entry, ok := <-entryC:
			if !ok {
				break Loop
			}
			nodesMap := entry.NodesMap()
			// service link changed will not trigger all keys pull event.
			switch entry.Message {
			case structs.MessageTypeUpdate:
				c.pubNodes(nodesMap, structs.EventUpdate)
			case structs.MessageTypeDelete:
				c.pubNodes(nodesMap, structs.EventDelete)
			case structs.MessageTypeAll:
				c.waitOnce.Do(func() {
					close(c.waitCh)
				})
				c.pubNodes(nodesMap, structs.EventAll)
			}
		case <-c.done:
			break Loop
		}
	}
	err := <-errCh
	if err != nil {
		c.logger.Printf("%s, will retry after: %v", err, c.client.Config().RetryDelay)
		time.Sleep(c.client.Config().RetryDelay)
		go c.start()
	}
	close(done)
}

// only publish update or delete message.
func (c *Configuration) pubNodes(nodesMap map[string]*structs.Node, eventType structs.EventType) {
	changedNodes := []*structs.Node{}

	c.rwmu.Lock()
	switch eventType {
	case structs.EventAll:
		keyClusters := map[string]bool{}
		for _, node := range nodesMap {
			keyClusters[fmt.Sprintf("%v/%v", node.Key, node.Cluster)] = true
			if c.set(node) {
				node.SetEventType(structs.EventUpdate)
				changedNodes = append(changedNodes, node)
			}
		}
		for key, clusterMap := range c.store {
			for cluster, node := range clusterMap {
				if !keyClusters[fmt.Sprintf("%v/%v", key, cluster)] {
					if c.delete(node) {
						node.SetEventType(structs.EventDelete)
						changedNodes = append(changedNodes, node)
					}
				}
			}
		}
	case structs.EventUpdate:
		for _, node := range nodesMap {
			if c.set(node) {
				node.SetEventType(eventType)
				changedNodes = append(changedNodes, node)
			}
		}
	case structs.EventDelete:
		for _, node := range nodesMap {
			if c.delete(node) {
				node.SetEventType(eventType)
				changedNodes = append(changedNodes, node)
			}
		}
	}
	if c.cache != nil {
		c.cache.Save(&c.store)
	}
	c.rwmu.Unlock()

	for _, node := range changedNodes {
		c.registry.Pub(node, node.Key)
	}
}

// Stop stop the configer.
func (c *Configuration) Stop() {
	if c.done != nil {
		close(c.done) // stop loop and watch goroutine.
	}
	c.registry.Clean()
}

// SetLogger sets the logger to be used for printing errors.
func (c *Configuration) SetLogger(l structs.Logger) {
	c.logger = l
}

// Get config value by key. Return error if exists.
func (c *Configuration) Get(key string) (string, error) {
	node, err := c.GetNode(key)
	if err != nil {
		return "", err
	}
	return node.Value, nil
}

// GetNode return the Node by key. Return error if exists.
func (c *Configuration) GetNode(key string) (*structs.Node, error) {
	c.tryStart()

	c.rwmu.RLock()
	defer c.rwmu.RUnlock()
	node, ok := c.get(key)
	if !ok {
		return nil, structs.ErrKeyNotExists
	}
	return node, nil
}

// GetAll return all key-value pair. Return error if exists.
func (c *Configuration) GetAll() (map[string]string, error) {
	c.tryStart()

	c.rwmu.RLock()
	defer c.rwmu.RUnlock()
	ret := make(map[string]string, len(c.store))
	for k := range c.store {
		if node, ok := c.get(k); ok {
			ret[k] = node.Value
		}
	}
	return ret, nil
}

// UnmarshalAll is used to unmarshal all config elemenets
func (c *Configuration) UnmarshalAll(i interface{}) error {
	values, err := c.GetAll()
	if err != nil {
		return err
	}
	if c.codec == nil {
		return errors.New("codec is nil, can not unmarshal values")
	}
	return c.codec.Unmarshal(values, i)
}

func (c *Configuration) get(key string) (*structs.Node, bool) {
	clusterNodes, ok := c.store[key]
	if !ok {
		return nil, ok
	}
	node, ok := clusterNodes[c.cluster]
	if !ok {
		if c.cluster == internal.DefaultCluster {
			return nil, ok
		}
		node, ok = clusterNodes[internal.DefaultCluster]
	}
	return node, ok
}

func (c *Configuration) set(node *structs.Node) bool {
	if node.Cluster != internal.DefaultCluster && node.Cluster != c.cluster {
		return false
	}
	clusterNodes, ok := c.store[node.Key]
	if !ok {
		clusterNodes = map[string]*structs.Node{}
		c.store[node.Key] = clusterNodes
	}

	oldNode, ok := c.get(node.Key)
	if ok && node.Cluster == oldNode.Cluster && node.Value == oldNode.Value {
		return false
	}
	clusterNodes[node.Cluster] = node
	// cluster is 'overall'..
	if ok && (node.Cluster == internal.DefaultCluster && oldNode.Cluster != internal.DefaultCluster) {
		return false
	}
	return true
}

func (c *Configuration) delete(node *structs.Node) bool {
	clusterNodes, ok := c.store[node.Key]
	if !ok {
		return false
	}
	if _, ok = clusterNodes[node.Cluster]; !ok {
		return false
	}
	delete(clusterNodes, node.Cluster)
	if len(clusterNodes) == 0 {
		delete(c.store, node.Key)
	}
	// cluster is 'overall'..
	oldNode, ok := c.get(node.Key)
	if ok && (node.Cluster == internal.DefaultCluster && oldNode.Cluster != internal.DefaultCluster) {
		return false
	}
	return true
}

// Watch the value change of specified key, and return the stop watch function.
func (c *Configuration) Watch(key string) (<-chan *structs.Event, func(), error) {
	msgC := c.registry.Sub(key)

	if err := c.tryStart(); err != nil {
		c.registry.UnsubAll(msgC)
		return nil, nil, err
	}

	eventC := make(chan *structs.Event)
	go func() {
		for msg := range msgC {
			eventC <- structs.NewEvent(msg.(*structs.Node))
		}
		close(eventC)
	}()

	return eventC, func() {
		c.registry.Unsub(key, msgC)
	}, nil
}

// WatchAll the value change of all key, and return the stop watch function.
func (c *Configuration) WatchAll() (<-chan *structs.Event, func(), error) {
	msgC := c.registry.SubAll()

	if err := c.tryStart(); err != nil {
		c.registry.UnsubAll(msgC)
		return nil, nil, err
	}

	eventC := make(chan *structs.Event)
	go func() {
		for msg := range msgC {
			eventC <- structs.NewEvent(msg.(*structs.Node))
		}
		close(eventC)
	}()

	return eventC, func() {
		c.registry.UnsubAll(msgC)
	}, nil
}

// Set used to add a new config.
func (c *Configuration) Set(key string, value []byte, comment []byte) error {
	_, err := c.do(http.MethodPost, key, value, comment)
	return err
}

// Update used to update a config. If comment given nil, the comment of the key will not be modified.
func (c *Configuration) Update(key string, value []byte, comment []byte) error {
	_, err := c.do(http.MethodPut, key, value, comment)
	return err
}

// Delete used to delete a config with key.
func (c *Configuration) Delete(key string) error {
	_, err := c.do(http.MethodDelete, key, nil, nil)
	return err
}

func (c *Configuration) do(method string, key string, value, comment []byte) ([]byte, error) {
	if key == "" {
		return nil, structs.ErrKeyEmpty
	}
	config := c.client.Config()

	doOptions := internal.DoOptions{
		Headers:  map[string]string{internal.HeaderAuth: config.Token},
		FormData: map[string]string{"key": key},
	}
	if strings.ToUpper(method) != http.MethodDelete {
		if value != nil {
			doOptions.FormData["value"] = string(value)
		}
		if comment != nil {
			doOptions.FormData["comment"] = string(comment)
		}
	}

	resp, err := c.client.Do(method, fmt.Sprintf("/api/config/%s/%s", config.Service, config.Cluster), doOptions)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received unexpected status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
