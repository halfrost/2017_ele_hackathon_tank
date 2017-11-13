package redis

import (
	"context"

	"github.com/garyburd/redigo/redis"
)

// Z represents sorted sets member
type Z struct {
	Score  float64
	Member interface{}
}

func (c *Client) zAdd(ctx context.Context, args []interface{}, n int, members []Z) (int, error) {
	for i, m := range members {
		args[n+2*i] = m.Score
		args[n+2*i+1] = m.Member
	}
	return redis.Int(c.Do(ctx, "ZADD", args...))
}

// ZAdd add one or more members to a sorted set.
func (c *Client) ZAdd(ctx context.Context, key string, members ...Z) (int, error) {
	args := make([]interface{}, 1+2*len(members))
	args[0] = key
	return c.zAdd(ctx, args, 1, members)
}

// ZAddNX with option NX, always add new elements.
func (c *Client) ZAddNX(ctx context.Context, key string, members ...Z) (int, error) {
	args := make([]interface{}, 2+2*len(members))
	args[0] = key
	args[1] = "NX"
	return c.zAdd(ctx, args, 2, members)
}

// ZAddXX with option XX, only update elements that already exists.
func (c *Client) ZAddXX(ctx context.Context, key string, members ...Z) (int, error) {
	args := make([]interface{}, 2+2*len(members))
	args[0] = key
	args[1] = "XX"
	return c.zAdd(ctx, args, 2, members)
}

// ZAddNXCh with options NX and CH, return the total number of elements changed.
func (c *Client) ZAddNXCh(ctx context.Context, key string, members ...Z) (int, error) {
	args := make([]interface{}, 3+2*len(members))
	args[0] = key
	args[1] = "NX"
	args[2] = "CH"
	return c.zAdd(ctx, args, 3, members)
}

// ZAddXXCh with options XX and CH, return the total number of elements changed.
func (c *Client) ZAddXXCh(ctx context.Context, key string, members ...Z) (int, error) {
	args := make([]interface{}, 3+2*len(members))
	args[0] = key
	args[1] = "XX"
	args[2] = "CH"
	return c.zAdd(ctx, args, 3, members)
}

func (c *Client) zIncr(ctx context.Context, args []interface{}, n int, member Z) (int, error) {
	args[n] = member.Score
	args[n+1] = member.Member
	return redis.Int(c.Do(ctx, "ZADD", args...))
}

// ZIncr like ZINCRBY.
func (c *Client) ZIncr(ctx context.Context, key string, member Z) (int, error) {
	args := make([]interface{}, 2+2)
	args[0] = key
	args[1] = "INCR"
	return c.zIncr(ctx, args, 2, member)
}

// ZIncrNX like ZINCRBY and with option NX.
func (c *Client) ZIncrNX(ctx context.Context, key string, member Z) (int, error) {
	args := make([]interface{}, 3+2)
	args[0] = key
	args[1] = "NX"
	args[2] = "INCR"
	return c.zIncr(ctx, args, 3, member)
}

// ZIncrXX like ZINCRBY and with option XX.
func (c *Client) ZIncrXX(ctx context.Context, key string, member Z) (int, error) {
	args := make([]interface{}, 3+2)
	args[0] = key
	args[1] = "XX"
	args[2] = "INCR"
	return c.zIncr(ctx, args, 3, member)
}

// ZCard get the number of members in a sorted set
func (c *Client) ZCard(ctx context.Context, key string) (int, error) {
	return redis.Int(c.Do(ctx, "ZCARD", key))
}

// ZCount count the members in a sorted set with scores within the given values
func (c *Client) ZCount(ctx context.Context, key, min, max string) (int, error) {
	return redis.Int(c.Do(ctx, "ZCOUNT", key, min, max))
}

// ZIncrBy increment the score of a member in a sorted set
func (c *Client) ZIncrBy(ctx context.Context, key, increment, member string) (string, error) {
	return redis.String(c.Do(ctx, "ZINCRBY", key, increment, member))
}

