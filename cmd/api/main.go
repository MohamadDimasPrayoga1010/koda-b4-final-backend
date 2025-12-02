package main

import (
	"koda-shortlink/internal/config"
	"koda-shortlink/internal/routers"

	"github.com/joho/godotenv"
	_ "koda-shortlink/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)


// @title Koda Shortlink API
// @version 1.0
// @description API documentation for Koda Shortlink
// @host localhost:8082
// @BasePath /api/v1
func main() {
	godotenv.Load()
	pg := config.InitDbConfig()
	r := routers.InitRouter(pg)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8082")
}