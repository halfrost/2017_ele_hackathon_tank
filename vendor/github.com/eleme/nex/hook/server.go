package hook

var (
	// BeforeServerStarting is triggered before thrift server open listening socket.
	BeforeServerStarting = NewNotifier("before_server_starting")
	// BeforeServerStoping is triggered only if there is no error(e.g. deregister app from Huskar)
	// before thrift server close listening socket,  server will wait for graceful time
	// if there is any request still processing.
	// DO NOT register any Observer which has a blocking OnNotify method.
	BeforeServerStoping = NewNotifier("before_server_stoping")
)
