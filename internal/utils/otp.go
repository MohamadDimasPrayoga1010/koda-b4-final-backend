package utils

import (
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateOTP(length int) string {
	var digits = "0123456789"
	otp := make([]byte, length)

	for i := range otp{
		otp[i] = digits[rand.Intn(len(digits))]
	}

	return string(otp)
}

func GenerateTokenForReset(email string) (string, error){
	secretKey := os.Getenv("JWT_OTP")

	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(1 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}