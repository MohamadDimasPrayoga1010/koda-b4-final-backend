package routers

import (
	"koda-shortlink/internal/handler"
	"koda-shortlink/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func UserRoutes(r *gin.Engine, pg *pgxpool.Pool) {
	profileController := handler.ProfileController{DB: pg}

	user := r.Group("/api/v1")
	{
		user.GET("/profile", middleware.AuthMiddleware(""), profileController.GetProfile)
		user.PATCH("/profile", middleware.AuthMiddleware(""), profileController.UpdateProfile)
	}

}
