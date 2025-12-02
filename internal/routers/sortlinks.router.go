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
	{	
		shortlinks.POST("/links",middleware.RateLimitMiddleware(5, 5*time.Minute) ,shortlinkController.CreateShortlink)
		shortlinks.GET("/links", shortlinkController.GetAllShortlinks)
		shortlinks.GET("/links/:shortCode", shortlinkController.GetShortlinkByCode)
		shortlinks.PUT("/links/:shortCode", shortlinkController.UpdateShortlink)
		shortlinks.DELETE("/links/:shortCode", shortlinkController.DeleteShortlink)
		shortlinks.GET("/dashboard/stats", shortlinkController.GetDashboardStats )
	}
	
	r.GET("/:shortCode", shortlinkController.GetShortlinksRedis)

}
