package middleware

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupCORS() gin.HandlerFunc {
	origin := os.Getenv("ALLOW_ORIGIN")
	
	config := cors.Config{
		AllowOrigins:     []string{origin, "http://localhost:5173"},
		AllowMethods:     []string{"GET","PATCH", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}

	return cors.New(config)
}
