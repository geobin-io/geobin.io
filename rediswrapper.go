package main

import (
	"time"

	redis "gopkg.in/redis.v1"
)

// mock our use of redis pubsub for modularity/testing purposes
type PubSubber interface {
	Subscribe(channels ...string) error
	Unsubscribe(channels ...string) error
}

// mock our use of redis client for modularity/testing purposes
type RedisClient interface {
	ZAdd(key string, members ...redis.Z) (int64, error)
	ZCount(key, min, max string) (int64, error)
	Expire(key string, dur time.Duration) (bool, error)
	Publish(channel, message string) (int64, error)
	ZRevRange(key, start, stop string) ([]string, error)
	Exists(key string) (bool, error)
	Get(key string) (string, error)
	Incr(key string) (int64, error)
}

// wraps redis.Client as a RedisClient above, which gives a simpler interface
// for basic usage and mocking
type redisWrapper struct {
	r *redis.Client
}

func NewRedisWrapper(r *redis.Client) RedisClient {
	return &redisWrapper{r}
}

func (rw *redisWrapper) ZAdd(key string, members ...redis.Z) (int64, error) {
	return rw.r.ZAdd(key, members...).Result()
}

func (rw *redisWrapper) ZCount(key, min, max string) (int64, error) {
	return rw.r.ZCount(key, min, max).Result()
}

func (rw *redisWrapper) Expire(key string, dur time.Duration) (bool, error) {
	return rw.r.Expire(key, dur).Result()
}

func (rw *redisWrapper) Publish(channel, message string) (int64, error) {
	return rw.r.Publish(channel, message).Result()
}

func (rw *redisWrapper) ZRevRange(key, start, stop string) ([]string, error) {
	return rw.r.ZRevRange(key, start, stop).Result()
}

func (rw *redisWrapper) Exists(key string) (bool, error) {
	return rw.r.Exists(key).Result()
}

func (rw *redisWrapper) Get(key string) (string, error) {
	return rw.r.Get(key).Result()
}

func (rw *redisWrapper) Incr(key string) (int64, error) {
	return rw.r.Incr(key).Result()
}
