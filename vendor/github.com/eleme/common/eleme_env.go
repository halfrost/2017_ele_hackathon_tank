package common

import (
	"errors"
	"io/ioutil"
	"net"
	"strings"

	"gopkg.in/yaml.v2"
)

// ZkConnCfg contains connection information of ZooKeeper
type ZkConnCfg struct {
	Hosts string `yaml:"hosts"`
	User  string `yaml:"username"`
	Pwd   string `yaml:"password"`
}

// HostList returns []string splited from Hosts
func (zk *ZkConnCfg) HostList() []string {
	var hosts []string
	for _, h := range strings.Split(zk.Hosts, ",") {
		if len(strings.TrimSpace(h)) > 0 {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

// Auth returns a string of "User:Pwd"
func (zk *ZkConnCfg) Auth() string {
	if zk.User != "" || zk.Pwd != "" {
		return strings.Join([]string{zk.User, zk.Pwd}, ":")
	}
	return ""
}

// ElemeEnv is a struct mapping /etc/eleme/env.yaml
type ElemeEnv struct {
	Env     string    `yaml:"env"`
	Cluster string    `yaml:"cluster"`
	Zk      ZkConnCfg `yaml:"zookeeper_configs"`
	OpsDB   struct {
		URL string `yaml:"url"`
	} `yaml:"ops_db"`
	StatsdURL string `yaml:"statsd_url"`
	IDC       string `yaml:"idc"`
}

// StatsdAddr returns statsd host with port e.g. xg-statsd.elenet.me:8125
func (elemeEnv *ElemeEnv) StatsdAddr() string {
	statsdURL := strings.TrimSpace(elemeEnv.StatsdURL)
	host := strings.TrimPrefix(statsdURL, "statsd://")
	host = strings.TrimSpace(host)

	if host != "" {
		if _, err := net.ResolveTCPAddr("tcp", host); err == nil {
			return host
		}
	}
	return ""
}

// ReadElemeEnv creates ElemeEnv from given file path
func ReadElemeEnv(filename string) (*ElemeEnv, error) {
	elemeEnv := new(ElemeEnv)

	fl, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New("Error parsing eleme config file")
	}

	err = yaml.Unmarshal(fl, elemeEnv)
	if err != nil {
		return nil, err
	}

	if elemeEnv.IDC == "" {
		// GOD bless...
		// no IDC given, default to XG
		elemeEnv.IDC = "xg"
	}
	return elemeEnv, nil
}