// ZStore is used as an arg to ZInterStore and ZUnionStore.
type ZStore struct {
	Weights []float64
	// Can be SUM, MIN or MAX.
	Aggregate string
}

// ZInterStore intersect multiple sorted sets and store the resulting sorted set in a new key
func (c *Client) ZInterStore(ctx context.Context, destination string, store ZStore, keys ...string) (int, error) {
	args := make([]interface{}, 2+len(keys))
	args[0] = destination
	args[1] = len(keys)
	for i, key := range keys {
		args[2+i] = key
	}
	if len(store.Weights) > 0 {
		args = append(args, "WEIGHTS")
		for _, weight := range store.Weights {
			args = append(args, weight)
		}
	}
	if store.Aggregate != "" {
		args = append(args, "AGGREGATE", store.Aggregate)
	}
	return redis.Int(c.Do(ctx, "ZINTERSTORE", args...))
}

// ZLexCount count the number of members in a sorted set between a given lexicographical range
func (c *Client) ZLexCount(ctx context.Context, key, min, max string) (int, error) {
	return redis.Int(c.Do(ctx, "ZLEXCOUNT", key, min, max))
}

func (c *Client) zRange(ctx context.Context, key string, start, stop int, withScores bool) ([]string, error) {
	args := []interface{}{key, start, stop}
	if withScores {
		args = append(args, "WITHSCORES")
	}
	return redis.Strings(c.Do(ctx, "ZRANGE", args...))
}

// ZRange return a range of members in a sorted set, by index
func (c *Client) ZRange(ctx context.Context, key string, start, stop int) ([]string, error) {
	return c.zRange(ctx, key, start, stop, false)
}

// ZRangeWithScores given the WITHSCORES option
func (c *Client) ZRangeWithScores(ctx context.Context, key string, start, stop int) ([]string, error) {
	return c.zRange(ctx, key, start, stop, true)
}

// ZRangeByOption contains the options for ZRangeBy command.
type ZRangeByOption struct {
	Min, Max      string
	Offset, Count int64
}

func (c *Client) zRangeBy(ctx context.Context, zcmd, key string, opt *ZRangeByOption, withScores bool) ([]string, error) {
	args := []interface{}{key, opt.Min, opt.Max}
	if withScores {
		args = append(args, "WITHSCORES")
	}
	if opt.Offset != 0 || opt.Count != 0 {
		args = append(args, "LIMIT", opt.Offset, opt.Count)
	}
	return redis.Strings(c.Do(ctx, zcmd, args...))
}

// ZRangeByLex return a range of members in a sorted set, by lexicographical range
func (c *Client) ZRangeByLex(ctx context.Context, key string, opt ZRangeByOption) ([]string, error) {
	return c.zRangeBy(ctx, "ZRANGEBYLEX", key, &opt, false)
}

// ZRangeByScore return a range of members in a sorted set, by score
func (c *Client) ZRangeByScore(ctx context.Context, key string, opt ZRangeByOption) ([]string, error) {
	return c.zRangeBy(ctx, "ZRANGEBYSCORE", key, &opt, false)
}

// ZRangeByScoreWithScores return a range of members in a sorted set, by score
func (c *Client) ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeByOption) ([]string, error) {
	return c.zRangeBy(ctx, "ZRANGEBYSCORE", key, &opt, true)
}

// ZRank determine the index of a member in a sorted set
func (c *Client) ZRank(ctx context.Context, key string, member string) (int, error) {
	return redis.Int(c.Do(ctx, "ZRANK", key, member))
}

// ZRem remove one or more members from a sorted set
func (c *Client) ZRem(ctx context.Context, key string, members ...string) (int, error) {
	args := make([]interface{}, 1+len(members))
	args[0] = key
	for i, member := range members {
		args[1+i] = member
	}
	return redis.Int(c.Do(ctx, "ZREM", args...))
}

