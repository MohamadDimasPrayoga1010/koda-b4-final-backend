package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"koda-shortlink/internal/models"
	"koda-shortlink/internal/utils"
	"koda-shortlink/pkg/response"
	"math"
	"strconv"
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
// @Description Generate a shortlink for the provided URL (works with or without authentication)
// @Tags Shortlinks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateShortlinkRequest true "Shortlink creation payload"
// @Success 201 {object} response.Response "Returns the created shortlink data"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/links [post]
func (sc *ShortlinkController) CreateShortlink(ctx *gin.Context) {
	var req CreateShortlinkRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	if !utils.ValidateURL(req.OriginalURL) {
		ctx.JSON(400, gin.H{
			"success": false,
			"message": "URL is not valid or unsupported",
		})
		return
	}

	shortCode := utils.GenerateShortCode(6)

	var uid *int64
	if userIDValue, exists := ctx.Get("userID"); exists {
		switch v := userIDValue.(type) {
		case int64:
			uid = &v
		case int:
			id := int64(v)
			uid = &id
		case float64:
			id := int64(v)
			uid = &id
		}
	}
	fmt.Println(uid)

	sl := models.Shortlink{
		OriginalURL: req.OriginalURL,
		ShortCode:   shortCode,
		UserID:      uid,
	}

	newSL, err := models.CreateShortlink(sc.DB, sl)
	if err != nil {
		ctx.JSON(500, gin.H{
			"success": false,
			"message": "Failed to create shortlink: " + err.Error(),
		})
		return
	}

	if uid != nil {
		rctx := context.Background()
		dashboardKey := fmt.Sprintf("analytics:user:%d:7d", *uid)
		utils.RedisClient.Del(rctx, dashboardKey)
	}

	ctx.JSON(201, gin.H{
		"success": true,
		"message": "Shortlink created successfully",
		"data": gin.H{
			"id":           newSL.ID,
			"original_url": newSL.OriginalURL,
			"short_code":   newSL.ShortCode,
			"status":       newSL.Status,
			"created_at":   newSL.CreatedAt,
		},
	})
}

