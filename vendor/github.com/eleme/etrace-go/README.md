#  golang etrace client#
```
package main
import etrace "github.com/eleme/etrace-go"
func main() {
	trace, _ := etrace.New(etrace.Config{
		AppID:           "web.mux",
		EtraceConfigURL: "etrace-config.alpha.elenet.me:2890",
	})
	root := trace.NewTransaction("web.mux^^^aaaaa", "api.gateway|1.1", "web.mux biz.box", "GET /v2/ping")
	child := root.Fork("type", "name")
	child.Commit()
	root.Commit()
}
```
