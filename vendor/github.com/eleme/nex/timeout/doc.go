// Package timeout provides timeout control middlewares for SOA,
// the timeout is based on API name which defines in thrift file.
//
// You can set timeout on Huksar config, the key format "HARD_TIMEOUT:{api_name}",
// the default timeout is 20*1e3 millisecond.
//
// NOTE: the actual timeout may be larger than the timeout value, that's understandable.
package timeout
