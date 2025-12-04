package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

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
// @Security BearerAuth
// @Success 200 {object} response.Response{data=map[string]interface{}} "Returns user profile and stats"
// @Failure 500 {object} response.Response "Failed to retrieve profile or stats"
// @Router /api/v1/profile [get]
func (pc *ProfileController) GetProfile(ctx *gin.Context) {
	rctx := context.Background()

	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, response.Response{
			Success: false,
			Message: "Unauthorized, userID not found",
		})
		return
	}

	var userID int
	switch v := userIDValue.(type) {
	case int:
		userID = v
	case int64:
		userID = int(v)
	case float64:
		userID = int(v)
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			ctx.JSON(500, response.Response{
				Success: false,
				Message: "Invalid userID string",
			})
			return
		}
		userID = parsed
	default:
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Invalid userID type",
		})
		return
	}

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

// UpdateProfile godoc
// @Summary Update user profile
// @Description Partially update user profile information (fullname, email, image in Base64 string) with Redis cache invalidation
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body object{fullname=string,email=string,image=string} false "JSON body containing fields to update"
// @Success 200 {object} response.Response{data=map[string]interface{}} "Profile updated successfully"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Failed to update profile"
// @Router /api/v1/profile [patch]
func (pc *ProfileController) UpdateProfile(ctx *gin.Context) {
	rctx := context.Background()

	userIDValue, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, response.Response{
			Success: false,
			Message: "Unauthorized, userID not found",
		})
		return
	}

	var userID int
	switch v := userIDValue.(type) {
	case int:
		userID = v
	case int64:
		userID = int(v)
	case float64:
		userID = int(v)
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.Response{
				Success: false,
				Message: "Invalid userID string",
			})
			return
		}
		userID = parsed
	default:
		ctx.JSON(http.StatusInternalServerError, response.Response{
			Success: false,
			Message: "Invalid userID type",
		})
		return
	}

	fullname := ctx.PostForm("fullname")
	email := ctx.PostForm("email")

	var image *string
	file, err := ctx.FormFile("image")
	if err == nil && file != nil {
		if file.Size > 2*1024*1024 {
			ctx.JSON(http.StatusBadRequest, response.Response{
				Success: false,
				Message: "Image size exceeds 2MB",
			})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			ctx.JSON(http.StatusBadRequest, response.Response{
				Success: false,
				Message: "Invalid image format. Only jpg, jpeg, png allowed",
			})
			return
		}

		filename := strconv.Itoa(userID) + ext
		dst := "uploads/" + filename
		
		if err := ctx.SaveUploadedFile(file, dst); err != nil {
			ctx.JSON(http.StatusInternalServerError, response.Response{
				Success: false,
				Message: "Failed to save image",
			})
			return
		}
		
		imageURL := "/uploads/" + filename
		image = &imageURL
	}

	if err := models.UpdateUserProfile(pc.DB, userID, &fullname, &email, image); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Response{
			Success: false,
			Message: "Failed to update profile",
		})
		return
	}

	userIDStr := strconv.Itoa(userID)
	profileCacheKey := "user:" + userIDStr + ":profile"
	statsCacheKey := "user:" + userIDStr + ":stats"
	_ = utils.RedisClient.Del(rctx, profileCacheKey, statsCacheKey)

	profile, _ := models.GetUserProfile(pc.DB, userID)
	stats, _ := models.GetDashboardStatsByUser(pc.DB, userID)

	ctx.JSON(http.StatusOK, response.Response{
		Success: true,
		Message: "Profile updated successfully",
		Data: gin.H{
			"profile": profile,
			"stats":   stats,
		},
	})
}