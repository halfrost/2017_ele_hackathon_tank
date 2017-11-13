package redis

import (
	"context"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Append append a value to a key
func (c *Client) Append(ctx context.Context, key string, value string) (int, error) {
	return redis.Int(c.Do(ctx, "APPEND", key, value))
}

// BitCount count set bits in a string
func (c *Client) BitCount(ctx context.Context, key string, poses ...int) (int, error) {
	args := make([]interface{}, 1+len(poses))
	args[0] = key
	switch len(poses) {
	case 0, 1:
	case 2:
		args[1] = poses[0]
		args[2] = poses[1]
	default:
		panic("too many arguments")
	}
	return redis.Int(c.Do(ctx, "BITCOUNT", args...))
}

// BitPos find first bit set or clear in a string
func (c *Client) BitPos(ctx context.Context, key string, bit int64, poses ...int64) (int, error) {
	args := make([]interface{}, 2+len(poses))
	args[0] = key
	args[1] = bit
	switch len(poses) {
	case 0:
	case 1:
		args[2] = poses[0]
	case 2:
		args[2] = poses[0]
		args[3] = poses[1]
	default:
		panic("too many arguments")
	}
	return redis.Int(c.Do(ctx, "BITPOS", args...))
}

// Decr decrement the integer value of a key by the given number
func (c *Client) Decr(ctx context.Context, key string) (int, error) {
	return redis.Int(c.Do(ctx, "DECR", key))
}

// DecrBy decrement the integer value of a key by the given number
func (c *Client) DecrBy(ctx context.Context, key string, decrement int) (int, error) {
	return redis.Int(c.Do(ctx, "DECRBY", key, decrement))
}

// Get get the value of a key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return redis.String(c.Do(ctx, "GET", key))
}

// GetBit returns the bit value at offset in the string value sorted at key
func (c *Client) GetBit(ctx context.Context, key string, offset int) (int, error) {
	return redis.Int(c.Do(ctx, "GETBIT", key, offset))
}

// GetRange get a substring of the string sorted at a key
func (c *Client) GetRange(ctx context.Context, key string, start int, end int) (string, error) {
	return redis.String(c.Do(ctx, "GETRANGE", key, start, end))
}

// GetSet set the string value of a key and return its old value
func (c *Client) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	return redis.String(c.Do(ctx, "GETSET", key, value))
}

// Incr increment the integer value of a key by one
func (c *Client) Incr(ctx context.Context, key string) (int, error) {
	return redis.Int(c.Do(ctx, "INCR", key))
}

// IncrBy increment the integer value of a key by the given amount
func (c *Client) IncrBy(ctx context.Context, key string, increment int) (int, error) {
	return redis.Int(c.Do(ctx, "INCRBY", key, increment))
}

// IncrByFloat increment the float value of a key by the given amount
func (c *Client) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) {
	return redis.Float64(c.Do(ctx, "INCRBYFLOAT", key, value))
}

// MGet get the values of all the given keys
func (c *Client) MGet(ctx context.Context, keys ...string) ([]string, error) {
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	return redis.Strings(c.Do(ctx, "MGET", args...))
}

// MSet set multiple keys to multiple values
func (c *Client) MSet(ctx context.Context, pairs ...string) (string, error) {
	args := make([]interface{}, len(pairs))
	for i, pair := range pairs {
		args[i] = pair
	}
	return redis.String(c.Do(ctx, "MSET", args...))
}

// Set set the string values of a key, support SET, SETEX, PSETEX commands.
// The negative expiration indicates never expires.
func (c *Client) Set(ctx context.Context, key string, value string, expiration time.Duration) (string, error) {
	args := make([]interface{}, 2, 4)
	args[0] = key
	args[1] = value
	if expiration > 0 {
		if usePrecise(expiration) {
			args = append(args, "PX", formatMs(expiration))
		} else {
			args = append(args, "EX", formatSec(expiration))
		}
	}
	return redis.String(c.Do(ctx, "SET", args...))
}

// SetBit sets or clears the bit at offset in the string value stored at key
func (c *Client) SetBit(ctx context.Context, key string, offset int, value int) (int, error) {
	return redis.Int(c.Do(ctx, "SETBIT", key, offset, value))
}

// SetNX set the value of a key, only if the key does not exist
func (c *Client) SetNX(ctx context.Context, key string, value string) (bool, error) {
	return redis.Bool(c.Do(ctx, "SETNX", key, value))
}

// SetRange overwrite part of a string at key starting at the specified offset
func (c *Client) SetRange(ctx context.Context, key string, offset int, value string) (int, error) {
	return redis.Int(c.Do(ctx, "SETRANGE", key, offset, value))
}

// StrLen get the length of the value stored in a key
func (c *Client) StrLen(ctx context.Context, key string) (int, error) {
	return redis.Int(c.Do(ctx, "STRLEN", key))
}
