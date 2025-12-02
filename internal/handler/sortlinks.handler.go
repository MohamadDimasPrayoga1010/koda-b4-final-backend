package handler

import (
	"context"
	"encoding/json"
	"koda-shortlink/internal/models"
	"koda-shortlink/internal/utils"
	"koda-shortlink/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShortlinkController struct {
	DB *pgxpool.Pool
}

type CreateShortlinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
}

// @Summary Create a new shortlink
// @Description Generate a shortlink for the provided URL
// @Tags Shortlinks
// @Accept json
// @Produce json
// @Param body body CreateShortlinkRequest true "Shortlink creation payload"
// @Success 201 {object} response.Response "Returns the created shortlink data"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/links [post]
func (sc *ShortlinkController) CreateShortlink(ctx *gin.Context) {
	var req CreateShortlinkRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, response.Response{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	shortCode := utils.GenerateShortCode(6)

	sl := models.Shortlink{
		OriginalURL: req.OriginalURL,
		ShortCode:   shortCode,
	}

	newSL, err := models.CreateShortlink(sc.DB, sl)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to create shortlink: " + err.Error(),
		})
		return
	}
	ctx.JSON(201, response.Response{
		Success: true,
		Message: "Shortlink created successfully",
		Data: gin.H{
			"id":           newSL.ID,
			"original_url": newSL.OriginalURL,
			"short_code":   newSL.ShortCode,
			"created_at":   newSL.CreatedAt,
		},
	})
}

// @Summary Get all shortlinks
// @Description Retrieve a list of all shortlinks
// @Tags Shortlinks
// @Produce json
// @Success 200 {object} response.Response "Returns list of shortlinks"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/links [get]
func (sc *ShortlinkController) GetAllShortlinks(ctx *gin.Context) {
	shortlinks, err := models.GetAllShortlinks(sc.DB)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to fetch shortlinks: " + err.Error(),
		})
		return
	}

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Shortlinks retrieved successfully",
		Data:    shortlinks,
	})
}

// @Summary Redirect shortlink by code
// @Description Redirects the user to the original URL based on the short code. Also logs the click and increments redirect count.
// @Tags Shortlinks
// @Accept json
// @Produce json
// @Param shortCode path string true "Shortlink code"
// @Success 302 {string} string "Redirects to the original URL"
// @Failure 404 {object} response.Response "Shortlink not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/links/{shortCode} [get]
func (sc *ShortlinkController) GetShortlinkByCode(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")

	sl, err := models.GetShortlinkByCode(sc.DB, shortCode)
	if err != nil {
		ctx.JSON(404, response.Response{
			Success: false,
			Message: "Shortlink not found",
		})
		return
	}

	if err := models.IncrementRedirectCount(sc.DB, sl.ID); err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to increment redirect count",
		})
		return
	}

	click := models.ShortlinkClick{
		ShortlinkID: sl.ID,
		IP:          ctx.ClientIP(),
		UserAgent:   ctx.Request.UserAgent(),
	}
	_ = models.LogClick(sc.DB, click)

	ctx.Redirect(302, sl.OriginalURL)
}

type UpdateShortlinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
	ShortCode   string `json:"short_code" binding:"omitempty,max=10"`
}

// UpdateShortlink godoc
// @Summary Update shortlink
// @Description Update original URL or generate/set new short code
// @Tags Shortlinks
// @Accept json
// @Produce json
// @Param shortCode path string true "Existing short code"
// @Param body body UpdateShortlinkRequest true "Update shortlink payload"
// @Success 200 {object} response.Response "Shortlink updated successfully"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 404 {object} response.Response "Shortlink not found"
// @Failure 409 {object} response.Response "Short code already in use"
// @Failure 500 {object} response.Response "Failed to update shortlink"
// @Router /api/v1/links/{shortCode} [put]
func (sc *ShortlinkController) UpdateShortlink(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")
	var req UpdateShortlinkRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, response.Response{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	sl, err := models.GetShortlinkByCode(sc.DB, shortCode)
	if err != nil {
		ctx.JSON(404, response.Response{
			Success: false,
			Message: "Shortlink not found",
		})
		return
	}

	if req.OriginalURL != "" {
		sl.OriginalURL = req.OriginalURL
	}

	if req.ShortCode == "" {
		sl.ShortCode = utils.GenerateShortCode(6)

	} else {

		if !utils.IsValidShortCode(req.ShortCode) {
			ctx.JSON(400, response.Response{
				Success: false,
				Message: "Short code must be alphanumeric only",
			})
			return
		}

		exists, _ := models.CheckShortCodeExists(sc.DB, req.ShortCode)
		if exists && req.ShortCode != sl.ShortCode {
			ctx.JSON(409, response.Response{
				Success: false,
				Message: "Short code is already in use",
			})
			return
		}

		sl.ShortCode = req.ShortCode
	}

	updatedSL, err := models.UpdateShortlink(sc.DB, sl)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to update shortlink",
		})
		return
	}

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Shortlink updated successfully",
		Data:    updatedSL,
	})
}

