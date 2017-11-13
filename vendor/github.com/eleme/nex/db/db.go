package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/damnever/cc"
	mysql "github.com/eleme/mysql" // Registering mysql driver
	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/endpoint"
	"github.com/eleme/nex/log"
	"github.com/eleme/nex/metric"
	"github.com/eleme/nex/timeout"
	"github.com/eleme/nex/tracking/etrace"
	pg "github.com/eleme/pq"
	tracker "github.com/eleme/thrift-tracker"
	"github.com/jmoiron/sqlx"
	json "github.com/json-iterator/go"
)

var (
	// needCommentOps is the list of operations that need to be logged
	needCommentOps = [7]string{"SELECT", "DELETE", "COMMIT", "ROLLBACK", "UPDATE", "INSERT", "SET"}
	defaultDriver  = "mysql"
)

const (
	defaultMaxOpenConns  int   = 500
	defaultMaxIdleConns  int   = 300
	defaultMaxLifetime   int64 = 300 // s
	defaultEnableLog     bool  = false
	defaultEnableMetrics bool  = false
)

// DB is nex's wrapper on top of db.DB.
type DB struct {
	*sqlx.DB
	appName                 string
	enableMetrics           bool
	enableLog               bool
	logger                  log.RPCContextLogger
	execContextEndpoint     endpoint.Endpoint
	beginTxEndpoint         endpoint.Endpoint
	pingContextEndpoint     endpoint.Endpoint
	queryContextEndpoint    endpoint.Endpoint
	queryRowContextEndpoint endpoint.Endpoint
}

type dbSet struct {
	master *DB
	slave  *DB
}

// BeginTx starts a transaction. All write operations should be done within a transaction.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	select {
	case <-ctx.Done():
		return nil, timeout.ErrTimeout
	default:
	}

	ctx = context.WithValue(ctx, ctxkeys.OthAPIName, "BEGIN")
	result, err := db.beginTxEndpoint(ctx, opts)
	if err != nil {
		return nil, err
	}
	if tx, ok := result.(*Tx); ok {
		return tx, nil
	}
	panic("Type assertion error!")
}

func makeBeginTxEndpoint(db *DB) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*sql.TxOptions)
		tx, err := db.DB.BeginTx(ctx, req)
		if err != nil {
			return nil, err
		}
		return newTx(ctx, tx, db)
	}
}

// PingContext tests if connection to mysql is ok
func (db *DB) PingContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return timeout.ErrTimeout
	default:
	}

	ctx = context.WithValue(ctx, ctxkeys.OthAPIName, "PING")
	_, err := db.pingContextEndpoint(ctx, nil)
	return err
}

func makePingContextEndpoint(db *DB) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return nil, db.DB.PingContext(ctx)
	}
}

// QueryContext executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return queryContextImpl(ctx, db.queryContextEndpoint, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row. QueryRowContext always returns a non-nil value. Errors are deferred until Row's Scan method is called.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return queryRowContextImpl(ctx, db.queryRowContextEndpoint, query, args...)
}

// ExecContext executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return execContextEndpointImpl(ctx, db.execContextEndpoint, query, args...)
}

func (db *DB) attachMiddleWare(middleWare endpoint.Middleware) {
	db.execContextEndpoint = middleWare(db.execContextEndpoint)
	db.beginTxEndpoint = middleWare(db.beginTxEndpoint)
	db.pingContextEndpoint = middleWare(db.pingContextEndpoint)
	db.queryContextEndpoint = middleWare(db.queryContextEndpoint)
	db.queryRowContextEndpoint = middleWare(db.queryRowContextEndpoint)
}

func openDB(appName string, logger log.RPCContextLogger, dsn string, dbConfig cc.Configer) (*DB, error) {
	driver := dbConfig.StringOr("driver", defaultDriver)
	dsn = attachDBSettings(driver, dsn)
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	maxLifetime := dbConfig.DurationOr("max_lifetime", defaultMaxLifetime) * time.Second
	maxOpenConns := dbConfig.IntOr("max_open_conns", defaultMaxIdleConns)
	maxIdleConns := dbConfig.IntOr("max_idle_conns", defaultMaxIdleConns)

	db.SetConnMaxLifetime(maxLifetime)
	db.SetMaxIdleConns(maxOpenConns)
	db.SetMaxOpenConns(maxIdleConns)

	nexDB := &DB{DB: db, appName: appName, logger: logger}

	nexDB.execContextEndpoint = makeExecContextEndpoint(db)
	nexDB.beginTxEndpoint = makeBeginTxEndpoint(nexDB)
	nexDB.pingContextEndpoint = makePingContextEndpoint(nexDB)
	nexDB.queryContextEndpoint = makeQueryContextEndpoint(db)
	nexDB.queryRowContextEndpoint = makeQueryRowContextEndpoint(db)

	nexDB.enableLog = dbConfig.BoolOr("enable_log", defaultEnableLog)
	nexDB.enableMetrics = dbConfig.BoolOr("enable_metrics", defaultEnableMetrics)

	nexDB.attachMiddleWare(etrace.EndpointEtraceDBMiddleware)
	if nexDB.enableLog {
		nexDB.attachMiddleWare(log.EndpointLoggingCommonMiddleware(logger))
	}
	if nexDB.enableMetrics {
		nexDB.attachMiddleWare(metric.EndpointStatsdCommonMiddleware(appName, driver))
	}

	nexDB.attachMiddleWare(appNameCtxMiddleware(appName))

	return nexDB, nil
}

