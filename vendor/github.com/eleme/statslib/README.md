statslib
=======

通过引入这个库，可以直接监控进程的相关性能指标，并发送到[statsd](https://github.com/etsy/statsd).  
> Please register a custom logger by call RegisterLogger before  StartStatsdService.


Usage
=======

```
	import "github.com/eleme/statslib"

	conf := statsd.DefaultConfig("127.0.0.1:8125", "mux")
	statsd.RegisterLogger(...)
	err := statsd.StartStatsdService(conf)
	if err != nil {
		log.Fatalln(err)
	}
	defer statsd.Close()
	......
	
	statsd.SetGauge([]string{"cpu"}, 34, true)
```


Metrics
=======
You can incdicates by config.RuntimeMetrics field.  

Currently the following internal runtime metrics are provided:
* runtime.num_goroutines
* runtime.gc_pause_ms
* runtime.gc_pause_total_ms
* runtime.alloc_bytes
* runtime.total_alloc_bytes
* runtime.sys_bytes
* runtime.heap_alloc_bytes
* runtime.heap_sys_bytes
* runtime.heap_idle_bytes
* runtime.heap_inuse_bytes
* runtime.heap_released_bytes
* runtime.heap_objects
* runtime.stack_inuse_bytes
* runtime.stack_sys_bytes
* runtime.num_gc