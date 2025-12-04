package routers

import (
	"koda-shortlink/internal/middleware"
	"koda-shortlink/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitRouter(pg *pgxpool.Pool) *gin.Engine {
	r := gin.Default()

	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, response.Response{
			Success: true,
			Message: "Backend is running well well well",
		})
	})
	r.Use(middleware.SetupCORS())
	r.Use(middleware.Logger())
	AuthRoutes(r, pg)
	ShortlinkRoutes(r, pg)
	UserRoutes(r, pg)
	return r
}