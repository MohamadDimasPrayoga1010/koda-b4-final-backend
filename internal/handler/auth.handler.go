package handler

import (
	"koda-shortlink/internal/models"
	"koda-shortlink/internal/utils"
	"koda-shortlink/pkg/response"
	"strings"


	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthController struct {
	DB *pgxpool.Pool
}


// @Summary Register a new user
// @Description Register a new user with fullname, email, password, and optional role. Default role is 'user' if not provided.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.UserRegister true "User registration payload" example({"fullname":"John Doe","email":"[john@example.com](mailto:john@example.com)","password":"secret123","role":"user"})
// @Success 201 {object} response.Response "Returns the created user data"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 409 {object} response.Response "Email already registered"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/auth/register [post]
func (ac *AuthController) Register(ctx *gin.Context) {
	var req models.UserRegister

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, response.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to hash password",
		})
		return
	}

	user, err := models.RegisterUser(ac.DB, req, hashed)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			ctx.JSON(409, response.Response{
				Success: false,
				Message: "Email already registered",
			})
			return
		}
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to register user",
		})
		return
	}

	if req.Role == "admin" {
		ctx.JSON(201, response.Response{
			Success: true,
			Message: "Admin registered successfully",
			Data:    user,
		})
		return
	}

	ctx.JSON(201, response.Response{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}

// LoginRequest godoc
// @Summary Login user
// @Description Login user dengan email dan password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.UserLogin true "Login payload"
// @Success 200 {object} response.Response{data=object{user=models.UserResponse,token=string,refreshToken=string}} "Returns user data and access token"
// @Failure 400 {object} response.Response "Invalid request body"
// @Failure 401 {object} response.Response "Email or password incorrect"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/auth/login [post]
func (ac *AuthController) Login(ctx *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(400, response.Response{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	user, hashedPassword, _, err := models.LoginUser(ac.DB, input.Email)
	if err != nil {
		ctx.JSON(401, response.Response{
			Success: false,
			Message: "Email or password incorrect",
		})
		return
	}

	ok, err := utils.VerifyPassword(input.Password, hashedPassword)
	if err != nil || !ok {
		ctx.JSON(401, response.Response{
			Success: false,
			Message: "Email or password incorrect",
		})
		return
	}

	accessToken, err := utils.GenerateToken(int(user.ID), user.Email, user.Role)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to generate access token",
		})
		return
	}
	refreshToken, err := utils.GenerateRefreshToken(int(user.ID), user.Email, user.Role)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to generate refresh token",
		})
		return
	}

	user.Token = accessToken

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Login success",
		Data: gin.H{
			"user":         user,
			"token":        accessToken,
			"refreshToken": refreshToken,
		},
	})
}


// RefreshTokenRequest godoc
// @Summary Refresh access token
// @Description Refresh access token menggunakan refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body object{refreshToken=string} true "Refresh Token payload"
// @Success 200 {object} response.Response{data=object{token=string}} "Returns new access token"
// @Failure 400 {object} response.Response "Refresh token required"
// @Failure 401 {object} response.Response "Invalid refresh token"
// @Failure 500 {object} response.Response "Failed to generate access token"
// @Router /api/v1/auth/refresh [post]
func (ac *AuthController) RefreshToken(ctx *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, response.Response{
			Success: false,
			Message: "refresh token required",
		})
		return
	}

	claims, err := utils.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(401, response.Response{
			Success: false,
			Message: "invalid refresh token",
		})
		return
	}

	newAccessToken, err := utils.GenerateToken(claims.Id, claims.Email, claims.Role)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "failed to generate access token",
		})
		return
	}

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "token refreshed",
		Data: gin.H{
			"token": newAccessToken,
		},
	})
}


// Logout godoc
// @Summary Logout user
// @Description Logout user (client harus menghapus token dari storage)
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response "Logout successful"
// @Failure 401 {object} response.Response "Unauthorized"
// @Router /api/v1/auth/logout [post]
func (ac *AuthController) Logout(ctx *gin.Context) {
	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Logout successful",
	})
}




