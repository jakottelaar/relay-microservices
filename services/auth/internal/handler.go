package internal

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
    service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
    return &AuthHandler{
        service: service,
    }
}

func (h *AuthHandler) SignUp(c *gin.Context) {
    var req SignUpRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    metadata := extractSessionMetadata(c)
    authResp, err := h.service.SignUp(c.Request.Context(), req, metadata)
    if err != nil {
        _ = c.Error(err)
        return
    }

    c.JSON(http.StatusCreated, authResp)
}

func (h *AuthHandler) SignIn(c *gin.Context) {
    var req SignInRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    metadata := extractSessionMetadata(c)
    authResp, err := h.service.SignIn(c.Request.Context(), req, metadata)
    if err != nil {
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

// Validate is the endpoint Traefik ForwardAuth calls
func (h *AuthHandler) Validate(c *gin.Context) {
    c.Status(http.StatusOK)
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

func extractSessionMetadata(c *gin.Context) SessionMetadata {
    return SessionMetadata{
        UserAgent: c.GetHeader("User-Agent"),
        IPAddress: c.ClientIP(),
    }
}