func openDBSet(appName string, logger log.RPCContextLogger, dbConfig cc.Configer) (*dbSet, error) {
	master, err := openDB(appName, logger, dbConfig.String("master"), dbConfig)
	if err != nil {
		return nil, err
	}
	slave, err := openDB(appName, logger, dbConfig.String("slave"), dbConfig)
	if err != nil {
		master.Close()
		return nil, err
	}
	return &dbSet{
		master: master,
		slave:  slave,
	}, nil
}

// Manager is the manager for all database connections.
type Manager struct {
	dbs map[string]*dbSet
}

// NewDBManager creates a new DBManager.
// appName is the current application's app_name,
// jsonDBSettings just like we defined in the package begin.
func NewDBManager(appName string, logger log.RPCContextLogger, jsonDBSettings string) (*Manager, error) {
	var dbSettings map[string]*json.RawMessage
	err := json.Unmarshal([]byte(jsonDBSettings), &dbSettings)
	if err != nil {
		return nil, err
	}

	dbm := &Manager{dbs: make(map[string]*dbSet, len(dbSettings))}
	for name, rawSetting := range dbSettings {
		dbConfig, err := cc.NewConfigFromJSON(*rawSetting)
		if err != nil {
			dbm.CloseAll()
			return nil, err
		}
		dbSet, err := openDBSet(appName, logger, dbConfig)
		if err != nil {
			dbm.CloseAll()
			return nil, err
		}
		dbm.dbs[name] = dbSet
	}
	return dbm, nil
}

// GetDBMaster return the master connection by name, return nil if not found.
func (dbm *Manager) GetDBMaster(name string) *DB {
	dbSet, exists := dbm.dbs[name]
	if !exists {
		return nil
	}
	return dbSet.master
}

// GetDBSlave return the slave connection by name, return nil if not found.
func (dbm *Manager) GetDBSlave(name string) *DB {
	dbSet, exists := dbm.dbs[name]
	if !exists {
		return nil
	}
	return dbSet.slave
}

// CloseAll close all db.
func (dbm *Manager) CloseAll() {
	for _, dbSet := range dbm.dbs {
		dbSet.master.Close()
		dbSet.slave.Close()
	}
}

func attachDBSettings(driver, dsn string) string {
	if driver == defaultDriver {
		dsn += "?interpolateParams=true&autocommit=0"
	}
	return dsn
}

func init() {
	mysql.SetSQLPreprocessor(insertContextComment)
	pg.SetSQLPreprocessor(insertContextComment)
}

// insertContextComment will prepend a comment to query, in this way we can pass additional information to DAL
func insertContextComment(ctx context.Context, query string) (string, error) {
	var needInsert bool

	tmpQuery := strings.ToUpper(query)
	for _, ops := range needCommentOps {
		if strings.HasPrefix(tmpQuery, ops) {
			needInsert = true
			break
		}
	}

	if !needInsert {
		return query, nil
	}

	candidates := make([]string, 0, 10)

	if value, is := ctx.Value(ctxkeys.AppName).(string); is {
		candidates = append(candidates, fmt.Sprintf("appid=%s", value))
	} else {
		candidates = append(candidates, "appid=unknown")
	}

	if value, is := ctx.Value(tracker.CtxKeyRequestID).(string); is {
		candidates = append(candidates, fmt.Sprintf("rid=%s", value))
	}

	if value, is := ctx.Value(tracker.CtxKeySequenceID).(string); is {
		candidates = append(candidates, fmt.Sprintf("rpcid=%s", value))
	}

	if requestMeta, is := ctx.Value(tracker.CtxKeyRequestMeta).(map[string]string); is {
		if shardingKey, in := requestMeta["routing-key"]; in {
			shardingKeySegs := strings.SplitN(shardingKey, "=", 2)
			if len(shardingKeySegs) == 2 {
				shardingKey, shardingValue := shardingKeySegs[0], shardingKeySegs[1]
				candidates = append(candidates, fmt.Sprintf("shardkey=%s", shardingKey))
				candidates = append(candidates, fmt.Sprintf("shardvalue=%s", shardingValue))
			}
		}
	}

	if len(candidates) > 0 {
		query = fmt.Sprintf("/*E:%s:E*/ %s", strings.Join(candidates, "&"), query)
	}
	return query, nil
}
