package redis

import (
	"time"

	"github.com/damnever/cc"
	"github.com/eleme/huskar-pool/pool"
	"github.com/eleme/nex/log"
	"github.com/garyburd/redigo/redis"
	json "github.com/json-iterator/go"
)

const (
	defaultConnectTimeout  int64 = 500 // ms
	defaultReadTimeout     int64 = 500 // ms
	defaultWriteTimeout    int64 = 500 // ms
	defaultPoolMaxActive   int   = 40
	defaultPoolIdleTimeout int64 = 60 // s
)

// PoolManager is used to manage redis connection pools.
type PoolManager struct {
	pools map[string]*Client
}

// NewPoolManager creates a new Redis connection pool manager, the jsonRedisSettings format:
//  {
//    "name": {
//      "url": "localhost:6479",
//      "connect_timeout": 500, // optional, default to 500 ms
//      "read_timeout": 500, // optional, default to 500 ms
//      "write_timeout": 500, // optional, default to 500 ms
//      "pool_max_active": 40, // optional, the pool size, default to 40
//      "pool_idle_timeout": 60, // optional, the connection idle timeout in pool, default to 60s
//      "enable_log": false, // optional, default to false
//	  }
//  }
func NewPoolManager(logger log.RPCContextLogger, jsonRedisSettings string) (*PoolManager, error) {
	var redisSettings map[string]*json.RawMessage
	err := json.Unmarshal([]byte(jsonRedisSettings), &redisSettings)
	if err != nil {
		return nil, err
	}

	pm := &PoolManager{pools: make(map[string]*Client, len(redisSettings))}
	for name, rawSettings := range redisSettings {
		redisConfig, err := cc.NewConfigFromJSON(*rawSettings)
		if err != nil {
			pm.CloseAll()
			return nil, err
		}
		if redisConfig.BoolOr("enable_log", false) {
			pm.pools[name] = NewPooledClient(logger, redisConfig)
		} else {
			pm.pools[name] = NewPooledClient(nil, redisConfig)
		}
	}

	return pm, nil
}

// GetPooledClient returns pool by name, nil returned if not found.
func (pm *PoolManager) GetPooledClient(name string) *Client {
	if pool, ok := pm.pools[name]; ok {
		return pool
	}
	return nil
}

// CloseAll closes all connection pools.
func (pm *PoolManager) CloseAll() {
	for _, pool := range pm.pools {
		pool.Close()
	}
}

// NewPooledClient creates a new redis client with pool support.
func NewPooledClient(logger log.RPCContextLogger, redisConfig cc.Configer) *Client {
	address := redisConfig.String("url")
	connectTimeout := redisConfig.DurationOr("connect_timeout", defaultConnectTimeout) * time.Millisecond
	readTimeout := redisConfig.DurationOr("read_timeout", defaultReadTimeout) * time.Millisecond
	writeTimeout := redisConfig.DurationOr("write_timeout", defaultWriteTimeout) * time.Millisecond
	options := []redis.DialOption{
		redis.DialConnectTimeout(connectTimeout),
		redis.DialReadTimeout(readTimeout),
		redis.DialWriteTimeout(writeTimeout),
	}
	factory := func() (pool.Resource, error) {
		conn, err := newRedisConn(address, options...)
		if err != nil {
			return nil, err
		}
		return &rawConn{conn: conn}, nil
	}

	maxActive := redisConfig.IntOr("pool_max_active", defaultPoolMaxActive)
	idleTimeout := redisConfig.DurationOr("pool_idle_timeout", defaultPoolIdleTimeout) * time.Second
	pool := pool.NewResourcePool(factory, maxActive, maxActive+10, idleTimeout)
	return newClient(address, pool, logger)
}
