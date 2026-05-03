package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moonick/internal/controller"

	"github.com/gin-gonic/gin"
)

func TestIPRateLimitRejectsFrequentRequestsWithinWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	limiter := newIPRateLimit(5*time.Second, func() time.Time {
		return now
	})

	r := gin.New()
	r.POST("/auth/login", limiter, func(c *gin.Context) {
		controller.ResponseSuccess(c, gin.H{"ok": true})
	})

	first := performIPLimitedRequest(t, r, "/auth/login", "203.0.113.10", "")
	assertMiddlewareResponseCode(t, first, controller.CodeSuccess, "")

	second := performIPLimitedRequest(t, r, "/auth/login", "203.0.113.10", "")
	assertMiddlewareResponseCode(t, second, controller.CodeInvalidParam, "请勿频繁操作")
}

func TestIPRateLimitAllowsRequestAfterWindowElapsed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	limiter := newIPRateLimit(5*time.Second, func() time.Time {
		return now
	})

	r := gin.New()
	r.POST("/auth/login", limiter, func(c *gin.Context) {
		controller.ResponseSuccess(c, gin.H{"ok": true})
	})

	first := performIPLimitedRequest(t, r, "/auth/login", "203.0.113.10", "")
	assertMiddlewareResponseCode(t, first, controller.CodeSuccess, "")

	now = now.Add(5*time.Second + time.Millisecond)

	second := performIPLimitedRequest(t, r, "/auth/login", "203.0.113.10", "")
	assertMiddlewareResponseCode(t, second, controller.CodeSuccess, "")
}

func TestIPRateLimitSupportsDifferentWindowsPerRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	loginLimiter := newIPRateLimit(5*time.Second, func() time.Time {
		return now
	})
	registerLimiter := newIPRateLimit(10*time.Second, func() time.Time {
		return now
	})

	r := gin.New()
	r.POST("/auth/login", loginLimiter, func(c *gin.Context) {
		controller.ResponseSuccess(c, gin.H{"route": "login"})
	})
	r.POST("/auth/register", registerLimiter, func(c *gin.Context) {
		controller.ResponseSuccess(c, gin.H{"route": "register"})
	})

	assertMiddlewareResponseCode(t, performIPLimitedRequest(t, r, "/auth/login", "203.0.113.10", ""), controller.CodeSuccess, "")
	assertMiddlewareResponseCode(t, performIPLimitedRequest(t, r, "/auth/register", "203.0.113.10", ""), controller.CodeSuccess, "")

	now = now.Add(6 * time.Second)

	assertMiddlewareResponseCode(t, performIPLimitedRequest(t, r, "/auth/login", "203.0.113.10", ""), controller.CodeSuccess, "")
	assertMiddlewareResponseCode(t, performIPLimitedRequest(t, r, "/auth/register", "203.0.113.10", ""), controller.CodeInvalidParam, "请勿频繁操作")
}

func TestGetClientIPPrefersFirstForwardedAddress(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ip", func(c *gin.Context) {
		c.String(http.StatusOK, GetClientIP(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.8, 10.0.0.1")
	req.Header.Set("X-Real-IP", "198.51.100.9")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Body.String() != "198.51.100.8" {
		t.Fatalf("expected first forwarded ip, got %q", rec.Body.String())
	}
}

func TestGetClientIPFallsBackToRealIPWhenForwardedInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ip", func(c *gin.Context) {
		c.String(http.StatusOK, GetClientIP(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.Header.Set("X-Forwarded-For", "unknown, bad-value")
	req.Header.Set("X-Real-IP", "198.51.100.9")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Body.String() != "198.51.100.9" {
		t.Fatalf("expected real ip fallback, got %q", rec.Body.String())
	}
}

func performIPLimitedRequest(t *testing.T, r http.Handler, path, forwardedFor, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, nil)
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func assertMiddlewareResponseCode(t *testing.T, rec *httptest.ResponseRecorder, expectedCode controller.MyCode, expectedMessage string) {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected http status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp controller.ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v, body=%s", err, rec.Body.String())
	}
	if resp.Code != expectedCode {
		t.Fatalf("expected response code %d, got %d, body=%s", expectedCode, resp.Code, rec.Body.String())
	}
	if expectedMessage != "" && resp.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, resp.Message)
	}
}
