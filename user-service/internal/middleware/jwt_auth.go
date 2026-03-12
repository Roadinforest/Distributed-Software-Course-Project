package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	jwtpkg "user-service/internal/pkg/jwt"
	"user-service/internal/pkg/response"
)

const userIDContextKey = "user_id"

func JWTAuth(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.JSON(c, http.StatusUnauthorized, "missing authorization header", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.JSON(c, http.StatusUnauthorized, "invalid authorization header", nil)
			c.Abort()
			return
		}

		claims, err := jwtManager.Parse(parts[1])
		if err != nil {
			response.JSON(c, http.StatusUnauthorized, "invalid token", nil)
			c.Abort()
			return
		}

		c.Set(userIDContextKey, claims.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(userIDContextKey)
	if !exists {
		return 0, false
	}

	id, ok := value.(uint64)
	return id, ok
}


