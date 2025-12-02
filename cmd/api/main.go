package main

import (
	"koda-shortlink/internal/config"
	"koda-shortlink/internal/routers"
	"koda-shortlink/internal/utils"

	_ "koda-shortlink/docs"

	"github.com/joho/godotenv"
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

	utils.InitRedis()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8082")
}