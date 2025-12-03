package middleware

import (
	"fmt"
	"koda-shortlink/internal/utils"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		fmt.Println(authHeader)
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.JSON(401, gin.H{"success": false, "message": "Missing or invalid Authorization header"})
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			ctx.JSON(500, gin.H{"success": false, "message": "Server missing JWT secret"})
			ctx.Abort()
			return
		}

		claims := &utils.UserPayload{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(401, gin.H{"success": false, "message": "Invalid token"})
			ctx.Abort()
			return
		}

		ctx.Set("userID", int64(claims.Id))
		ctx.Set("userEmail", claims.Email)
		ctx.Set("userRole", claims.Role)

		if requiredRole != "" && claims.Role != requiredRole {
			ctx.JSON(403, gin.H{"success": false, "message": "No permission"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}


func OptAuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Println("option")
			ctx.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			ctx.JSON(500, gin.H{"success": false, "message": "Server missing JWT secret"})
			ctx.Abort()
			return
		}

		claims := &utils.UserPayload{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		fmt.Println(token.Valid)

		if err != nil || !token.Valid {
			ctx.JSON(401, gin.H{"success": false, "message": "Invalid token"})
			ctx.Abort()
			return
		}

		ctx.Set("userID", int64(claims.Id))
		ctx.Set("userEmail", claims.Email)
		ctx.Set("userRole", claims.Role)

		if requiredRole != "" && claims.Role != requiredRole {
			ctx.JSON(403, gin.H{"success": false, "message": "No permission"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
