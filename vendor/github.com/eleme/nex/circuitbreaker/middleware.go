package circuitbreaker

import (
	"context"

	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/endpoint"
)

// endpointCircuitBreakerSOAMiddleware is a middleware which do useful works like circuit breaker.
func endpointCircuitBreakerSOAMiddleware(cb *CircuitBreaker, who string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			var api string
			if who == "soa" {
				api = ctx.Value(ctxkeys.APIName).(string)
			} else if who == "soa_client" {
				api = ctx.Value(ctxkeys.RemoteAddr).(string)
			}
			defer cb.RecordMetrics(api, &err)
			err = cb.Test(api)
			if err != nil {
				return
			}

			response, err = next(ctx, request)
			return
		}
	}
}

// EndpointCircuitBreakerSOAServerMiddleware is used for soa server side.
func EndpointCircuitBreakerSOAServerMiddleware(cb *CircuitBreaker) endpoint.Middleware {
	return endpointCircuitBreakerSOAMiddleware(cb, "soa")
}

// EndpointCircuitBreakerSOAClientMiddleware is used for soa client side.
func EndpointCircuitBreakerSOAClientMiddleware(cb *CircuitBreaker) endpoint.Middleware {
	return endpointCircuitBreakerSOAMiddleware(cb, "soa_client")
}
