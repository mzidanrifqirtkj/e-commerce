package config

import (
	"net"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

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
// The address comes from REDIS_HOST/REDIS_PORT, which the process is expected
// to have already loaded. A missing pair is fatal here rather than silent: the
// resulting ":" address would otherwise only surface as a misleading 401 from
// the token middleware on the first request.
func NewRedisClient() *redis.Client {
	redisOnce.Do(func() {
		host := viper.GetString("REDIS_HOST")
		port := viper.GetString("REDIS_PORT")

		if host == "" || port == "" {
			log.Fatal().Msg("[NewRedisClient-1] REDIS_HOST/REDIS_PORT is not set")
		}

		redisClient = redis.NewClient(&redis.Options{
			Addr: net.JoinHostPort(host, port),
		})
	})

	return redisClient
}
