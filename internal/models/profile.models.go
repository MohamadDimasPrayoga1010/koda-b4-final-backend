package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserProfileResponse struct {
	Fullname string  `json:"fullname"`
	Email    string  `json:"email"`
	Image    *string `json:"image"`
}

func GetUserProfile(db *pgxpool.Pool, userID int) (UserProfileResponse, error) {
	var profile UserProfileResponse

	err := db.QueryRow(context.Background(),
		`SELECT u.fullname, u.email, p.image
	 FROM users u
	 LEFT JOIN profile p ON u.id = p.user_id
	 WHERE u.id=$1`, userID,
	).Scan(&profile.Fullname, &profile.Email, &profile.Image)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserProfileResponse{}, errors.New("user not found")
		}
		return UserProfileResponse{}, err
	}

	return profile, nil

}

func UpdateUserProfile(db *pgxpool.Pool, userID int, fullname, email, image *string) error {
	_, err := db.Exec(context.Background(),
		`UPDATE users u
		 SET fullname = COALESCE($1, u.fullname),
		     email = COALESCE($2, u.email)
		 WHERE u.id = $3`,
		fullname, email, userID,
	)
	if err != nil {
		return err
	}

	if image != nil {
		_, err := db.Exec(context.Background(),
			`INSERT INTO profile(user_id, image) 
		 VALUES ($1, $2)
		 ON CONFLICT(user_id) DO UPDATE SET image = EXCLUDED.image`,
			userID, image,
		)
		if err != nil {
			return err
		}
	}

	return nil

}
