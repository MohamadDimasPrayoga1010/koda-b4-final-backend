package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)


type UserRegister struct {
    Fullname string `json:"fullname" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Role     string `json:"role"`
} 


type UserResponse struct {
    ID        int64     `json:"id"`
    Fullname  string    `json:"fullname"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    Token     string    `json:"token,omitempty"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

type UserLogin struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func RegisterUser(db *pgxpool.Pool, user UserRegister, hashedPassword string) (UserResponse, error) {
    var resp UserResponse

    role := user.Role
    if role == "" {
        role = "user"
    }

    query := `
        INSERT INTO users (fullname, email, password, role, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        RETURNING id, fullname, email, role, created_at, updated_at
    `

    err := db.QueryRow(
        context.Background(),
        query,
        user.Fullname,
        user.Email,
        hashedPassword,
        role,
    ).Scan(
        &resp.ID,
        &resp.Fullname,
        &resp.Email,
        &resp.Role,
        &resp.CreatedAt,
        &resp.UpdatedAt,
    )

    if err != nil {
        return UserResponse{}, err
    }

    return resp, nil
}


func LoginUser(db *pgxpool.Pool, email string) (*UserResponse, string, string, error) {
	var user UserResponse
	var hashedPassword string

	query := `
		SELECT id, fullname, email, password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := db.QueryRow(context.Background(), query, email).Scan(
		&user.ID,
		&user.Fullname,
		&user.Email,
		&hashedPassword,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, "", "", err
	}

	return &user, hashedPassword, user.Role, nil
}

