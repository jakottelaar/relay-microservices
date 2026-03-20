package internal

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
    service AuthService
    log     *zap.Logger
}

func NewAuthHandler(service AuthService, log *zap.Logger) *AuthHandler {
    return &AuthHandler{
        service: service,
        log:     log,
    }
}

func (h *AuthHandler) SignUp(c *gin.Context) {
    var req SignUpRequest
    if err := c.BindJSON(&req); err != nil {
        h.log.Warn("Invalid request body",
            zap.Error(err),
            zap.String("path", c.Request.URL.Path),
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    h.log.Info("Sign-up attempt",
        zap.String("email", req.Email),
        zap.String("ip", c.ClientIP()),
    )

    metadata := extractSessionMetadata(c)
    authResp, err := h.service.SignUp(c.Request.Context(), req, metadata)
    if err != nil {
        h.log.Error("Sign-up failed",
            zap.Error(err),
            zap.String("email", req.Email),
        )
        _ = c.Error(err)
        return
    }

    h.log.Info("Sign-up successful",
        zap.String("email", req.Email),
        zap.Int64("account_id", authResp.Account.ID),
    )

    c.JSON(http.StatusCreated, authResp)
}

func (h *AuthHandler) SignIn(c *gin.Context) {
    var req SignInRequest
    if err := c.BindJSON(&req); err != nil {
        h.log.Warn("Invalid request body",
            zap.Error(err),
            zap.String("path", c.Request.URL.Path),
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    metadata := extractSessionMetadata(c)
    authResp, err := h.service.SignIn(c.Request.Context(), req, metadata)
    if err != nil {
        h.log.Error("Sign-in failed",
            zap.Error(err),
        )
        _ = c.Error(err)
        return
    }

    c.JSON(http.StatusOK, authResp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
    var req RefreshTokenRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    metadata := extractSessionMetadata(c)
    authResp, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken, metadata)
    if err != nil {
        _ = c.Error(err)
        return
    }

    c.JSON(http.StatusOK, authResp)
}

func (h *AuthHandler) SignOut(c *gin.Context) {
    var req RefreshTokenRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    err := h.service.SignOut(c.Request.Context(), req.RefreshToken)
    if err != nil {
        _ = c.Error(err)
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) GetSessionById(c *gin.Context) {
    sessionId := c.Param("id")
    id, err := strconv.ParseInt(sessionId, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
        return
    }

    session, err := h.service.GetSessionByID(c.Request.Context(), id)
    if err != nil {
        _ = c.Error(err)
        return
    }

    c.JSON(http.StatusOK, session)
}

func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
        return
    }

    err := h.service.RevokeAllSessions(c.Request.Context(), userID.(int64))
    if err != nil {
        _ = c.Error(err)
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "All sessions revoked successfully"})
}

func (h *AuthHandler) RevokeSessionById(c *gin.Context) {
    sessionId := c.Param("id")
    id, err := strconv.ParseInt(sessionId, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID"})
        return
    }

    err = h.service.RevokeSessionById(c.Request.Context(), id)
    if err != nil {
        _ = c.Error(err)
        return
    }

    c.JSON(http.StatusNoContent, nil)
}

func extractSessionMetadata(c *gin.Context) SessionMetadata {
    return SessionMetadata{
        UserAgent: c.GetHeader("User-Agent"),
        IPAddress: c.ClientIP(),
    }
}