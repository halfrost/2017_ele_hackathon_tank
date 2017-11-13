package nex

import (
	"errors"
	"sync"

	"github.com/damnever/cc"
	huskarpool "github.com/eleme/huskar-pool"
	"github.com/eleme/huskar/config"
	"github.com/eleme/huskar/service"
	"github.com/eleme/huskar/toggle"
	"github.com/eleme/nex/client/kafka"
	"github.com/eleme/nex/client/redis"
	"github.com/eleme/nex/db"
	"github.com/eleme/nex/tracking/etrace"
	"github.com/eleme/samaritan/sdk"
)

var (
	// ErrAlreadyInited is returned if nex resources isn't initialized.
	ErrAlreadyInited = errors.New("nex already inited")
	// ErrNexConfigNotLoad is returned if nex config isn't loaded.
	ErrNexConfigNotLoad = errors.New("nex config not load")
	// ErrHuskarConfigerNotInit is returned if huskar configer isn't initialized.
	ErrHuskarConfigerNotInit = errors.New("huskar configer not init")
	// ErrHuskarTogglerNotInit is returned if huskar toggler isn't initialized.
	ErrHuskarTogglerNotInit = errors.New("huskar toggler not init")
	// ErrHuskarServiceNotInit is returned if huskar service isn't initialized.
	ErrHuskarServiceNotInit = errors.New("huskar service not init")
	// ErrHuskarPoolNotInit is returned if huskar pool isn't initialized.
	ErrHuskarPoolNotInit = errors.New("huskar pool not init")
	// ErrETraceNotInit is returned if etrace isn't initialized.
	ErrETraceNotInit = errors.New("etrace not init")
	// ErrRedisNotInit is returned if redis pool isn't initialized.
	ErrRedisNotInit = errors.New("redis not init")
	// ErrDBNotInit is returned if database connector isn't initialized.
	ErrDBNotInit = errors.New("database not init")
	// ErrSamClientNotInit is returned if samaritan client isn't initialized.
	ErrSamClientNotInit = errors.New("samaritan client not init")
	// ErrKafkaClientNotInit is returned if kafka client isn't initialized.
	ErrKafkaClientNotInit = errors.New("kafka client not init")
)

var hub = &struct {
	*sync.RWMutex
	Inited            bool
	NexConfig         cc.Configer
	HuskarConfiger    config.Configer
	HuskarToggler     toggle.Toggler
	HuskarRegistrator service.Registrator
	HuskarPool        *huskarpool.Huskar
	ETrace            *etrace.Trace
	RedisPools        *redis.PoolManager
	DBManager         *db.Manager
	SamClient         sdk.Client
	KafkaClient       kafka.Client
}{
	RWMutex:           &sync.RWMutex{},
	Inited:            false,
	NexConfig:         nil,
	HuskarConfiger:    nil,
	HuskarToggler:     nil,
	HuskarRegistrator: nil,
	HuskarPool:        nil,
	ETrace:            nil,
	RedisPools:        nil,
	DBManager:         nil,
	SamClient:         nil,
	KafkaClient:       nil,
}

// GetNexConfig returns the initialized nex config, otherwise panic.
func GetNexConfig() cc.Configer {
	hub.RLock()
	defer hub.RUnlock()
	if hub.NexConfig == nil {
		panic(ErrNexConfigNotLoad)
	}
	return hub.NexConfig
}

// GetHuskarConfiger returns the initialized huskar configer, otherwise panic.
func GetHuskarConfiger() config.Configer {
	hub.RLock()
	defer hub.RUnlock()
	if hub.HuskarConfiger == nil {
		panic(ErrHuskarConfigerNotInit)
	}
	return hub.HuskarConfiger
}

// GetHuskarToggler returns the initialized huskar toggler, otherwise panic.
func GetHuskarToggler() toggle.Toggler {
	hub.RLock()
	defer hub.RUnlock()
	if hub.HuskarToggler == nil {
		panic(ErrHuskarTogglerNotInit)
	}
	return hub.HuskarToggler
}

// GetHuskarRegistrator returns the initialized huskar registrator, otherwise panic.
func GetHuskarRegistrator() service.Registrator {
	hub.RLock()
	defer hub.RUnlock()
	if hub.HuskarRegistrator == nil {
		panic(ErrHuskarServiceNotInit)
	}
	return hub.HuskarRegistrator

}

// GetHuskarPool returns the initialized huskar pool, otherwise panic, the Huskar pool creates pools,
// not a normal pool.
func GetHuskarPool() *huskarpool.Huskar {
	hub.RLock()
	defer hub.RUnlock()
	if hub.HuskarPool == nil {
		panic(ErrHuskarPoolNotInit)
	}
	return hub.HuskarPool
}

// GetRedisPools returns the initialized redis pool manager, otherwise panic.
func GetRedisPools() *redis.PoolManager {
	hub.RLock()
	defer hub.RUnlock()
	if hub.RedisPools == nil {
		panic(ErrRedisNotInit)
	}
	return hub.RedisPools
}

// GetETrace returns the initialized etrace client, otherwise panic.
func GetETrace() *etrace.Trace {
	hub.RLock()
	defer hub.RUnlock()
	if hub.ETrace == nil {
		panic(ErrETraceNotInit)
	}
	return hub.ETrace
}

// GetDBManager returns the initialized database manager, otherwise panic.
func GetDBManager() *db.Manager {
	hub.RLock()
	defer hub.RUnlock()
	if hub.DBManager == nil {
		panic(ErrDBNotInit)
	}
	return hub.DBManager
}

// GetSamClient returns the initialized samaritan client, otherwise panic.
func GetSamClient() sdk.Client {
	hub.RLock()
	defer hub.RUnlock()
	if hub.SamClient == nil {
		panic(ErrSamClientNotInit)
	}
	return hub.SamClient
}

// GetKafkaClient returns the initialized kafka client,otherwise panic.
func GetKafkaClient() kafka.Client {
	hub.RLock()
	defer hub.RUnlock()
	if hub.KafkaClient == nil {
		panic(ErrKafkaClientNotInit)
	}
	return hub.KafkaClient
}
