package handler

import (
	"net/http"

	"koda-shortlink/internal/config"
	"koda-shortlink/internal/routers"
	"koda-shortlink/internal/utils"

	_ "koda-shortlink/docs"

	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	godotenv.Load()

	pg := config.InitDbConfig()
	router := routers.InitRouter(pg)

	utils.InitRedis()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.ServeHTTP(w, r)
}
