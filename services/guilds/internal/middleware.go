package internal

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jakottelaar/relay-microservices/shared/errors"
)

func UserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			c.Error(errors.NewUnauthorizedError("missing user ID"))
			c.Abort()
			return
		}
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.Error(errors.NewUnauthorizedError("invalid user ID"))
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}