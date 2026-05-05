package cache

import (
	"testing"
	"time"

	"moonick/internal/config"
)

func TestBuildRedisOptionsFromConfig(t *testing.T) {
	cfg := config.RedisConfig{
		Addr:         "127.0.0.1:6379",
		Password:     "secret",
		DB:           2,
		PoolSize:     16,
		MinIdleConns: 4,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 4 * time.Second,
	}

	opts := buildRedisOptions(cfg)

	if opts.Addr != cfg.Addr {
		t.Fatalf("addr = %q, want %q", opts.Addr, cfg.Addr)
	}
	if opts.Password != cfg.Password {
		t.Fatalf("password = %q, want %q", opts.Password, cfg.Password)
	}
	if opts.DB != cfg.DB {
		t.Fatalf("db = %d, want %d", opts.DB, cfg.DB)
	}
	if opts.PoolSize != cfg.PoolSize {
		t.Fatalf("pool size = %d, want %d", opts.PoolSize, cfg.PoolSize)
	}
	if opts.MinIdleConns != cfg.MinIdleConns {
		t.Fatalf("min idle conns = %d, want %d", opts.MinIdleConns, cfg.MinIdleConns)
	}
	if opts.DialTimeout != cfg.DialTimeout {
		t.Fatalf("dial timeout = %s, want %s", opts.DialTimeout, cfg.DialTimeout)
	}
	if opts.ReadTimeout != cfg.ReadTimeout {
		t.Fatalf("read timeout = %s, want %s", opts.ReadTimeout, cfg.ReadTimeout)
	}
	if opts.WriteTimeout != cfg.WriteTimeout {
		t.Fatalf("write timeout = %s, want %s", opts.WriteTimeout, cfg.WriteTimeout)
	}
}

func TestPrefixKey(t *testing.T) {
	client := &RedisClient{
		keyPrefix: "moonick:",
	}

	if got := client.PrefixKey("auth:login:203.0.113.10"); got != "moonick:auth:login:203.0.113.10" {
		t.Fatalf("prefixed key = %q", got)
	}

	if got := client.PrefixKey(""); got != "moonick:" {
		t.Fatalf("empty key prefixed = %q", got)
	}
}
