package internal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jakottelaar/relay-microservices/services/auth/config"
)

var (
    ErrInvalidToken = errors.New("invalid token")
    ErrExpiredToken = errors.New("token has expired")
)

const (
    AuthorizationHeader = "Authorization"
    BearerPrefix        = "Bearer "
    UserIDKey           = "user_id"
    SessionIDKey        = "session_id"
    
    // Headers that Traefik will forward to services
    HeaderUserID    = "X-User-ID"
    HeaderSessionID = "X-Session-ID"
)

type Claims struct {
    UserID    int64  `json:"user_id"`
    SessionID int64  `json:"session_id"`
    jwt.RegisteredClaims
}

type JWTManager interface {
    GenerateToken(userID int64, sessionID int64) (string, error)
    ValidateToken(tokenString string) (*Claims, error)
}

type jwtManager struct {
    secret []byte
    cfg    *config.Config
    repo   *AuthRepository
}

func NewJWTManager(cfg *config.Config, repo *AuthRepository) (*jwtManager, error) {
    if cfg.JWTSecret == "" {
        return nil, errors.New("JWT secret is required")
    }

    return &jwtManager{
        secret: []byte(cfg.JWTSecret),
        cfg:    cfg,
        repo:   repo,
    }, nil
}

func (m *jwtManager) GenerateToken(userID int64, sessionID int64) (string, error) {
    claims := Claims{
        UserID:    userID,
        SessionID: sessionID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.cfg.AccessTokenExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    m.cfg.JWTIssuer,
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secret)
}

func (m *jwtManager) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return m.secret, nil
        },
    )

    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }

    if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
        return nil, ErrExpiredToken
    }

    return claims, nil
}

// Used by the /auth/validate endpoint for Traefik ForwardAuth
func ValidateMiddleware(jwtManager *jwtManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader(AuthorizationHeader)
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        if !strings.HasPrefix(authHeader, BearerPrefix) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
            c.Abort()
            return
        }

        token := strings.TrimPrefix(authHeader, BearerPrefix)

        claims, err := jwtManager.ValidateToken(token)
        if err != nil {
            if errors.Is(err, ErrExpiredToken) {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
            } else {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            }
            c.Abort()
            return
        }

        // Check if session is still valid
        session, err := jwtManager.repo.Queries.GetSessionByID(c.Request.Context(), claims.SessionID)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})
            c.Abort()
            return
        }

        if session.RevokedAt.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Session revoked"})
            c.Abort()
            return
        }

        if session.ExpiresAt.Time.Before(time.Now()) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
            c.Abort()
            return
        }

        // Update last_used_at
        _ = jwtManager.repo.Queries.UpdateSessionLastUsed(c.Request.Context(), claims.SessionID)

        c.Header(HeaderUserID, fmt.Sprintf("%d", claims.UserID))
        c.Header(HeaderSessionID, fmt.Sprintf("%d", claims.SessionID))

        c.Next()
    }
}

func RequireAuth(jwtManager *jwtManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader(AuthorizationHeader)
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        if !strings.HasPrefix(authHeader, BearerPrefix) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
            c.Abort()
            return
        }

        token := strings.TrimPrefix(authHeader, BearerPrefix)

        claims, err := jwtManager.ValidateToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        // Check session validity
        session, err := jwtManager.repo.Queries.GetSessionByID(c.Request.Context(), claims.SessionID)
        if err != nil || session.RevokedAt.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Session revoked or invalid"})
            c.Abort()
            return
        }

        c.Set(UserIDKey, claims.UserID)
        c.Set(SessionIDKey, claims.SessionID)

        c.Next()
    }
}

func GenerateRefreshToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("failed to generate refresh token: %w", err)
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

func HashRefreshToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return base64.URLEncoding.EncodeToString(hash[:])
}

func GetUserID(c *gin.Context) (int64, bool) {
    userID, exists := c.Get(UserIDKey)
    if !exists {
        return 0, false
    }
    id, ok := userID.(int64)
    return id, ok
}