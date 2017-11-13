package db

import (
	"context"
	"database/sql"

	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/endpoint"
	nexlog "github.com/eleme/nex/log"
	"github.com/eleme/nex/metric"
	"github.com/eleme/nex/tracking/etrace"
)

// Tx is a wrapper on database/sql/Tx
type Tx struct {
	*sql.Tx
	ctx                     context.Context
	execContextEndpoint     endpoint.Endpoint
	queryContextEndpoint    endpoint.Endpoint
	commitEndpoint          endpoint.Endpoint
	queryRowContextEndpoint endpoint.Endpoint
	rollbackEndpoint        endpoint.Endpoint
}

// QueryContext executes a query that returns rows, typically a SELECT.
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return queryContextImpl(ctx, tx.queryContextEndpoint, query, args...)
}

// ExecContext executes a query that doesn't return rows. For example: an INSERT and UPDATE.
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return execContextEndpointImpl(ctx, tx.execContextEndpoint, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row. QueryRowContext always returns a non-nil value. Errors are deferred until Row's Scan method is called.
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return queryRowContextImpl(ctx, tx.queryRowContextEndpoint, query, args...)
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() error {
	ctx := context.WithValue(tx.ctx, ctxkeys.OthAPIName, "ROLLBACK")
	_, err := tx.rollbackEndpoint(ctx, nil)
	return err
}

// Commit commits the transaction.
func (tx *Tx) Commit() error {
	ctx := context.WithValue(tx.ctx, ctxkeys.OthAPIName, "COMMIT")
	_, err := tx.commitEndpoint(ctx, nil)
	return err
}

func makeTxRollbackEndpoint(tx *sql.Tx) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, tx.Rollback()
	}
}

func makeTxCommitEndpoint(tx *sql.Tx) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, tx.Commit()
	}
}

func (tx *Tx) attachMiddleWare(middleWare endpoint.Middleware) {
	tx.execContextEndpoint = middleWare(tx.execContextEndpoint)
	tx.queryRowContextEndpoint = middleWare(tx.queryRowContextEndpoint)
	tx.queryContextEndpoint = middleWare(tx.queryContextEndpoint)
	tx.rollbackEndpoint = middleWare(tx.rollbackEndpoint)
	tx.commitEndpoint = middleWare(tx.commitEndpoint)
}

func newTx(ctx context.Context, tx *sql.Tx, db *DB) (*Tx, error) {
	nexTx := &Tx{Tx: tx, ctx: ctx}

	nexTx.execContextEndpoint = makeExecContextEndpoint(tx)
	nexTx.queryRowContextEndpoint = makeQueryRowContextEndpoint(tx)
	nexTx.queryContextEndpoint = makeQueryContextEndpoint(tx)
	nexTx.rollbackEndpoint = makeTxRollbackEndpoint(tx)
	nexTx.commitEndpoint = makeTxCommitEndpoint(tx)

	nexTx.attachMiddleWare(etrace.EndpointEtraceDBMiddleware)
	if db.enableLog {
		nexTx.attachMiddleWare(nexlog.EndpointLoggingCommonMiddleware(db.logger))
	}

	if db.enableMetrics {
		nexTx.attachMiddleWare(metric.EndpointStatsdCommonMiddleware(db.appName, "mysql"))
	}
	nexTx.attachMiddleWare(appNameCtxMiddleware(db.appName))

	return nexTx, nil
}
