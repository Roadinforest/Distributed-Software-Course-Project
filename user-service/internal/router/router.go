package router

import (
	"net/http"

	"user-service/internal/handler"
	"user-service/internal/middleware"
	jwtpkg "user-service/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func New(userHandler *handler.UserHandler, jwtManager *jwtpkg.Manager) *gin.Engine {
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	users := api.Group("/users")
	{
		users.POST("/register", userHandler.Register)
		users.POST("/login", userHandler.Login)

		authed := users.Group("")
		authed.Use(middleware.JWTAuth(jwtManager))
		{
			authed.GET("/profile", userHandler.Profile)
			authed.PUT("/profile", userHandler.UpdateProfile)
		}
	}

	return r
}
