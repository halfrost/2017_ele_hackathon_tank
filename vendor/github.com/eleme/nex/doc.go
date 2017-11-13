// Package nex is eleme's Golang thrift service framework.
//
// The design principle is as following:
//
//  * DO NOT share global variables across the entire project, keep the structure stupid.
//  * Use code generation to walk around golang's limited meta programming capability.
//  * All sub packages should act as an independent library without special environment dependencies, including generated code.
//  * Unified method to add middleware to any kinds of component(with `endpoint.Endpoint`).
//  * Enjoy it or refactor it..
//
// Something you should know:
//
//  * The current API is not stable, be careful.
//  * Always pass `ctx context.Context` to every function call.
//  * There're lots of generated codes, leave'em be. (filename starts with `auto-` or there's comment stating not to modify)
//  * Golang requires lots of typing, it's normal(compared to python).
//
// Quick start
//
// Execute the following bash command for quick bootstrapping.
//
//  $ # Install nex the easy way, tested in MacOS
//  $ curl -L https://goo.gl/f3rYM8 | bash
//  $ # Bootstrap a nex app
//  $ nex bootstrap --appID arch.nex --serviceName Note # you may need --thriftFile option
//
// Add/modify thrift interface:
//
//  1. Modify your thrift file(Add/Change definitions), add service dependencies in `thriftfs` folder.
//  2. Regenerate code
//     $ nex regen
//  3. Write handler(Should follow `Service` interface generated in `auto-handler.go` file)
//  4. Recompile
//     $ make server client
//  5. (Optional, Recommended) Write a test to check logics(Will be ran in goci)
//  6. (Optional, Not recommended) Write a client(either in golang or python) to verify logics manually.
//  7. (Optional, Casual) Use https://github.com/wooparadog/thriftpy-cli/ to try your apis in IPython interactively.
//
// How to deploy to servers
//
// Nex comes with a deployment suite. It includes various scripts to deploy to online servers by eless.
// If you already have an app id in eless, you only have to add eless hooks in your web hook settings,
// see http://wiki.ele.to:8090/pages/viewpage.action?pageId=39204511 for details.
//
// Meanwhile you should know the following things:
//
//  * Nex uses `systemd` to manage process. The service name is `${APP_ID}.service`. So you can:
//    $ sudo systemctl restart ${APP_ID} # to restart service
//    $ sudo systemctl status ${APP_ID}  # to check service status
//  * Nex uses `rsyslog` and `logrotate` to manage logs.
//  * Logs will be stored in: `/data/log/app/${app_id}/`. It will be automatically sync with `elk`
//  * Logs will rotate every day and will be kept for 7 days
//
// How to upgrade to new version
//
// Nex uses godep as package management tool. If you want to upgrade to a newer version:
//
//  $ # First, need to fetch the newest code
//  $ go get -u github.com/eleme/nex
//  $ # cd into nex
//  $ cd `go list -f '{{.Dir}}' github.com/eleme/nex`
//  $ # checkout out the desired nex version
//  $ git checkout v0.1
//  $ # restore all dependencies to specified version in nex
//  $ godep restore  # Ref: https://github.com/tools/godep
//  $ make install
//
//  $ cd `go list -f '{{.Dir}}' github.com/eleme/your-project`
//  $ # rm Godeps and vendor dirs. (bug: https://github.com/tools/godep/issues/498)
//  $ rm -r Godeps vendor
//  $ # Recreate dependencies
//  $ make dep
//  $ # Regenerate codes
//  $ nex regen
//
// See http://github.com/eleme/nex for more information.
package nex
