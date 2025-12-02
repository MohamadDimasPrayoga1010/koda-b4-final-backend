package models

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
    ID           int
    UserID       int
    RefreshToken string
    ExpiresAt    time.Time
}

func CreateSession(db *pgxpool.Pool, userID int, token, userAgent, ip string, expiresAt time.Time) error {
    ctx := context.Background()
    _, err := db.Exec(ctx,
        "INSERT INTO sessions (user_id, refresh_token, user_agent, ip_address, expires_at) VALUES ($1,$2,$3,$4,$5)",
        userID, token, userAgent, ip, expiresAt,
    )
    return err
}

func GetSessionByToken(db *pgxpool.Pool, token string) (Session, error) {
    ctx := context.Background()
    var s Session
    err := db.QueryRow(ctx, "SELECT id, user_id, refresh_token, expires_at FROM sessions WHERE refresh_token=$1", token).
        Scan(&s.ID, &s.UserID, &s.RefreshToken, &s.ExpiresAt)
    return s, err
}

func DeleteSession(db *pgxpool.Pool, token string) error {
    ctx := context.Background()
    _, err := db.Exec(ctx, "DELETE FROM sessions WHERE refresh_token=$1", token)
    return err
}
