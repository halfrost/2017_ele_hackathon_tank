package redis

import (
	"context"

	"github.com/garyburd/redigo/redis"
)

// SAdd add one or more members to a set
func (c *Client) SAdd(ctx context.Context, key string, members ...string) (int, error) {
	args := make([]interface{}, 1+len(members))
	args[0] = key
	for i, member := range members {
		args[1+i] = member
	}
	return redis.Int(c.Do(ctx, "SADD", args...))
}

// SCard get the number of members in a set
func (c *Client) SCard(ctx context.Context, key string) (int, error) {
	return redis.Int(c.Do(ctx, "SCARD", key))
}

// SDiff subtract multiple sets
func (c *Client) SDiff(ctx context.Context, keys ...string) ([]string, error) {
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	return redis.Strings(c.Do(ctx, "SDIFF", args...))
}

// SDiffStore subtract multiple sets and store the resulting set in a key
func (c *Client) SDiffStore(ctx context.Context, destination string, keys ...string) (int, error) {
	args := make([]interface{}, 1+len(keys))
	args[0] = destination
	for i, key := range keys {
		args[1+i] = key
	}
	return redis.Int(c.Do(ctx, "SDIFFSTORE", args...))
}

// SInter intersect multiple sets
func (c *Client) SInter(ctx context.Context, keys ...string) ([]string, error) {
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	return redis.Strings(c.Do(ctx, "SINTER", args...))
}

// SInterStore intersect multiple sets and store the resulting set in a key
func (c *Client) SInterStore(ctx context.Context, destination string, keys ...string) (int, error) {
	args := make([]interface{}, 1+len(keys))
	args[0] = destination
	for i, key := range keys {
		args[1+i] = key
	}
	return redis.Int(c.Do(ctx, "SINTERSTORE", args...))
}

// SIsMember determine if a given value is a member of a set
func (c *Client) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	return redis.Bool(c.Do(ctx, "SISMEMBER", key, member))
}

// SMembers get all the members in a set
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return redis.Strings(c.Do(ctx, "SMEMBERS", key))
}

// SMove move a member from one set to another
func (c *Client) SMove(ctx context.Context, source, destination, member string) (bool, error) {
	return redis.Bool(c.Do(ctx, "SMOVE", source, destination, member))
}

// SPop remove and return one random member from a set
func (c *Client) SPop(ctx context.Context, key string) (string, error) {
	return redis.String(c.Do(ctx, "SPOP", key))
}

// SPopN remove and return multiple random members from a set
func (c *Client) SPopN(ctx context.Context, key string, count int) ([]string, error) {
	return redis.Strings(c.Do(ctx, "SPOP", key, count))
}

// SRandMember get one random members from a set
func (c *Client) SRandMember(ctx context.Context, key string) (string, error) {
	return redis.String(c.Do(ctx, "SRANDMEMBER", key))
}

// SRandMemberN get multiple random members from a set
func (c *Client) SRandMemberN(ctx context.Context, key string, count int) ([]string, error) {
	return redis.Strings(c.Do(ctx, "SRANDMEMBER", key, count))
}

// SRem remove one or more members from a set
func (c *Client) SRem(ctx context.Context, key string, members ...string) (int, error) {
	args := make([]interface{}, 1+len(members))
	args[0] = key
	for i, member := range members {
		args[1+i] = member
	}
	return redis.Int(c.Do(ctx, "SREM", args...))
}

// SUnion add multiple sets
func (c *Client) SUnion(ctx context.Context, keys ...string) ([]string, error) {
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	return redis.Strings(c.Do(ctx, "SUNION", args...))
}

// SUnionStore add multiple sets and store the resulting set in a key
func (c *Client) SUnionStore(ctx context.Context, destination string, keys ...string) (int, error) {
	args := make([]interface{}, 1+len(keys))
	args[0] = destination
	for i, key := range keys {
		args[1+i] = key
	}
	return redis.Int(c.Do(ctx, "SUNIONSTORE", args...))
}
