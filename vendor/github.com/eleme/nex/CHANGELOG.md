0.1.1
-----

* AMQP&Etrace[Fix]: Fix Etrace for AMQP. NOTE: the consumer handler is changed(amqp.Delivery->message.Delivery). #205
* AMQP[Fix]: Fix header(SOA_CONTEXT) type. #203

0.1.0
-----

* Log event with pool stats. #199
* Add servic name for RPC metrics. #196
* Fix ETrace: client app id && rpc id. #192 #193 #195
* Send exception event to ETrace. #190
* Add samaritan SDK. #186
* Add postgres support. #188
* Add util function to return default error message for TApplicationException if 1:message field is missing. #184
* Add command to update nex. #183
* Replace JSON lib. #178
* Signing the HTTP/JSON RPC request. #172
* Add `client_timeout` option && change default settings. #179
* Classify circuit breaker stats by host on client side. #176
* Add middleware to limit the max in progress requests. #174
* Distribute lock pressure to multiple loggers. #167
* Add empty local huskar config.json, toggle.json and gitignore. #173
* Refactor HTTP client(reduce memory allocs). #171
* Disable syslog in docker. #166
* Do not check context meta, bypass all user context meta. #165
* Update circuit breaker (refactor). #164
* Fix quickstart script. #163
* Add default handler to serve thrift file. #160
* Add meta info when register service to Huskar. #159 #169
* Simplify cluster for SOA modes(SOA route support, etc.). #158
* Fix potential race condition when return ErrTimeout from timeout middleware. #156
* Only close Redis connection when error is fatal. #153
* Add server hooks(BeforeServerStarting/BeforeServerStoping) & Reject ping request after server closed listening socket. #149
* Support adding custom middlewares to server side. #150
* Fix ETrace for message queue producer. Lock protection for Redis sub transactions, just in case. #145
* Ignore ping for ETrace, logging and circuit breaker, ignore unnecessary client log. #144
* Fix wrong config value when multiple cluster exists. #143
* Adjust cluster name for multi-datacenter. #142
* Fix ETrace total duration for Redis commands. #141
* Add Redis pipeline (without transaction). #139
* Ignore /etc/eleme/env.yaml in develop environment. #136
* Add a new logger with context support, may be BROKEN CHANGING. #132
* Remove outdated docs. #131
* Add builtin HTTP APIs for profiling.
* Recover panic from handler GoRoutine. #128 #129
* Fix ETrace usage(link event). #127
* Disable corlored log in docker. #126
* Use the MESOS_TASK_ID as the key to register service in docker. #125
* Add docker support for eless build script. #124
* Add graceful shutdown support and convert error into app defined thrift exception. #123
* Make client side code can be generated independently and add more docs. #122
* Enhance logging messages, marshal thrift exception to readable error message. #120 #121
* Add health checking for RPC clien. #119
* Update Huskar SDK. #118
* Add Huskar pool support for RPC client and remove goproxy support for RPC client. #117
* Only initialize service once. #114
* Ensure safe import in code template. #113
* Custom listening address in app.yaml or command line. #111 #115
* Add timeout controling middleware. #110
* Fix pkg-config issue in installation script. #109
* Add --projectName option for nex bootstrap. #107
* Fix wrong pool usage in redis and RPC client. #106
* Add ETrace support for AMQP client. #105
* Update ETrace SDK. #102
* Add more documents. #100 #101 #103 #104 #116
* Simplify thrift client usage. #99
* Add ETrace support for redis and simplify reids pool usage. #98
* Add API downgrading middleware and fix curcuit breaker. #97
* Enhance installation script. #92
* Add mock huskar client. #90
* Add more documents and tools. #87 #88 #91 #93 #94 #95 #96
* Add ETrace support for thrift server and RPC client. #86
* Add context and middlewares for db. #84
* Add template to generate eless build script. #83
* Add context for panic when initialize resources. #82
* Add syslog handler for logging. #81
* Add installation script. #80
* Add database connector. #77
* Add thrift tracker support. #76
* Add curcuit breaker middleware. #75
* Add git hooks. #74
* Add HTTP/JSON client support. #72
* Add context support for redis client. #70
* RPC client related fixes. #67 #68
* Add --thriftFile option for nex bootstrap. #65
* Add template to generate client side endpoints with context support, add client side statsd and logging middleware. #64
* Add server side statsd and logging middleware. #62
* Add template to generate server side endpoints with context support. #61
* New application structure. #56 #57 #58
* Fix not existing GOAPTH. #54

**NOTE: BROKEN CHANGING FROM HERE.**

0.0.3
-----

* Thrift client pool. #35
* Generate thrift client project template. #36
* Thrift client metrics record. #37
* Redis client metrics record. #38
* Read huskar config from environ. #39

0.0.2
-----

* Processor metrics record.
* Message Consumer/Producer.
* Register/Deregister self.
* RPC client with load balance.
* Corvus commands.


0.0.1
-----

* Implement the basic version of thrift server framework.
