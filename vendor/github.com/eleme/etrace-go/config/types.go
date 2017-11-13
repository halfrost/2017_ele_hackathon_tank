package config

// RemoteConfig is the configuration fetched from etrace server.
type RemoteConfig struct {
	ConfigKey      string `json:"configKey"`
	Enabled        bool   `json:"enabled"`
	TagCount       int    `json:"tagCount"`
	TagSize        int    `json:"tagSize"`
	DataSize       int    `json:"dataSize"`
	MessageCount   int    `json:"messageCount"`
	LongConnection bool   `json:"longConnection"`
}

func defaultRemoteConfig() RemoteConfig {
	return RemoteConfig{
		ConfigKey:      "",
		Enabled:        true,
		TagCount:       8,
		TagSize:        256,
		DataSize:       2048,
		MessageCount:   500,
		LongConnection: true,
	}
}

// Collector is the thrift service address where transaction messages will be sent to.
// The list of collectors is fetched from etrace server.
type Collector struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
