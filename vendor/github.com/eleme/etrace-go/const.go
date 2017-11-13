package etrace

const (
	// TypeService is the type name for SOA service transaction.
	TypeService = "SOAService"
	// TypeCall is the type name for SOA client transaction.
	TypeCall = "SOACall"
	// TypeRemoteCall is the type name for remote call event.
	TypeRemoteCall = "RemoteCall"
	// TypeLink is the type name for remote call link event.
	TypeLink = "EtraceLink"
	// TypeSQL is the type name for SQL related transaction.
	TypeSQL = "SQL"
	// TypeRedis is the type name for Redis related transaction.
	TypeRedis = "Redis"
	// TypeURL is the type name for HTTP related transaction.
	TypeURL = "URL"

	// TypeRMQProduce is the type name for mq producer related transaction.
	TypeRMQProduce = "RMQ_PRODUCE"
	// TypeRMQConsume is the type name for mq consumer related transaction.
	TypeRMQConsume = "RMQ_CONSUME"
	// TypeException is the type name for exception.
	TypeException = "Exception"
	// TagServiceClientApp is the tag name to record client's app id on SOA service side.
	TagServiceClientApp = "SOAService.clientApp"
	// TagServiceClientIP is the tag name to record client's IP on SOA service side.
	TagServiceClientIP = "SOAService.clientIP"
	// TagServiceResult is the tag name to record the result on SOA service side.
	TagServiceResult = "SOAService.resultCode"
	// TagCallServiceApp is the tag name to record server's app id on SOA client side.
	TagCallServiceApp = "SOACall.serviceApp"
	// TagCallServiceIP is the tag name to record server's Ip on SOA client side.
	TagCallServiceIP = "SOACall.serviceIP"
	// TagCallResult is the tag name to record the result on SOA client side.
	TagCallResult = "SOACall.resultCode"
	// TagShadingKey is the tag name to record shading key.
	TagShadingKey = "shardingkey"
	// TagSQLDatabase TODO.
	TagSQLDatabase = "SQL.database"
	// TagSQLMethod TODO.
	TagSQLMethod = "SQL.method"

	// StatusSuccess is the status value when success.
	StatusSuccess = "0"
)
