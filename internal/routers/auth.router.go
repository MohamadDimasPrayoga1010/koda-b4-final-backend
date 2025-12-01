package routers

import (
	"koda-shortlink/internal/handler"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthRoutes(r *gin.Engine, pg *pgxpool.Pool) {
	authController := handler.AuthController{DB: pg}

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/logout", authController.Logout)
	}
}
