package etrace

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	econfig "github.com/eleme/etrace-go/config"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// Config The container.
type Config econfig.Config

func withDefaultConfig(cfg Config) econfig.Config {
	return econfig.WithDefaultConfig(econfig.Config(cfg))

}

// CreateRequestID create request id based on appid.
func CreateRequestID(appid string) string {
	randValue := rand.Int63()
	nsecs := time.Now().UnixNano() / 1e6
	return fmt.Sprintf("%s^^%d|%d", appid, randValue, nsecs)
}

// CreateRPCID create rpc id based on appid and upstream rpc id.
func CreateRPCID(appid string, parent string) (rpcid string, clientAppID string) {
	if parent == "" {
		parent = "1"
	}
	parts := strings.Split(parent, "|")
	if len(parts) < 2 {
		rpcid = appid + "|" + parent
		clientAppID = "unknown"
	} else {
		rpcid = appid + "|" + parts[len(parts)-1]
		clientAppID = parts[0]
	}
	return
}
