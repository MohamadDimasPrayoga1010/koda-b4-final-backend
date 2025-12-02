package routers

import (
"koda-shortlink/internal/handler"
"koda-shortlink/internal/middleware"


"github.com/gin-gonic/gin"
"github.com/jackc/pgx/v5/pgxpool"


)

func UserRoutes(r *gin.Engine, pg *pgxpool.Pool) {
profileController := handler.ProfileController{DB: pg}
dashboardController := handler.ShortlinkController{DB: pg}


user := r.Group("/api/v1")
{
	user.GET("/profile", middleware.AuthMiddleware(""), profileController.GetProfile)
	user.GET("/dashboard/stats", middleware.AuthMiddleware(""), dashboardController.GetDashboardStats)
}


}
