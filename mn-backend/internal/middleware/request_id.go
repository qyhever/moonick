package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"moonick/internal/controller"
	jwtpkg "moonick/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(requestIDHeader, requestID)
		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Next()
	}
}

func RequireUserAuth(manager *jwtpkg.Manager) gin.HandlerFunc {
	return requireAuth(manager, "user")
}

func RequireAdminAuth(manager *jwtpkg.Manager) gin.HandlerFunc {
	return requireAuth(manager, "admin")
}

func requireAuth(manager *jwtpkg.Manager, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if manager == nil || manager.ConfigError() != nil {
			abortServerError(c)
			return
		}

		token, err := jwtpkg.ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			abortNeedLogin(c)
			return
		}

		claims, err := manager.ValidateAccessToken(token, requiredRole)
		if err != nil {
			switch err {
			case jwtpkg.ErrInvalidTokenRole:
				abortForbidden(c)
			default:
				abortInvalidToken(c)
			}
			return
		}

		c.Set(jwtpkg.ContextClaimsKey, claims)
		c.Next()
	}
}

func abortNeedLogin(c *gin.Context) {
	controller.ResponseFailedWithMsg(c, controller.CodeNeedLogin, controller.CodeNeedLogin.Msg())
	c.Abort()
}

func abortForbidden(c *gin.Context) {
	controller.ResponseFailedWithMsg(c, controller.CodePermissionDenied, controller.CodePermissionDenied.Msg())
	c.Abort()
}

func abortInvalidToken(c *gin.Context) {
	controller.ResponseFailedWithMsg(c, controller.CodeInvalidToken, controller.CodeInvalidToken.Msg())
	c.Abort()
}

func abortServerError(c *gin.Context) {
	controller.ResponseFailedWithMsg(c, controller.CodeServerBusy, "jwt configuration invalid")
	c.Abort()
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "request-id-unavailable"
	}
	return hex.EncodeToString(buf)
}
