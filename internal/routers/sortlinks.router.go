package routers

import (
	"koda-shortlink/internal/handler"
	"koda-shortlink/internal/middleware"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ShortlinkRoutes(r *gin.Engine, pg *pgxpool.Pool) {
	shortlinkController := handler.ShortlinkController{DB: pg}

	shortlinks := r.Group("/api/v1")
	opt := shortlinks.Group(("/"))
	opt.Use(middleware.OptAuthMiddleware(""))
	opt.POST("/links",middleware.RateLimitMiddleware(5, 5*time.Minute) ,shortlinkController.CreateShortlink)
	{
		shortlinks.GET("/links", middleware.AuthMiddleware(""),shortlinkController.GetAllShortlinks)
		shortlinks.GET("/links/:shortCode", middleware.AuthMiddleware(""),shortlinkController.GetShortlinkByCode)
		shortlinks.PUT("/links/:shortCode",middleware.AuthMiddleware("") ,shortlinkController.UpdateShortlink)
		shortlinks.DELETE("/links/:shortCode", middleware.AuthMiddleware(""),shortlinkController.DeleteShortlink)
		shortlinks.GET("/dashboard/stats", middleware.AuthMiddleware(""),shortlinkController.GetDashboardStats )
	}
	
	r.GET("/:shortCode", shortlinkController.GetShortlinksRedis)

}
