package db

import (
	"context"
	"database/sql"

	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/endpoint"
	"github.com/eleme/nex/timeout"
	"github.com/eleme/nex/tracking/etrace"
)

type execContextRequest struct {
	query string
	args  []interface{}
}

type execContextRunner interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type queryRowContextRunner interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type queryContextRunner interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func makeQueryContextEndpoint(impl queryContextRunner) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(execContextRequest)
		return impl.QueryContext(ctx, req.query, req.args...)
	}
}

func queryContextImpl(ctx context.Context, impl endpoint.Endpoint, query string, args ...interface{}) (*sql.Rows, error) {
	select {
	case <-ctx.Done():
		return nil, timeout.ErrTimeout
	default:
	}

	req := execContextRequest{query: query, args: args}
	ctx = context.WithValue(ctx, ctxkeys.OthAPIName, etrace.BuildSQL(query, args...))
	result, err := impl(ctx, req)
	if err != nil {
		return nil, err
	}
	if rows, ok := result.(*sql.Rows); ok {
		return rows, nil
	}
	panic("Type assertion error!")

}

func makeQueryRowContextEndpoint(impl queryRowContextRunner) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(execContextRequest)
		return impl.QueryRowContext(ctx, req.query, req.args...), nil
	}
}

func queryRowContextImpl(ctx context.Context, impl endpoint.Endpoint, query string, args ...interface{}) *sql.Row {
	req := execContextRequest{query: query, args: args}
	ctx = context.WithValue(ctx, ctxkeys.OthAPIName, etrace.BuildSQL(query, args...))
	result, _ := impl(ctx, req)
	if sqlResult, ok := result.(*sql.Row); ok {
		return sqlResult
	}
	panic("Type assertion error!")
}

func makeExecContextEndpoint(impl execContextRunner) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(execContextRequest)
		return impl.ExecContext(ctx, req.query, req.args...)
	}
}

func execContextEndpointImpl(ctx context.Context, impl endpoint.Endpoint, query string, args ...interface{}) (sql.Result, error) {
	select {
	case <-ctx.Done():
		return nil, timeout.ErrTimeout
	default:
	}

	req := execContextRequest{query: query, args: args}
	ctx = context.WithValue(ctx, ctxkeys.OthAPIName, etrace.BuildSQL(query, args...))
	result, err := impl(ctx, req)
	if err != nil {
		return nil, err
	}
	if sqlResult, ok := result.(sql.Result); ok {
		return sqlResult, nil
	}
	panic("Type assertion error!")
}
