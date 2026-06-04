package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
)

type UserHandler struct {
	service UserService
}

func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userID, err := sonyflake.ParseID(c.Param("id"))
    if err != nil {
        c.Error(err)
        return
    }

	user, err := h.service.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}