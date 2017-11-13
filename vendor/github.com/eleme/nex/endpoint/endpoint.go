// Package endpoint defines an abstraction for RPCs.
//
// ## How to write a endpoint-based middleware:
//
// - Without extra args:
//  func EndpointXXXMiddleware(next Endpoint) Endpoint {
//		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
//			defer func(begin time.Time) {
//				// DO XXX THINGS HERE
//    		}(time.Now())
//			return next(ctx, request)
//		}
//	}
//
// - With extra args:
//	func EndpointXXXMiddleware(args ArgType, ...) Middleware {
//		return func(next Endpoint) Endpoint {
//			return func(ctx context.Context, request interface{}) (response interface{}, err error) {
//				defer func(begin time.Time) {
//					// DO XXX THINGS HERE
//				}(time.Now())
//				return next(ctx, request)
//			}
//		}
//	}
//
// ## Get it to work:
// Add it to cmd/gogen/template.
package endpoint

import (
	"context"
)

// Endpoint is the fundamental building block of servers and clients.
// It represents a single RPC method.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

// Nop is an endpoint that does nothing and returns a nil error.
// Useful for tests.
func Nop(context.Context, interface{}) (interface{}, error) { return struct{}{}, nil }

// Middleware is a chainable behavior modifier for endpoints.
type Middleware func(Endpoint) Endpoint

// Chain is a helper function for composing middlewares. Requests will
// traverse them in the order they're declared. That is, the first middleware
// is treated as the outermost middleware.
func Chain(outer Middleware, others ...Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(others) - 1; i >= 0; i-- { // reverse
			next = others[i](next)
		}
		return outer(next)
	}
}
