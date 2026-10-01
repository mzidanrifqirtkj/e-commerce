package config

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

// NewRedisClient returns the shared Redis client, building it on first use.
//
// redis.NewClient only allocates a connection pool, it does not dial, so a
// single client is reused for the whole process instead of one pool per call.
// Commands fail with an error when Redis is unreachable, which the callers
// already handle, so there is no connect-time Ping here.
//
// ponytail: address is hardcoded; move to REDIS_HOST/REDIS_PORT config when
// more than one environment has to point somewhere else.
func NewRedisClient() *redis.Client {
	redisOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6380",
		})
	})

	return redisClient
}
