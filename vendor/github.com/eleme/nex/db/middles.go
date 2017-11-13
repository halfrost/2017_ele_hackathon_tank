package db

import (
	"context"

	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/endpoint"
)

func appNameCtxMiddleware(appName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx = context.WithValue(ctx, ctxkeys.AppName, appName)
			return next(ctx, request)
		}
	}
}
