package nex

import (
	"fmt"
	"strings"
	"time"

	"github.com/damnever/cc"
	"github.com/eleme/huskar/config"
	"github.com/eleme/huskar/structs"
	"github.com/eleme/nex/consts/huskarkeys"
)

const (
	elemeEnvFile  string = "/etc/eleme/env.yaml"
	elessEnvFile  string = "eless_env.yaml"
	appConfigFile string = "app.yaml"

	soaModeOrig    string = "orig"
	soaModePrefix  string = "prefix"
	soaModeRoute   string = "route"
	defaultSOAMode string = soaModeOrig

	soaIntentProxy  string = "proxy"
	soaIntentDirect string = "direct"
)

var defaultConfigs = map[string]interface{}{
	"version": fmt.Sprintf("%s", Version), // placeholder
}

func initNexConfig() (cc.Configer, error) {
	cfg, err := cc.NewConfigFromFile(appConfigFile)
	if err != nil {
		return nil, err
	}
	if !cfg.Bool("dev") {
		if err = cfg.MergeFromFile(elessEnvFile); err != nil {
			return nil, err
		}
		if err = cfg.MergeFromFile(elemeEnvFile); err != nil {
			return nil, err
		}
	}

	// For multiple datacenter
	var mdCluster string
	cluster := cfg.String("cluster")
	ezone := cfg.String("ezone")
	soamode := cfg.StringOr("soa_mode", defaultSOAMode)
	if (soamode == soaModeRoute || soamode == soaModePrefix) && ezone != "" { // prefix or route mode
		mdCluster = fmt.Sprintf("%s-%s", ezone, cluster)
	} else { // orig mode and xx do nothing
		mdCluster = cluster
	}
	cfg.Set("cluster", mdCluster)

	for k, v := range defaultConfigs {
		cfg.SetDefault(k, v)
	}
	return cfg, nil
}

func huskarConfigFromNex(nexCfg cc.Configer) structs.Config {
	huskarConfig := nexCfg.Config("huskar")
	dialTimeout := huskarConfig.DurationOr("dial_timeout", 5) * time.Second
	waitTimeout := huskarConfig.DurationOr("wait_timeout", 5) * time.Second
	retryDelay := huskarConfig.DurationOr("retry_delay", 2) * time.Second
	return structs.Config{
		Endpoint:    nexCfg.StringOr("huskar_api_url", huskarConfig.String("endpoint")),
		Token:       nexCfg.StringOr("huskar_api_token", huskarConfig.String("token")),
		Service:     nexCfg.String("app_name"),
		Cluster:     nexCfg.String("cluster"),
		SOAMode:     nexCfg.String("soa_mode"),
		DialTimeout: dialTimeout,
		WaitTimeout: waitTimeout,
		RetryDelay:  retryDelay,
	}
}

// BuildSOAClusterOrIntent build a cluster name by SOA mode.
func BuildSOAClusterOrIntent(name string, huskarConfiger config.Configer, nexCfg cc.Configer) string {
	soaMode := nexCfg.String("soa_mode")
	if soaMode == soaModeRoute {
		intent, err := huskarConfiger.Get(fmt.Sprintf(huskarkeys.RPCIntent, name))
		if err != nil {
			intent = soaIntentDirect
		}
		return intent
	}

	ezone := nexCfg.String("ezone")
	cluster, err := huskarConfiger.Get(fmt.Sprintf(huskarkeys.RPCCluster, name))
	if err != nil {
		cluster = nexCfg.String("cluster")
	}
	if soaMode == soaModePrefix && ezone != "" && !strings.HasPrefix(cluster, fmt.Sprintf("%s-", ezone)) {
		return fmt.Sprintf("%s-%s", ezone, cluster)
	}
	return cluster
}
