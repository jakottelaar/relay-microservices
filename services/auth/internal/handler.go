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

	var req *SignUpRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.service.SignUp(c.Request.Context(), *req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, &SignUpResponse{
		AccessToken: "",
		RefreshToken: "",
		Account: account,
	})
}