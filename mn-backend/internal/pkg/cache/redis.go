package cache

import (
	"context"
	"strings"

	"moonick/internal/config"

	redis "github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client    *redis.Client
	keyPrefix string
}

var globalRedis *RedisClient

func NewRedisClient(cfg config.RedisConfig) *RedisClient {
	return &RedisClient{
		client:    redis.NewClient(buildRedisOptions(cfg)),
		keyPrefix: strings.TrimSpace(cfg.KeyPrefix),
	}
}

func InitRedis(cfg config.RedisConfig) *RedisClient {
	globalRedis = NewRedisClient(cfg)
	return globalRedis
}

func SetRedis(client *RedisClient) {
	globalRedis = client
}

func GetRedis() *RedisClient {
	return globalRedis
}

func buildRedisOptions(cfg config.RedisConfig) *redis.Options {
	return &redis.Options{
		Addr:         strings.TrimSpace(cfg.Addr),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
}

func (c *RedisClient) Raw() *redis.Client {
	if c == nil {
		return nil
	}
	return c.client
}

func (c *RedisClient) PrefixKey(key string) string {
	if c == nil {
		return key
	}
	return c.keyPrefix + key
}

func (c *RedisClient) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Ping(ctx).Err()
}

func (c *RedisClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}
