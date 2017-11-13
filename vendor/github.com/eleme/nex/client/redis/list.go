package redis

import (
	"context"

	"github.com/garyburd/redigo/redis"
)

// LIndex get an element from a list by its index
func (c *Client) LIndex(ctx context.Context, key string, index int) (string, error) {
	return redis.String(c.Do(ctx, "LINDEX", key, index))
}

// LInsert insert an element before or after another element in a list
func (c *Client) LInsert(ctx context.Context, key, op, pivot, value string) (int, error) {
	return redis.Int(c.Do(ctx, "LINSERT", key, op, pivot, value))
}

// LLen get the length of a list
func (c *Client) LLen(ctx context.Context, key string) (int, error) {
	return redis.Int(c.Do(ctx, "LLEN", key))
}

// LPop remove and get the first element in a list
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	return redis.String(c.Do(ctx, "LPOP", key))
}

// LPush prepend one or multiple values to a list
func (c *Client) LPush(ctx context.Context, key string, values ...string) (int, error) {
	args := make([]interface{}, 1+len(values))
	args[0] = key
	for i, value := range values {
		args[1+i] = value
	}
	return redis.Int(c.Do(ctx, "LPUSH", args...))
}

// LPushX prepend a value to a list, only if the list exists
func (c *Client) LPushX(ctx context.Context, key string, value string) (int, error) {
	return redis.Int(c.Do(ctx, "LPUSHX", key, value))
}

// LRange get a range of elements from a list
func (c *Client) LRange(ctx context.Context, key string, start, stop int) ([]string, error) {
	return redis.Strings(c.Do(ctx, "LRANGE", key, start, stop))
}

// LRem remove elements from a list
func (c *Client) LRem(ctx context.Context, key string, count int, value string) (int, error) {
	return redis.Int(c.Do(ctx, "LREM", key, count, value))
}

// LSet set the value of an element in a list by its index
func (c *Client) LSet(ctx context.Context, key string, index int, value string) (string, error) {
	return redis.String(c.Do(ctx, "LSET", key, index, value))
}

// LTrim trim a list to the specified range
func (c *Client) LTrim(ctx context.Context, key string, start int, stop int) (string, error) {
	return redis.String(c.Do(ctx, "LTRIM", key, start, stop))
}

// RPop remove and get the last element in a list
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	return redis.String(c.Do(ctx, "RPOP", key))
}

// RPopLPush remove the last element in a list, prepend it to another list and return it
func (c *Client) RPopLPush(ctx context.Context, source, destination string) (string, error) {
	return redis.String(c.Do(ctx, "RPOPLPUSH", source, destination))
}

// RPush append one or multiple values to a list
func (c *Client) RPush(ctx context.Context, key string, values ...string) (int, error) {
	args := make([]interface{}, 1+len(values))
	args[0] = key
	for i, value := range values {
		args[1+i] = value
	}
	return redis.Int(c.Do(ctx, "RPUSH", args...))
}

// RPushX append a value to a list, only if the list exists
func (c *Client) RPushX(ctx context.Context, key string, value string) (int, error) {
	return redis.Int(c.Do(ctx, "RPUSHX", key, value))
}
