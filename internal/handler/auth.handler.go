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

	token, err := utils.GenerateToken(int(user.ID), user.Email, user.Role)
	if err != nil {
		ctx.JSON(500, response.Response{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	user.Token = token

	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Login success",
		Data:    user,
	})
}

func (ac *AuthController) Logout(ctx *gin.Context) {
	ctx.JSON(200, response.Response{
		Success: true,
		Message: "Logout successful",
	})
}