// DeleteShortlink godoc
// @Summary Delete a shortlink
// @Description Delete shortlink by its short code
// @Tags Shortlinks
// @Produce json
// @Param shortCode path string true "Short code to delete"
// @Success 200 {object} response.Response "Shortlink deleted successfully"
// @Failure 404 {object} response.Response "Shortlink not found"
// @Failure 500 {object} response.Response "Failed to delete shortlink"
// @Router /api/v1/links/{shortCode} [delete]
func (sc *ShortlinkController) DeleteShortlink(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")

	_, err := models.GetShortlinkByCode(sc.DB, shortCode)
	if err != nil {
		ctx.JSON(404, response.Response{
			Success: false,
			Message: "Shortlink not found",
		})
		return
	}

	err = models.DeleteShortlink(sc.DB, shortCode)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to delete shortlink",
		})
		return
	}

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Shortlink deleted successfully",
	})
}

// @Summary Resolve shortlink to original URL
// @Description Resolve shortlink: hit Redis first, then DB fallback.
// @Description Click counter is incremented in Redis. Analytics logged asynchronously.
// @Tags Redirect
// @Produce json
// @Param shortCode path string true "Short code"
// @Success 200 {object} response.Response "Original URL returned successfully"
// @Failure 404 {object} response.Response "Shortlink not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /{shortCode} [get]
func (sc *ShortlinkController) GetShortlinksRedis(ctx *gin.Context) {
	shortCode := ctx.Param("shortCode")
	rctx := context.Background()

	destKey := "link:" + shortCode + ":destination"
	clickKey := "link:" + shortCode + ":clicks"

	var sl models.Shortlink

	val, err := utils.RedisClient.Get(rctx, destKey).Result()
	if err == nil {
		_ = json.Unmarshal([]byte(val), &sl)
	} else {
		sl, err = models.GetShortlinkByCode(sc.DB, shortCode)
		if err != nil {
			ctx.JSON(404, response.Response{
				Success: false,
				Message: "Shortlink not found",
			})
			return
		}

		jsonData, _ := json.Marshal(sl)
		utils.RedisClient.Set(rctx, destKey, jsonData, 24*time.Hour)
	}
	utils.RedisClient.Incr(rctx, clickKey)

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Shortlink resolved",
		Data: gin.H{
			"original_url": sl.OriginalURL,
		},
	})

	go models.IncrementRedirectCount(sc.DB, sl.ID)

	go models.LogClick(sc.DB, models.ShortlinkClick{
		ShortlinkID: sl.ID,
		IP:          ctx.ClientIP(),
		UserAgent:   ctx.Request.UserAgent(),
	})
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Retrieve overall shortlink statistics including total links, total visits, average click rate, visits growth, and last 7 days visitor chart
// @Tags Dashboard
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}} "Returns dashboard statistics"
// @Failure 500 {object} response.Response "Failed to retrieve dashboard stats"
// @Router /api/v1/dashboard/stats [get]
func (sc *ShortlinkController) GetDashboardStats(ctx *gin.Context) {
	rctx := context.Background()
	dashboardCacheKey := "analytics:global:7d"

	val, err := utils.RedisClient.Get(rctx, dashboardCacheKey).Result()
	if err == nil && val != "" {
		var stats models.DashboardStats
		if err := json.Unmarshal([]byte(val), &stats); err == nil {
			ctx.JSON(200, response.Response{
				Success: true,
				Message: "Dashboard stats retrieved successfully (from cache)",
				Data: gin.H{
					"total_links":    stats.TotalLinks,
					"total_visits":   stats.TotalVisits,
					"avg_click_rate": stats.AvgClickRate,
					"visits_growth":  stats.VisitsGrowth,
					"last_7_days":    stats.Last7Days,
				},
			})
			return
		}
	}

	stats, err := models.GetDashboardStats(sc.DB)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to retrieve dashboard stats",
		})
		return
	}

	jsonData, _ := json.Marshal(stats)
	utils.RedisClient.Set(rctx, dashboardCacheKey, jsonData, time.Hour)

	for _, sl := range stats.Last7DaysShortlinks {
		destKey := "link:" + sl.ShortCode + ":destination"
		jsonSL, _ := json.Marshal(sl)
		utils.RedisClient.Set(rctx, destKey, jsonSL, 24*time.Hour)
	}

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Dashboard stats retrieved successfully",
		Data: gin.H{
			"total_links":    stats.TotalLinks,
			"total_visits":   stats.TotalVisits,
			"avg_click_rate": stats.AvgClickRate,
			"visits_growth":  stats.VisitsGrowth,
			"last_7_days":    stats.Last7Days,
		},
	})

}
