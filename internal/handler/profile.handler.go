package handler

import (
	"context"
	"encoding/json"
	"koda-shortlink/internal/models"
	"koda-shortlink/internal/utils"
	"koda-shortlink/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileController struct {
	DB *pgxpool.Pool
}

// GetProfile godoc
// @Summary Get user profile
// @Description Retrieve user profile with image, fullname, email and user-specific stats, with Redis caching
// @Tags Profile
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}} "Returns user profile and stats"
// @Failure 500 {object} response.Response "Failed to retrieve profile or stats"
// @Router /api/v1/profile [get]
func (pc *ProfileController) GetProfile(ctx *gin.Context) {
	rctx := context.Background()
	userID := ctx.GetInt("userID")
	userIDStr := strconv.Itoa(userID)

	profileCacheKey := "user:" + userIDStr + ":profile"
	statsCacheKey := "user:" + userIDStr + ":stats"

	val, err := utils.RedisClient.Get(rctx, profileCacheKey).Result()
	var profile models.UserProfileResponse
	if err == nil && val != "" {
		if err := json.Unmarshal([]byte(val), &profile); err == nil {
			statsVal, _ := utils.RedisClient.Get(rctx, statsCacheKey).Result()
			var stats models.DashboardStats
			if statsVal != "" {
				_ = json.Unmarshal([]byte(statsVal), &stats)
			}

			ctx.JSON(200, response.Response{
				Success: true,
				Message: "Profile retrieved successfully (from cache)",
				Data: gin.H{
					"profile": profile,
					"stats":   stats,
				},
			})
			return
		}
	}

	profile, err = models.GetUserProfile(pc.DB, userID)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to retrieve profile",
		})
		return
	}

	jsonProfile, _ := json.Marshal(profile)
	_ = utils.RedisClient.Set(rctx, profileCacheKey, jsonProfile, time.Hour)

	stats, err := models.GetDashboardStatsByUser(pc.DB, userID)
	if err == nil {
		jsonStats, _ := json.Marshal(stats)
		_ = utils.RedisClient.Set(rctx, statsCacheKey, jsonStats, time.Hour)
	}

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Profile retrieved successfully",
		Data: gin.H{
			"profile": profile,
			"stats":   stats,
		},
	})

}
