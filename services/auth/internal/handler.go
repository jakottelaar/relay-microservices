package internal

import (
	"net/http"

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

func extractSessionMetadata(c *gin.Context) SessionMetadata {
	return SessionMetadata{
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	}
}