package main

import (
	"fmt"
	"log"

	"user-service/internal/bootstrap"
	"user-service/internal/handler"
	jwtpkg "user-service/internal/pkg/jwt"
	"user-service/internal/repository"
	"user-service/internal/router"
	"user-service/internal/service"
)

func main() {
	app, err := bootstrap.Initialize("./config")
	if err != nil {
		log.Fatalf("initialize app failed: %v", err)
	}

	jwtManager := jwtpkg.NewManager(app.Config.JWT.Secret, app.Config.JWT.Expire)
	userRepo := repository.NewUserRepository(app.DB)
	authService := service.NewAuthService(userRepo, jwtManager)
	userHandler := handler.NewUserHandler(authService)

	r := router.New(userHandler, jwtManager)
	addr := fmt.Sprintf(":%d", app.Config.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("start server failed: %v", err)
	}
}
