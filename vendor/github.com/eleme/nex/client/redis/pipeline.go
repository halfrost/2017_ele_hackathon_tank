package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/eleme/huskar-pool/pool"
	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/endpoint"
	"github.com/eleme/nex/log"
	"github.com/eleme/nex/timeout"
	"github.com/eleme/nex/tracking/etrace"
	"github.com/garyburd/redigo/redis"
)

var (
	// ErrZeroCmdToExecute is returned when no commands in pipeline can be send.
	ErrZeroCmdToExecute = errors.New("no command can be execute, must call Command first")
)

// PipelineReply represents a single one pipeline reply.
type PipelineReply struct {
	V   interface{}
	Err error
}

// Pipeline sends a batch commands to the server once, transaction is not support.
type Pipeline struct {
	ctx          context.Context
	resource     pool.Resource
	client       *Client
	cmds         []doCmdRequest
	execEndpoint endpoint.Endpoint
}

// Pipeline creates a new pipeline, always create a new pipeline to do pipeline work.
// Example:
//     pipe, err := client.Pipeline(context.TODO())
//     assert(err != nil)
//     pipe.Command("SET", "key1", "val1")
//     pipe.Command("SET", "key2", "val2")
//     replies, err := pipe.Execute()
//     assert(err != nil)
//     for _, reply := range replies {
//         v, err := redis.String(reply.V, reply.Err)
//         assert(err != nil)
//         assert(v == "OK")
//     }
func (c *Client) Pipeline(ctx context.Context) (*Pipeline, error) {
	conn, err := c.connFactory.Get(ctx)
	if err != nil {
		return nil, err
	}
	redisConn := conn.(*rawConn).conn
	execEndpoint := makePipelineExecEndpoint(redisConn)
	execEndpoint = etrace.EndpointEtraceRedisMiddleware(execEndpoint)
	if c.logger != nil {
		execEndpoint = log.EndpointLoggingCommonMiddleware(c.logger)(execEndpoint)
	}
	return &Pipeline{
		ctx:          ctx,
		resource:     conn,
		client:       c,
		cmds:         []doCmdRequest{},
		execEndpoint: execEndpoint,
	}, nil
}

// Command caches the redis command.
func (p *Pipeline) Command(cmd string, args ...interface{}) {
	request := doCmdRequest{CmdName: cmd, Args: args}
	p.cmds = append(p.cmds, request)
}

// Execute sends all pipelined commands to the server.
func (p *Pipeline) Execute() ([]PipelineReply, error) {
	select {
	case <-p.ctx.Done():
		return nil, timeout.ErrTimeout
	default:
	}

	ctx := context.WithValue(p.ctx, ctxkeys.OthAPIName, fmt.Sprintf("%s@PIPELINE{%d}", p.client.addr, len(p.cmds)))
	reply, err := p.execEndpoint(ctx, p.cmds)

	if err != nil && err != ErrZeroCmdToExecute {
		p.client.connFactory.Put(nil)
		return nil, err
	}
	p.client.connFactory.Put(p.resource)
	return reply.([]PipelineReply), nil
}

func makePipelineExecEndpoint(conn redis.Conn) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cmds := request.([]doCmdRequest)
		n := len(cmds)
		if n == 0 {
			return nil, ErrZeroCmdToExecute
		}

		for _, cmd := range cmds {
			if err := conn.Send(cmd.CmdName, cmd.Args...); err != nil {
				return nil, err
			}
		}
		if err := conn.Flush(); err != nil {
			return nil, err
		}

		replies := make([]PipelineReply, n, n)
		for i := 0; i < n; i++ {
			reply, err := conn.Receive()
			replies[i] = PipelineReply{V: reply, Err: err}
		}

		return replies, conn.Err()
	}
}
