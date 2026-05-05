package middleware

import (
	"net"
	"strings"
	"sync"
	"time"

	"moonick/internal/controller"

	"github.com/gin-gonic/gin"
)

const rateLimitMessage = "请勿频繁操作"

type ipRateLimit struct {
	window    time.Duration
	limit     int
	now       func() time.Time
	mu        sync.Mutex
	hitsByKey map[string][]time.Time
}

func NewIPRateLimit(window time.Duration, limit int) gin.HandlerFunc {
	return newIPRateLimit(window, limit, time.Now)
}

func newIPRateLimit(window time.Duration, limit int, now func() time.Time) gin.HandlerFunc {
	limiter := &ipRateLimit{
		window:    window,
		limit:     limit,
		now:       now,
		hitsByKey: make(map[string][]time.Time),
	}

	return limiter.handle
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
	now := l.now()
	key := route + "|" + ip
	windowStart := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	hits := l.hitsByKey[key]
	kept := hits[:0]
	for _, hitAt := range hits {
		if !hitAt.Before(windowStart) {
			kept = append(kept, hitAt)
		}
	}

	if len(kept) >= l.limit {
		l.hitsByKey[key] = kept
		return false
	}

	l.hitsByKey[key] = append(kept, now)
	return true
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
