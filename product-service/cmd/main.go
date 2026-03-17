package main

import (
	"fmt"
	"log"
	"os"

	"product-service/internal/bootstrap"
	"product-service/internal/handler"
	"product-service/internal/repository"
	"product-service/internal/router"
	"product-service/internal/service"
)

func main() {
	app, err := bootstrap.Initialize("./config")
	if err != nil {
		log.Fatalf("initialize app failed: %v", err)
	}

	productRepo := repository.NewProductRepository(app.DB)
	productService := service.NewProductService(productRepo, app.Redis, app.Config)
	productHandler := handler.NewProductHandler(productService)

	r := router.New(productHandler)
	addr := fmt.Sprintf(":%d", app.Config.Server.Port)
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "product-service"
	}
	log.Printf("starting server instance=%s addr=%s", instanceID, addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("start server failed: %v", err)
	}
}
