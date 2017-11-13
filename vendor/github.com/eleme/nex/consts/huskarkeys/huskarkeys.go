// Package huskarkeys contains the config keys in Huskar.
package huskarkeys

const (
	// RPCCluster is the key format to get the rpc client's cluster from Huskar.
	RPCCluster string = "CLUSTER:%s"
	// RPCIntent is the key format to get the rpc client's SOA intent from Huskar.
	RPCIntent string = "INTENT:%s"
	// RedisSettings is the key to get redis settings from Huskar.
	RedisSettings string = "REDIS_SETTINGS"
	// DBSettings  is the key to get db settings from Huskar.
	DBSettings string = "DB_SETTINGS"
	// HardTimeout is the key format to get the timeout(ms) for API.
	HardTimeout string = "HARD_TIMEOUT:%s"
	// KafkaSettings is the key to get kafka settings form Huskar.
	KafkaSettings string = "KAFKA_SETTINGS"
)
