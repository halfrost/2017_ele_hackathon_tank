package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eleme/huskar-pool/pool"
	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/endpoint"
	"github.com/eleme/nex/log"
	"github.com/eleme/nex/timeout"
	"github.com/eleme/nex/tracking/etrace"

	"github.com/garyburd/redigo/redis"
)

// Client represents a client connection to Redis server.
type Client struct {
	connFactory connFactory
	addr        string
	logger      log.RPCContextLogger
	doEndpoint  endpoint.Endpoint
}

// NewClient creates a new redis client without pool support.
func NewClient(addr string, logger log.RPCContextLogger, options ...redis.DialOption) (*Client, error) {
	conn, err := newRedisConn(addr, options...)
	if err != nil {
		return nil, err
	}
	connFactory := newRawConnFactory(conn)
	return newClient(addr, connFactory, logger), nil
}

type connFactory interface {
	Get(ctx context.Context) (pool.Resource, error)
	Put(r pool.Resource)
	Close()
}

type rawConn struct {
	conn redis.Conn
}

func (rc *rawConn) Close() error {
	return rc.conn.Close()
}

type rawConnFactory struct {
	rawConn *rawConn
}

func newRawConnFactory(conn redis.Conn) *rawConnFactory {
	return &rawConnFactory{
		rawConn: &rawConn{conn: conn},
	}
}

func (rcf *rawConnFactory) Get(_ context.Context) (pool.Resource, error) { return rcf.rawConn, nil }

func (rcf *rawConnFactory) Put(_ pool.Resource) {}

func (rcf *rawConnFactory) Close() { rcf.rawConn.Close() }

func (*rawConnFactory) Stats() (capacity, available, maxCap, waitCount int64, waitTime, idleTimeout time.Duration) {
	return
}
func (*rawConnFactory) StatsJSON() string         { return "{}" }
func (*rawConnFactory) StatsReadableJSON() string { return "{}" }

func newRedisConn(addr string, options ...redis.DialOption) (redis.Conn, error) {
	var conn redis.Conn
	var err error
	if strings.HasPrefix(addr, "redis://") {
		conn, err = redis.DialURL(addr, options...)
	} else {
		conn, err = redis.Dial("tcp", addr, options...)
	}
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func newClient(addr string, connF connFactory, logger log.RPCContextLogger) *Client {
	doEndpoint := makeRedisDoCmdEndpoint(connF)
	// NOTE: add middlewares here.
	doEndpoint = etrace.EndpointEtraceRedisMiddleware(doEndpoint)
	if logger != nil {
		doEndpoint = log.EndpointLoggingCommonMiddleware(logger)(doEndpoint)
	}
	return &Client{
		logger:      logger,
		connFactory: connF,
		addr:        addr,
		doEndpoint:  doEndpoint,
	}
}

type doCmdRequest struct {
	CmdName string
	Args    []interface{}
}

func makeRedisDoCmdEndpoint(connF connFactory) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(doCmdRequest)
		conn, err := connF.Get(ctx)
		if err != nil {
			return nil, err
		}
		redisConn := conn.(*rawConn).conn
		response, err := redisConn.Do(req.CmdName, req.Args...)
		if err != nil {
			if redisConn.Err() != nil {
				redisConn.Close()
				connF.Put(nil)
			}
			return nil, err
		}
		connF.Put(conn)
		return response, nil
	}
}

// Do override redis.Conn.Do method with metrics recording.
func (c *Client) Do(ctx context.Context, commandName string, args ...interface{}) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, timeout.ErrTimeout
	default:
	}
	ctx = context.WithValue(ctx, ctxkeys.OthAPIName, fmt.Sprintf("%s@%s", c.addr, commandName))
	cmdName := strings.ToLower(commandName)
	request := doCmdRequest{CmdName: cmdName, Args: args}
	return c.doEndpoint(ctx, request)
}

// Close closes the connection, close all, if it has a pool in it.
func (c *Client) Close() error {
	c.connFactory.Close()
	return nil
}

func formatInt(i int64) string {
	return strconv.FormatInt(i, 10)
}

func usePrecise(dur time.Duration) bool {
	return dur < time.Second || dur%time.Second != 0
}

func formatMs(dur time.Duration) string {
	return formatInt(int64(dur / time.Millisecond))
}

func formatSec(dur time.Duration) string {
	return formatInt(int64(dur / time.Second))
}

// Del delete a key
func (c *Client) Del(ctx context.Context, keys ...string) (int, error) {
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	return redis.Int(c.Do(ctx, "DEL", args...))
}

// Dump return a serialized version of the value stored at the specified key
func (c *Client) Dump(ctx context.Context, key string) (string, error) {
	return redis.String(c.Do(ctx, "DUMP", key))
}

// Exists determine if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	return redis.Bool(c.Do(ctx, "EXISTS", key))
}

// Expire set a key's time to live in seconds
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return redis.Bool(c.Do(ctx, "EXPIRE", key, formatSec(expiration)))
}

// ExpireAt set the expiration for a key as a UNIX timestamp
func (c *Client) ExpireAt(ctx context.Context, key string, ts time.Time) (bool, error) {
	return redis.Bool(c.Do(ctx, "EXPIREAT", key, ts.Unix()))
}

// TTL get the time to live for a key
func (c *Client) TTL(ctx context.Context, key string) (int, error) {
	return redis.Int(c.Do(ctx, "TTL", key))
}

// Persist remove the expiration from a key
func (c *Client) Persist(ctx context.Context, key string) (bool, error) {
	return redis.Bool(c.Do(ctx, "PERSIST", key))
}

// PExpire set a key's time to live in milliseconds
func (c *Client) PExpire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return redis.Bool(c.Do(ctx, "PEXPIRE", key, formatMs(expiration)))
}

// PExpireAt set the expiration for a key as a UNIX timestamp specified in milliseconds
func (c *Client) PExpireAt(ctx context.Context, key string, ts time.Time) (bool, error) {
	return redis.Bool(c.Do(ctx, "PEXPIREAT", key, ts.UnixNano()/int64(time.Millisecond)))
}

// PTTL get the expiration for a key as a UNIX timestamp specified in milliseconds
func (c *Client) PTTL(ctx context.Context, key string) (int, error) {
	return redis.Int(c.Do(ctx, "PTTL", key))
}

// Restore create a key using the provided serialized value, previously obtained using DUMP
func (c *Client) Restore(ctx context.Context, key string, ttl time.Duration, value string) (string, error) {
	return redis.String(c.Do(ctx, "RESTORE", key, formatMs(ttl), value))
}

// Type determine the type stored at key
func (c *Client) Type(ctx context.Context, key string) (string, error) {
	return redis.String(c.Do(ctx, "TYPE", key))
}

// Info get information and statistics about the server
func (c *Client) Info(ctx context.Context) ([]byte, error) {
	return redis.Bytes(c.Do(ctx, "INFO"))
}

// Time return the current server time
func (c *Client) Time(ctx context.Context) ([]int, error) {
	return redis.Ints(c.Do(ctx, "TIME"))
}

// Eval execute a Lua script server side
func (c *Client) Eval(ctx context.Context, script string, keys []string, args []string) ([]interface{}, error) {
	cmdArgs := make([]interface{}, 2+len(keys)+len(args))
	cmdArgs[0] = script
	cmdArgs[1] = strconv.Itoa(len(keys))
	for i, key := range keys {
		cmdArgs[2+i] = key
	}
	pos := 2 + len(keys)
	for i, arg := range args {
		cmdArgs[pos+i] = arg
	}
	return redis.MultiBulk(c.Do(ctx, "EVAL", cmdArgs...))
}
