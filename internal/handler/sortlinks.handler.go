package handler

import (
	"koda-shortlink/internal/models"
	"koda-shortlink/internal/utils"
	"koda-shortlink/pkg/response"

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


// @Summary Get a shortlink by code
// @Description Retrieve a shortlink by its short code and log the click
// @Tags Shortlinks
// @Accept json
// @Produce json
// @Param shortCode path string true "Shortlink code"
// @Success 200 {object} response.Response "Returns the shortlink data"
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
	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Shortlink retrieved successfully",
		Data: gin.H{
			"id":            sl.ID,
			"original_url":  sl.OriginalURL,
			"short_code":    sl.ShortCode,
			"redirect_count": sl.RedirectCount + 1, 
			"created_at":    sl.CreatedAt,
			"updated_at":    sl.UpdatedAt,
		},
	})
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
