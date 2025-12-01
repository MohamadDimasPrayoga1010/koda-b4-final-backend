package main

import (
	"koda-shortlink/internal/config"
	"koda-shortlink/internal/routers"

	"github.com/joho/godotenv"
)


func main() {
	godotenv.Load()
	pg := config.InitDbConfig()
	r := routers.InitRouter(pg)

	r.Run(":8082")
}