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
	now       func() time.Time
	mu        sync.Mutex
	lastByKey map[string]time.Time
}

func NewIPRateLimit(window time.Duration) gin.HandlerFunc {
	return newIPRateLimit(window, time.Now)
}

func newIPRateLimit(window time.Duration, now func() time.Time) gin.HandlerFunc {
	limiter := &ipRateLimit{
		window:    window,
		now:       now,
		lastByKey: make(map[string]time.Time),
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

	l.mu.Lock()
	defer l.mu.Unlock()

	if last, ok := l.lastByKey[key]; ok && now.Sub(last) < l.window {
		return false
	}

	l.lastByKey[key] = now
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
