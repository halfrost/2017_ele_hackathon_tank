package config

import (
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

const (
	defaultAppID   = "arch.etracego"
	elemeEnvPath   = "/etc/eleme/env.yaml"
	alphaEtraceURL = "etrace-config.alpha.elenet.me:2890"
)

var (
	errAppIDEmpty     = errors.New("appid is empty")
	errConfigURLEmpty = errors.New("etrace config url empty")

	elemeConfig ElemeConfig
)

func init() {
	configYaml, err := ioutil.ReadFile(elemeEnvPath)
	if err == nil {
		yaml.Unmarshal(configYaml, &elemeConfig)
	}
	if elemeConfig.EtraceURI == "" {
		elemeConfig.EtraceURI = alphaEtraceURL
	}
	if elemeConfig.Cluster == "" {
		elemeConfig.Cluster = "unknown"
	}
	if elemeConfig.EZone == "" {
		elemeConfig.EZone = "unknown"
	}
	if elemeConfig.IDC == "" {
		elemeConfig.IDC = "unknown"
	}
}

// Logger is to log information.
type Logger interface {
	Printf(format string, v ...interface{})
}

// ElemeConfig is  default config  from environment.
type ElemeConfig struct {
	Cluster   string `yaml:"cluster"`
	EtraceURI string `yaml:"etrace_uri"`
	EZone     string `yaml:"ezone"`
	IDC       string `yaml:"idc"`
}

// Config The container.
type Config struct {
	AppID                string        `json:"appid"`
	HostIP               string        `json:"ip"`
	HostName             string        `json:"hostname"`
	Cluster              string        `json:"cluster"`
	EZone                string        `json:"ezone"`
	IDC                  string        `josn:"idc"`
	Topic                string        `json:"topic"`
	MesosTaskID          string        `json:"mesosTaskId"`
	EleapposLabel        string        `json:"eleapposLabel"`
	EleapposSlaveFqdn    string        `json:"eleapposSlaveFqdn"`
	EtraceHTTPTimeout    time.Duration `json:"etraceHTTPTimeout"`
	EtraceMaxCacheCount  int           `json:"etraceMaxCacheCount"`
	EtraceMaxCacheTime   time.Duration `json:"etraceMaxCacheTime"`
	EtraceConfigURL      string        `json:"etraceConfigURL"`
	EtraceConfigInterval time.Duration `json:"etraceConfigInterval"`
	EtraceListInterval   time.Duration `json:"etraceListInterval"`
	MessageMaxRegryNum   int           `json:"messageMaxRegryNum"`
	Logger               Logger        `json:"-"`
	Remoter              Remoter       `json:"-"`
}

// WithDefaultConfig uses default value for empty fields.
func WithDefaultConfig(cfg Config) Config {
	u := strings.TrimPrefix(cfg.EtraceConfigURL, "http://")
	cfg.EtraceConfigURL = strings.TrimSuffix(u, "/")
	if cfg.AppID == "" {
		cfg.AppID = defaultAppID
	}
	if cfg.EtraceConfigURL == "" {
		cfg.EtraceConfigURL = elemeConfig.EtraceURI
	}
	if cfg.HostIP == "" {
		cfg.HostIP, _ = getLocalIP()
	}
	if cfg.HostName == "" {
		cfg.HostName, _ = os.Hostname()
	}
	if cfg.Cluster == "" {
		cfg.Cluster = elemeConfig.Cluster
	}
	if cfg.EZone == "" {
		cfg.EZone = elemeConfig.EZone
	}
	if cfg.IDC == "" {
		cfg.IDC = elemeConfig.IDC
	}
	if cfg.EtraceHTTPTimeout <= 0 {
		cfg.EtraceHTTPTimeout = time.Duration(3 * time.Second)
	}
	if cfg.EtraceMaxCacheCount <= 0 {
		cfg.EtraceMaxCacheCount = 500
	}
	if cfg.EtraceMaxCacheTime <= 0 {
		cfg.EtraceMaxCacheTime = time.Duration(2 * time.Second)
	}
	if cfg.EtraceConfigInterval <= 0 {
		cfg.EtraceConfigInterval = time.Duration(time.Minute)
	}
	if cfg.EtraceListInterval <= 0 {
		cfg.EtraceListInterval = time.Duration(3 * time.Minute)
	}
	if cfg.MessageMaxRegryNum <= 0 {
		cfg.MessageMaxRegryNum = 3
	}
	if cfg.Topic == "" {
		cfg.Topic = "null"
	}
	cfg.MesosTaskID = fromEnvWithDefault("MESOS_TASK_ID", "null")
	cfg.EleapposLabel = fromEnvWithDefault("ELEAPPOS_LABEL", "null")
	cfg.EleapposSlaveFqdn = fromEnvWithDefault("ELEAPPOS_SLAVE_FQDN", "null")
	if cfg.Logger == nil {
		cfg.Logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	}
	if cfg.Remoter == nil {
		cfg.Remoter = NewRemoter(cfg.EtraceConfigURL, cfg.AppID, cfg.HostIP, cfg.EtraceConfigInterval, cfg.EtraceHTTPTimeout)
	}
	return cfg
}

func fromEnvWithDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultValue
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok {
			if !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("local IP not found")
}