// ZRemRangeByLex remove all members in a sorted set between the given lexicographical range
func (c *Client) ZRemRangeByLex(ctx context.Context, key, min, max string) (int, error) {
	return redis.Int(c.Do(ctx, "ZREMRANGEBYLEX", key, min, max))
}

// ZRemRangeByRank remove all members in a sorted set within the given indexes
func (c *Client) ZRemRangeByRank(ctx context.Context, key string, start, stop int) (int, error) {
	return redis.Int(c.Do(ctx, "ZREMRANGEBYRANK", key, start, stop))
}

// ZRemRangeByScore remove all members in a sorted set within the given scores
func (c *Client) ZRemRangeByScore(ctx context.Context, key, min, max string) (int, error) {
	return redis.Int(c.Do(ctx, "ZREMRANGEBYSCORE", key, min, max))
}

// ZRevRange return a range of members in a sorted set, by index, with scores ordered from high to low
func (c *Client) ZRevRange(ctx context.Context, key, start, stop string) ([]string, error) {
	return redis.Strings(c.Do(ctx, "ZREVRANGE", key, start, stop))
}

// ZRevRangeWithScores return a range of members in a sorted set, by index, with scores ordered from high to low
func (c *Client) ZRevRangeWithScores(ctx context.Context, key, start, stop string) ([]string, error) {
	return redis.Strings(c.Do(ctx, "ZREVRANGE", key, start, stop, "WITHSCORES"))
}

func (c *Client) zRevRangeBy(ctx context.Context, zcmd, key string, opt *ZRangeByOption, withScores bool) ([]string, error) {
	args := []interface{}{key, opt.Min, opt.Max}

	if withScores {
		args = append(args, "WITHSCORES")
	}

	if opt.Offset != 0 || opt.Count != 0 {
		args = append(args, "LIMIT", opt.Offset, opt.Count)
	}
	return redis.Strings(c.Do(ctx, zcmd, args...))
}

// ZRevRangeByLex return a range of members in a sorted set, by lexicographical range, ordered from higher to lower strings.
func (c *Client) ZRevRangeByLex(ctx context.Context, key string, opt ZRangeByOption) ([]string, error) {
	return c.zRevRangeBy(ctx, "ZREVRANGEBYLEX", key, &opt, false)
}

// ZRevRangeByScore return a range of members in a sorted set, by score, with scores ordered from high to low
func (c *Client) ZRevRangeByScore(ctx context.Context, key string, opt ZRangeByOption) ([]string, error) {
	return c.zRevRangeBy(ctx, "ZREVRANGEBYSCORE", key, &opt, false)
}

// ZRevRangeByScoreWithScores return a range of members in a sorted set, by score, with scores ordered from high to low
func (c *Client) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeByOption) ([]string, error) {
	return c.zRevRangeBy(ctx, "ZREVRANGEBYSCORE", key, &opt, true)
}

// ZRevRank determine the index of a member in a sorted set, with scores ordered from high to low
func (c *Client) ZRevRank(ctx context.Context, key string, member string) (int, error) {
	return redis.Int(c.Do(ctx, "ZREVRANK", key, member))
}

// ZScore get the score associated with the given member in a sorted set
func (c *Client) ZScore(ctx context.Context, key string, member string) (string, error) {
	return redis.String(c.Do(ctx, "ZSCORE", key, member))
}

// ZUnionStore add multiple sorted sets and store the resulting sorted set in a new key
func (c *Client) ZUnionStore(ctx context.Context, destination string, store ZStore, keys ...string) (int, error) {
	args := make([]interface{}, 2+len(keys))
	args[0] = destination
	args[1] = len(keys)
	for i, key := range keys {
		args[2+i] = key
	}
	if len(store.Weights) > 0 {
		args = append(args, "WEIGHTS")
		for _, weight := range store.Weights {
			args = append(args, weight)
		}
	}
	if store.Aggregate != "" {
		args = append(args, "AGGREGATE", store.Aggregate)
	}
	return redis.Int(c.Do(ctx, "ZUNIONSTORE", args...))
}
