package main

// @title           API Koda Shortlink Documentation
// @version         1.0
// @description     Dokumentasi REST API menggunakan Gin dan Swagger

// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
import (
	"koda-shortlink/internal/config"
	"koda-shortlink/internal/routers"
	"koda-shortlink/internal/utils"

	_ "koda-shortlink/docs"

	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	godotenv.Load()
	pg := config.InitDbConfig()
	r := routers.InitRouter(pg)

	utils.InitRedis()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8082")
}