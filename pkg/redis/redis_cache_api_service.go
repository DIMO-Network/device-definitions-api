//go:generate mockgen -source redis_cache_api_service.go -destination mocks/redis_cache_api_service_mock.go -package mocks

package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/go-redis/redis/v8"
)

// CacheService combines methods of redis client and redis clustered client to have one impl that works for both, reduced to only what we use
type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	FlushAll(ctx context.Context) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Close() error
}

// NewRedisCacheService establishes connection to Redis and creates client. db is the 0-16 db instance to use from redis.
func NewRedisCacheService(settings *config.Settings, db int) CacheService {
	var tlsConfig *tls.Config
	if settings.RedisTLS {
		tlsConfig = new(tls.Config)
	}

	var r CacheService
	// handle redis cluster in prod
	if settings.Environment == "prod" { // || settings.Environment == "dev"
		cc := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:     []string{settings.RedisURL},
			Password:  settings.RedisPassword,
			TLSConfig: tlsConfig,
		})
		cc.Do(context.Background(), fmt.Sprintf("SELECT %d", db)) // todo check that this works in prod
		r = cc
	} else {
		c := redis.NewClient(&redis.Options{
			Addr:      settings.RedisURL,
			Password:  settings.RedisPassword,
			TLSConfig: tlsConfig,
			DB:        db,
		})
		r = c
	}

	return r
}
