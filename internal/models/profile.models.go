package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserProfileResponse struct {
    Fullname string `json:"fullname"`
    Email    string `json:"email"`
    Image    string `json:"image"`
}

func GetUserProfile(db *pgxpool.Pool, userID int) (UserProfileResponse, error) {
    var profile UserProfileResponse
    err := db.QueryRow(context.Background(),
        `SELECT u.fullname, u.email, p.image
         FROM users u
         LEFT JOIN profile p ON u.id = p.user_id
         WHERE u.id=$1`, userID,
    ).Scan(&profile.Fullname, &profile.Email, &profile.Image)
    return profile, err
}
