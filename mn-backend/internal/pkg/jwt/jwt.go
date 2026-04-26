package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrEmptySecret       = errors.New("jwt secret is empty")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrMissingBearer     = errors.New("missing bearer token")
	ErrInvalidTokenType  = errors.New("invalid token type")
	ErrInvalidTokenRole  = errors.New("invalid token role")
	ErrInvalidAccessTTL  = errors.New("jwt access token ttl must be greater than zero")
	ErrInvalidRefreshTTL = errors.New("jwt refresh token ttl must be greater than zero")
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	ContextClaimsKey = "jwtClaims"
)

type Config struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type Claims struct {
	Subject   string `json:"sub"`
	Role      string `json:"role,omitempty"`
	TokenType string `json:"typ,omitempty"`
	ExpiresAt int64  `json:"exp"`
}

type Manager struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	configErr       error
}

func NewManager(cfg Config) *Manager {
	manager := &Manager{
		secret:          []byte(cfg.Secret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
	manager.configErr = validateConfig(cfg)
	return manager
}

func (m *Manager) GenerateAccessToken(subject, role string) (string, error) {
	return m.generateToken(subject, role, TokenTypeAccess, m.accessTokenTTL)
}

func (m *Manager) GenerateRefreshToken(subject string) (string, error) {
	return m.generateToken(subject, "", TokenTypeRefresh, m.refreshTokenTTL)
}

func (m *Manager) ValidateAccessToken(token string, requiredRole string) (*Claims, error) {
	claims, err := m.Parse(token)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(claims.TokenType, TokenTypeAccess) {
		return nil, ErrInvalidTokenType
	}
	if requiredRole != "" && !strings.EqualFold(claims.Role, requiredRole) {
		return nil, ErrInvalidTokenRole
	}
	return claims, nil
}

func (m *Manager) ConfigError() error {
	if m == nil {
		return ErrEmptySecret
	}
	return m.configErr
}

func (m *Manager) Parse(token string) (*Claims, error) {
	if err := m.ConfigError(); err != nil {
		return nil, err
	}
	if len(m.secret) == 0 {
		return nil, ErrEmptySecret
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	expectedSig := sign(payloadBytes, m.secret)
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.TokenType == "" {
		if claims.Role == "" {
			claims.TokenType = TokenTypeRefresh
		} else {
			claims.TokenType = TokenTypeAccess
		}
	}
	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, ErrTokenExpired
	}
	return &claims, nil
}

func ExtractBearerToken(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", ErrMissingBearer
	}
	return strings.TrimSpace(parts[1]), nil
}

func (m *Manager) generateToken(subject, role, tokenType string, ttl time.Duration) (string, error) {
	if err := m.ConfigError(); err != nil {
		return "", err
	}

	claims := Claims{
		Subject:   subject,
		Role:      role,
		TokenType: tokenType,
	}
	if ttl > 0 {
		claims.ExpiresAt = time.Now().Add(ttl).Unix()
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return payload + "." + sign(payloadBytes, m.secret), nil
}

func sign(payload []byte, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func validateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Secret) == "" {
		return ErrEmptySecret
	}
	if cfg.AccessTokenTTL <= 0 {
		return ErrInvalidAccessTTL
	}
	if cfg.RefreshTokenTTL <= 0 {
		return ErrInvalidRefreshTTL
	}
	return nil
}
