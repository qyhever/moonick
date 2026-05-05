package middleware

import (
	"context"
	"log"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"moonick/internal/controller"
	"moonick/internal/pkg/cache"

	"github.com/gin-gonic/gin"
	redis "github.com/redis/go-redis/v9"
)

const rateLimitMessage = "请勿频繁操作"

var redisSlidingWindowScript = redis.NewScript(`
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", ARGV[1])
local count = redis.call("ZCARD", KEYS[1])
if count >= tonumber(ARGV[2]) then
  redis.call("EXPIRE", KEYS[1], ARGV[4])
  return 0
end
redis.call("ZADD", KEYS[1], ARGV[3], ARGV[3])
redis.call("EXPIRE", KEYS[1], ARGV[4])
return 1
`)

type rateLimitStore interface {
	Allow(key string, window time.Duration, limit int, now time.Time) (bool, error)
}

type ipRateLimit struct {
	window time.Duration
	limit  int
	now    func() time.Time
	store  rateLimitStore
}

type memoryRateLimitStore struct {
	mu        sync.Mutex
	hitsByKey map[string][]time.Time
}

type redisRateLimitStore struct {
	client *cache.RedisClient
}

func NewIPRateLimit(window time.Duration, limit int) gin.HandlerFunc {
	return newIPRateLimitWithStore(window, limit, time.Now, newRateLimitStore())
}

func newIPRateLimit(window time.Duration, limit int, now func() time.Time) gin.HandlerFunc {
	return newIPRateLimitWithStore(window, limit, now, newMemoryRateLimitStore())
}

func newIPRateLimitWithStore(window time.Duration, limit int, now func() time.Time, store rateLimitStore) gin.HandlerFunc {
	limiter := &ipRateLimit{
		window: window,
		limit:  limit,
		now:    now,
		store:  store,
	}

	return limiter.handle
}

func newRateLimitStore() rateLimitStore {
	if client := cache.GetRedis(); client != nil && client.Raw() != nil {
		return newRedisRateLimitStore(client)
	}
	return newMemoryRateLimitStore()
}

func newMemoryRateLimitStore() rateLimitStore {
	return &memoryRateLimitStore{
		hitsByKey: make(map[string][]time.Time),
	}
}

func newRedisRateLimitStore(client *cache.RedisClient) rateLimitStore {
	return &redisRateLimitStore{client: client}
}

func (l *ipRateLimit) handle(c *gin.Context) {
	ip := GetClientIP(c)
	if ip == "" {
		ip = "unknown"
	}

	if !l.allow(c.FullPath(), ip) {
		controller.ResponseFailedWithMsg(c, controller.CodeInvalidParam, rateLimitMessage)
		c.Abort()
		return
	}

	c.Next()
}

func (l *ipRateLimit) allow(route, ip string) bool {
	allowed, err := l.store.Allow(route+"|"+ip, l.window, l.limit, l.now())
	if err != nil {
		log.Printf("rate limit store failed for route=%s ip=%s: %v", route, ip, err)
		return true
	}
	return allowed
}

func (s *memoryRateLimitStore) Allow(key string, window time.Duration, limit int, now time.Time) (bool, error) {
	windowStart := now.Add(-window)

	s.mu.Lock()
	defer s.mu.Unlock()

	hits := s.hitsByKey[key]
	kept := hits[:0]
	for _, hitAt := range hits {
		if !hitAt.Before(windowStart) {
			kept = append(kept, hitAt)
		}
	}

	if len(kept) >= limit {
		s.hitsByKey[key] = kept
		return false, nil
	}

	s.hitsByKey[key] = append(kept, now)
	return true, nil
}

func (s *redisRateLimitStore) Allow(key string, window time.Duration, limit int, now time.Time) (bool, error) {
	if s == nil || s.client == nil || s.client.Raw() == nil {
		return true, nil
	}

	redisKey := s.client.PrefixKey("rate_limit:" + key)
	nowUnixNano := now.UnixNano()
	windowStartUnixNano := now.Add(-window).UnixNano()
	ttlSeconds := int(math.Ceil(window.Seconds()))
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}

	result, err := redisSlidingWindowScript.Run(
		context.Background(),
		s.client.Raw(),
		[]string{redisKey},
		windowStartUnixNano,
		limit,
		nowUnixNano,
		ttlSeconds,
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func GetClientIP(c *gin.Context) string {
	for _, candidate := range strings.Split(c.GetHeader("X-Forwarded-For"), ",") {
		if ip := normalizeIP(candidate); ip != "" {
			return ip
		}
	}

	if ip := normalizeIP(c.GetHeader("X-Real-IP")); ip != "" {
		return ip
	}

	if ip := normalizeIP(c.ClientIP()); ip != "" {
		return ip
	}

	return ""
}

func normalizeIP(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || strings.EqualFold(value, "unknown") {
		return ""
	}

	if parsed := net.ParseIP(value); parsed != nil {
		return parsed.String()
	}

	host, _, err := net.SplitHostPort(value)
	if err != nil {
		return ""
	}
	if parsed := net.ParseIP(strings.TrimSpace(host)); parsed != nil {
		return parsed.String()
	}

	return ""
}
