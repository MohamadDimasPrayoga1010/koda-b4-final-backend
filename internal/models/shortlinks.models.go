package models

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Shortlink struct {
	ID            int
	UserID        *int
	OriginalURL   string
	ShortCode     string
	RedirectCount int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ShortlinkClick struct {
	ID          int
	ShortlinkID int
	IP          string
	UserAgent   string
	CreatedAt   time.Time
}

func CreateShortlink(db *pgxpool.Pool, sl Shortlink) (Shortlink, error) {
	err := db.QueryRow(
		context.Background(),
		`INSERT INTO shortlinks (user_id, original_url, short_code) 
		 VALUES ($1,$2,$3) 
		 RETURNING id, created_at, updated_at`,
		sl.UserID, sl.OriginalURL, sl.ShortCode,
	).Scan(&sl.ID, &sl.CreatedAt, &sl.UpdatedAt)
	return sl, err
}

func GetAllShortlinks(db *pgxpool.Pool) ([]Shortlink, error) {
	rows, err := db.Query(context.Background(), 
		`SELECT id, user_id, original_url, short_code, redirect_count, created_at, updated_at 
		 FROM shortlinks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Shortlink
	for rows.Next() {
		var sl Shortlink
		if err := rows.Scan(&sl.ID, &sl.UserID, &sl.OriginalURL, &sl.ShortCode, &sl.RedirectCount, &sl.CreatedAt, &sl.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, sl)
	}
	return result, nil
}


func GetShortlinkByCode(db *pgxpool.Pool, code string) (Shortlink, error) {
	var sl Shortlink
	err := db.QueryRow(
		context.Background(),
		`SELECT id, user_id, original_url, short_code, redirect_count, created_at, updated_at 
		 FROM shortlinks WHERE short_code=$1`,
		code,
	).Scan(&sl.ID, &sl.UserID, &sl.OriginalURL, &sl.ShortCode, &sl.RedirectCount, &sl.CreatedAt, &sl.UpdatedAt)
	return sl, err
}

func IncrementRedirectCount(db *pgxpool.Pool, shortlinkID int) error {
	_, err := db.Exec(
		context.Background(),
		`UPDATE shortlinks 
		 SET redirect_count = redirect_count + 1, updated_at = now() 
		 WHERE id=$1`,
		shortlinkID,
	)
	return err
}

func LogClick(db *pgxpool.Pool, click ShortlinkClick) error {
	_, err := db.Exec(
		context.Background(),
		`INSERT INTO shortlink_clicks (shortlink_id, ip_address, user_agent, clicked_at) 
		 VALUES ($1,$2,$3, now())`,
		click.ShortlinkID, click.IP, click.UserAgent,
	)
	return err
}


func UpdateShortlink(db *pgxpool.Pool, sl Shortlink) (Shortlink, error) {
	err := db.QueryRow(
		context.Background(),
		`UPDATE shortlinks 
		 SET original_url=$1, short_code=$2, updated_at=now() 
		 WHERE id=$3
		 RETURNING id, user_id, original_url, short_code, redirect_count, created_at, updated_at`,
		sl.OriginalURL, sl.ShortCode, sl.ID,
	).Scan(&sl.ID, &sl.UserID, &sl.OriginalURL, &sl.ShortCode, &sl.RedirectCount, &sl.CreatedAt, &sl.UpdatedAt)
	return sl, err
}

func CheckShortCodeExists(db *pgxpool.Pool, code string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		context.Background(),
		`SELECT EXISTS(SELECT 1 FROM shortlinks WHERE short_code=$1)`,
		code,
	).Scan(&exists)

	return exists, err
}


func DeleteShortlink(db *pgxpool.Pool, shortCode string) error {
	_, err := db.Exec(context.Background(),
		`DELETE FROM shortlinks WHERE short_code=$1`,
		shortCode,
	)
	return err
}

