package redis

import (
	"context"

	"github.com/garyburd/redigo/redis"
)

// HDel delete one or more hash fields
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int, error) {
	args := make([]interface{}, 1+len(fields))
	args[0] = key
	for i, field := range fields {
		args[1+i] = field
	}
	return redis.Int(c.Do(ctx, "HDEL", args...))
}

// HExists determine if a hash field exists
func (c *Client) HExists(ctx context.Context, key string, field string) (bool, error) {
	return redis.Bool(c.Do(ctx, "HEXISTS", key, field))
}

// HGet get the value of a hash field
func (c *Client) HGet(ctx context.Context, key string, field string) (string, error) {
	return redis.String(c.Do(ctx, "HGET", key, field))
}

// HGetAll get all the fields and values in a hash
func (c *Client) HGetAll(ctx context.Context, key string) ([]string, error) {
	return redis.Strings(c.Do(ctx, "HGETALL", key))
}

// HGetAllMap get all the fields and values in a hash and return as map
func (c *Client) HGetAllMap(ctx context.Context, key string) (map[string]string, error) {
	allSlice, err := c.HGetAll(ctx, key)
	if err != nil {
		return nil, err
	}
	allMap := make(map[string]string, len(allSlice)/2)
	var k string
	for i, value := range allSlice {
		if i%2 == 0 {
			k = value
		} else {
			allMap[k] = value
		}
	}
	return allMap, nil
}

// HIncrBy increment the integer value of a hash field by the given number
func (c *Client) HIncrBy(ctx context.Context, key string, field string, increment int) (int, error) {
	return redis.Int(c.Do(ctx, "HINCRBY", key, field, increment))
}

// HIncrByFloat increment the float value of a hash field by the given amount
func (c *Client) HIncrByFloat(ctx context.Context, key string, field string, increment float64) (float64, error) {
	return redis.Float64(c.Do(ctx, "HINCRBYFLOAT", key, field, increment))
}

// HKeys get all the fields in a hash
func (c *Client) HKeys(ctx context.Context, key string) ([]string, error) {
	return redis.Strings(c.Do(ctx, "HKEYS", key))
}

// HLen get the number of fields in a hash
func (c *Client) HLen(ctx context.Context, key string) (int, error) {
	return redis.Int(c.Do(ctx, "HLEN", key))
}

// HMGet the values of all the given hash fields
func (c *Client) HMGet(ctx context.Context, key string, fields ...string) ([]string, error) {
	args := make([]interface{}, 1+len(fields))
	args[0] = key
	for i, field := range fields {
		args[1+i] = field
	}
	return redis.Strings(c.Do(ctx, "HMGET", args...))
}

// HMSet set multiple hash fields to multiple values
func (c *Client) HMSet(ctx context.Context, key string, field string, value string, pairs ...string) (string, error) {
	args := make([]interface{}, 3+len(pairs))
	args[0] = key
	args[1] = field
	args[2] = value
	for i, pair := range pairs {
		args[3+i] = pair
	}
	return redis.String(c.Do(ctx, "HMSET", args...))
}

// HMSetMap set hash fields map to multiple values
func (c *Client) HMSetMap(ctx context.Context, key string, fields map[string]string) (string, error) {
	args := make([]interface{}, 1+len(fields)*2)
	args[0] = key
	i := 1
	for k, v := range fields {
		args[i] = k
		args[i+1] = v
		i += 2
	}
	return redis.String(c.Do(ctx, "HMSET", args...))
}

// HSet set the string value of a hash field
func (c *Client) HSet(ctx context.Context, key string, field string, value string) (bool, error) {
	return redis.Bool(c.Do(ctx, "HSET", key, field, value))
}

// HSetNX set the value of a hash field, only if the field does not exist
func (c *Client) HSetNX(ctx context.Context, key string, field string, value string) (bool, error) {
	return redis.Bool(c.Do(ctx, "HSETNX", key, field, value))
}

// HStrLen get the length of the value of a hash field
func (c *Client) HStrLen(ctx context.Context, key string, field string) (int, error) {
	return redis.Int(c.Do(ctx, "HSTRLEN", key, field))
}

// HVals get all the values in a hash
func (c *Client) HVals(ctx context.Context, key string) ([]string, error) {
	return redis.Strings(c.Do(ctx, "HVALS", key))
}
