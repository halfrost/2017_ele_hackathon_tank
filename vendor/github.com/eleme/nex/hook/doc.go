// Package hook provides interfaces for user that wish to
// listen for the occurrence of a specific event.
//
// The Server Hooks Example
//
// Server hooks take nil as the event argument currently, you can deal with it if you like :-)
//
//  type onServerStoping struct {}
//  func (h onServerStarting) OnNotify(_ interface{}) {
//      fmt.Println("before server stoping")  // NOTE: DO NOT DO TIME-CONSUMING WORK
//  }
//  BeforeServerStoping.Register(onServerStoping{})
package hook
