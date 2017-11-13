// Package redis implements all Corvus supported commands.
//
// If you use nex.GetRedisPools, you must set up "REDIS_SETTINGS" in Huskar config,
// the format detail is in NewPoolManager.
//
// The simple usage:
//
//  func GetVal(ctx context.Context) (string, error) (
//      rpm := nex.GetRedisPools()
//      c := rpm.GetPooledClient("nex")
//      return c.Get(ctx, "key")
//  }
//
//
// You can learn more from document, it is pretty simple.
package redis