// @Summary Get all shortlinks
// @Description Retrieve a list of all shortlinks for authenticated user
// @Tags Shortlinks
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "Returns list of shortlinks"
// @Failure 401 {object} response.Response "User not authenticated"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/links [get]
func (sc *ShortlinkController) GetAllShortlinks(ctx *gin.Context) {
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, response.Response{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	var userID int64
	switch v := userIDValue.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case float64:
		userID = int64(v)
	default:
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Invalid user ID type",
		})
		return
	}

	limitQuery := ctx.DefaultQuery("limit", "10")
	pageQuery := ctx.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitQuery)
	if err != nil || limit <= 0 {
		limit = 10
	}

	page, err := strconv.Atoi(pageQuery)
	if err != nil || page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	shortlinks, total, err := models.GetAllShortlinks(sc.DB, userID, limit, offset)
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
		Data: gin.H{
			"items": shortlinks,
			"pagination": gin.H{
				"total": total,
				"limit": limit,
				"page":  page,
				"pages": int(math.Ceil(float64(total) / float64(limit))),
				"next":  page*limit < total,
				"back":  page > 1,
			},
		},
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

	if sl.Status == "inactive" {
		ctx.JSON(403, response.Response{
			Success: false,
			Message: "This shortlink is currently inactive",
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
	OriginalURL string `json:"originalUrl" binding:"required"`
	ShortCode   string `json:"shortCode"`
	Status      string `json:"status"`
}

// UpdateShortlink godoc
// @Summary Update shortlink
// @Description Update original URL or generate/set new short code (requires authentication)
// @Tags Shortlinks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param shortCode path string true "Existing short code"
// @Param body body UpdateShortlinkRequest true "Update shortlink payload"
// @Success 200 {object} response.Response "Shortlink updated successfully"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 401 {object} response.Response "User not authenticated"
// @Failure 403 {object} response.Response "No permission to update this link"
// @Failure 404 {object} response.Response "Shortlink not found"
// @Failure 409 {object} response.Response "Short code already in use"
// @Failure 500 {object} response.Response "Failed to update shortlink"
// @Router /api/v1/links/{shortCode} [put]
func (sc *ShortlinkController) UpdateShortlink(ctx *gin.Context) {
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, response.Response{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	var userID int64
	switch v := userIDValue.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case float64:
		userID = int64(v)
	}

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

	if sl.UserID == nil || *sl.UserID != userID {
		ctx.JSON(403, response.Response{
			Success: false,
			Message: "You don't have permission to update this link",
		})
		return
	}

	if req.OriginalURL != "" {
		sl.OriginalURL = req.OriginalURL
	}

	if req.Status != "" {
		sl.Status = req.Status
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

	rctx := context.Background()
	destKey := "link:" + shortCode + ":destination"
	dashboardKey := fmt.Sprintf("analytics:user:%d:7d", userID)

	utils.RedisClient.Del(rctx, destKey)
	utils.RedisClient.Del(rctx, dashboardKey)
	utils.RedisClient.Del(rctx, "analytics:global:7d")

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Shortlink updated successfully",
		Data:    updatedSL,
	})
}

// DeleteShortlink godoc
// @Summary Delete a shortlink
// @Description Delete shortlink by its short code (requires authentication)
// @Tags Shortlinks
// @Produce json
// @Security BearerAuth
// @Param shortCode path string true "Short code to delete"
// @Success 200 {object} response.Response "Shortlink deleted successfully"
// @Failure 401 {object} response.Response "User not authenticated"
// @Failure 403 {object} response.Response "No permission to delete this link"
// @Failure 404 {object} response.Response "Shortlink not found"
// @Failure 500 {object} response.Response "Failed to delete shortlink"
// @Router /api/v1/links/{shortCode} [delete]
func (sc *ShortlinkController) DeleteShortlink(ctx *gin.Context) {
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, response.Response{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	var userID int64
	switch v := userIDValue.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case float64:
		userID = int64(v)
	}

	shortCode := ctx.Param("shortCode")

	sl, err := models.GetShortlinkByCode(sc.DB, shortCode)
	if err != nil {
		ctx.JSON(404, response.Response{
			Success: false,
			Message: "Shortlink not found",
		})
		return
	}

	if sl.UserID == nil || *sl.UserID != userID {
		ctx.JSON(403, response.Response{
			Success: false,
			Message: "You don't have permission to delete this link",
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

	rctx := context.Background()
	destKey := "link:" + shortCode + ":destination"
	dashboardKey := fmt.Sprintf("analytics:user:%d:7d", userID)

	utils.RedisClient.Del(rctx, destKey)
	utils.RedisClient.Del(rctx, dashboardKey)
	utils.RedisClient.Del(rctx, "analytics:global:7d")

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

	var sl models.Shortlink

	val, err := utils.RedisClient.Get(rctx, destKey).Result()
	if err == nil && val != "" {
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

	if sl.Status == "inactive" {
		ctx.JSON(403, response.Response{
			Success: false,
			Message: "This shortlink is currently inactive",
		})
		return
	}

	if userIDValue, exists := ctx.Get("userID"); exists {
		var userID int64
		switch v := userIDValue.(type) {
		case int64:
			userID = v
		case int:
			userID = int64(v)
		case float64:
			userID = int64(v)
		}

		clickKey := fmt.Sprintf("link:%s:clicks:user:%d", shortCode, userID)
		if err := utils.RedisClient.Incr(rctx, clickKey).Err(); err != nil {
			fmt.Println("Redis Incr error:", err)
		}
	}

	ctx.Redirect(302, sl.OriginalURL)

	go func() {
		if err := models.IncrementRedirectCount(sc.DB, sl.ID); err == nil {
			rctx := context.Background()
			if sl.UserID != nil {
				dashboardKey := fmt.Sprintf("analytics:user:%d:7d", *sl.UserID)
				utils.RedisClient.Del(rctx, dashboardKey)
			}

			utils.RedisClient.Del(rctx, "analytics:global:7d")
		}
		_ = models.LogClick(sc.DB, models.ShortlinkClick{
			ShortlinkID: sl.ID,
			IP:          ctx.ClientIP(),
			UserAgent:   ctx.Request.UserAgent(),
		})
	}()
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Retrieve overall shortlink statistics for authenticated user
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=map[string]interface{}} "Returns dashboard statistics"
// @Failure 401 {object} response.Response "User not authenticated"
// @Failure 500 {object} response.Response "Failed to retrieve dashboard stats"
// @Router /api/v1/dashboard/stats [get]
func (sc *ShortlinkController) GetDashboardStats(ctx *gin.Context) {
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, response.Response{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}


	var userID int64
	switch v := userIDValue.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case float64:
		userID = int64(v)
	default:
		ctx.JSON(400, response.Response{
			Success: false,
			Message: "Invalid userID type",
		})
		return
	}

	rctx := context.Background()
	dashboardCacheKey := fmt.Sprintf("analytics:user:%d:7d", userID)

	val, err := utils.RedisClient.Get(rctx, dashboardCacheKey).Result()
	if err == nil && val != "" {
		var stats models.DashboardStats
		if err := json.Unmarshal([]byte(val), &stats); err == nil {
			ctx.JSON(200, response.Response{
				Success: true,
				Message: "Dashboard stats retrieved successfully (from cache)",
				Data: gin.H{
					"totalLinks":   stats.TotalLinks,
					"totalVisits":  stats.TotalVisits,
					"avgClickRate": stats.AvgClickRate,
					"visitsGrowth": stats.VisitsGrowth,
					"last7Days":    stats.Last7Days,
				},
			})
			return
		}
	}

	stats, err := models.GetDashboardStatsByUser(sc.DB, int(userID))
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to retrieve dashboard stats",
		})
		return
	}

	jsonData, _ := json.Marshal(stats)
	utils.RedisClient.Set(rctx, dashboardCacheKey, jsonData, time.Hour)

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Dashboard stats retrieved successfully",
		Data: gin.H{
			"totalLinks":   stats.TotalLinks,
			"totalVisits":  stats.TotalVisits,
			"avgClickRate": stats.AvgClickRate,
			"visitsGrowth": stats.VisitsGrowth,
			"last7Days":    stats.Last7Days,
		},
	})

}
