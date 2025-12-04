package models

import (
	"context"
	"koda-shortlink/internal/utils"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Shortlink struct {
	ID            int     `json:"id"`
	UserID        *int64  `json:"userId"`
	OriginalURL   string  `json:"originalUrl"`
	ShortCode     string  `json:"shortCode"`
	RedirectCount int     `json:"redirectCount"`
	Status        string  `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}


type ShortlinkClick struct {
	ID          int       `json:"id"`
	ShortlinkID int       `json:"shortlinkId"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"userAgent"`
	CreatedAt   time.Time `json:"createdAt"`
}

func CreateShortlink(db *pgxpool.Pool, sl Shortlink) (Shortlink, error) {
    if sl.Status == "" {
        sl.Status = "active"
    }

    err := db.QueryRow(
        context.Background(),
        `INSERT INTO shortlinks (user_id, original_url, short_code, status)
         VALUES ($1, $2, $3, $4)
         RETURNING id, status, created_at, updated_at`,
        sl.UserID, sl.OriginalURL, sl.ShortCode, sl.Status,
    ).Scan(&sl.ID, &sl.Status, &sl.CreatedAt, &sl.UpdatedAt)

    return sl, err
}


func GetAllShortlinks(db *pgxpool.Pool, userID int64, limit, offset int) ([]Shortlink, int, error) {
	var total int
	err := db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM shortlinks WHERE user_id=$1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(context.Background(),
		`SELECT id, user_id, original_url, short_code, redirect_count, created_at, updated_at, status 
		 FROM shortlinks 
		 WHERE user_id=$1 
		 ORDER BY created_at DESC 
		 LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []Shortlink
	for rows.Next() {
		var sl Shortlink
		if err := rows.Scan(&sl.ID, &sl.UserID, &sl.OriginalURL, &sl.ShortCode, &sl.RedirectCount, &sl.CreatedAt, &sl.UpdatedAt, &sl.Status); err != nil {
			return nil, 0, err
		}
		result = append(result, sl)
	}

	return result, total, nil
}

func GetShortlinkByCode(db *pgxpool.Pool, code string) (Shortlink, error) {
	var sl Shortlink
	err := db.QueryRow(
		context.Background(),
		`SELECT id, user_id, original_url, short_code, redirect_count, status, created_at, updated_at 
		 FROM shortlinks WHERE short_code=$1`,
		code,
	).Scan(&sl.ID, &sl.UserID, &sl.OriginalURL, &sl.ShortCode, &sl.RedirectCount, &sl.Status, &sl.CreatedAt, &sl.UpdatedAt)
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
	if err == nil {
		utils.RedisClient.Del(context.Background(), "analytics:global:7d")
	}
	return err
}

func LogClick(db *pgxpool.Pool, click ShortlinkClick) error {
	_, err := db.Exec(
		context.Background(),
		`INSERT INTO shortlink_clicks (shortlink_id, ip_address, user_agent, clicked_at) 
		 VALUES ($1, $2, $3, now())`,
		click.ShortlinkID, click.IP, click.UserAgent,
	)
	return err
}

func UpdateShortlink(db *pgxpool.Pool, sl Shortlink) (Shortlink, error) {
	err := db.QueryRow(
		context.Background(),
		`UPDATE shortlinks 
		 SET original_url=$1, short_code=$2, status=$3, updated_at=now() 
		 WHERE id=$4
		 RETURNING id, user_id, original_url, short_code, redirect_count, status, created_at, updated_at`,
		sl.OriginalURL, sl.ShortCode, sl.Status, sl.ID,
	).Scan(&sl.ID, &sl.UserID, &sl.OriginalURL, &sl.ShortCode, &sl.RedirectCount, &sl.Status, &sl.CreatedAt, &sl.UpdatedAt)
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

type DailyVisit struct {
	Date   string `json:"date"`
	Visits int    `json:"visits"`
}

type DashboardStats struct {
	TotalLinks          int          `json:"totalLinks"`
	TotalVisits         int          `json:"totalVisits"`
	AvgClickRate        float64      `json:"avgClickRate"`
	VisitsGrowth        float64      `json:"visitsGrowth"`
	Last7Days           []DailyVisit `json:"last7Days"`
	Last7DaysShortlinks []Shortlink  `json:"last7DaysShortlinks"`
}

func GetDashboardStats(db *pgxpool.Pool) (DashboardStats, error) {
	var stats DashboardStats

	err := db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM shortlinks`,
	).Scan(&stats.TotalLinks)
	if err != nil {
		return stats, err
	}

	err = db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM shortlink_clicks`,
	).Scan(&stats.TotalVisits)
	if err != nil {
		return stats, err
	}

	if stats.TotalLinks > 0 {
		stats.AvgClickRate = float64(stats.TotalVisits) / float64(stats.TotalLinks)
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -7)
	var thisWeek, lastWeek int

	err = db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM shortlink_clicks 
		 WHERE clicked_at >= $1`, weekStart,
	).Scan(&thisWeek)
	if err != nil {
		thisWeek = 0
	}

	lastWeekStart := weekStart.AddDate(0, 0, -7)
	lastWeekEnd := weekStart
	err = db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM shortlink_clicks 
		 WHERE clicked_at >= $1 AND clicked_at < $2`,
		lastWeekStart, lastWeekEnd,
	).Scan(&lastWeek)
	if err != nil {
		lastWeek = 0
	}

	if lastWeek > 0 {
		stats.VisitsGrowth = float64(thisWeek-lastWeek) / float64(lastWeek) * 100
	} else if thisWeek > 0 {
		stats.VisitsGrowth = 100
	}

	stats.Last7Days = make([]DailyVisit, 7)
	for i := 0; i < 7; i++ {
		day := weekStart.AddDate(0, 0, i)
		var count int
		db.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM shortlink_clicks 
			 WHERE DATE(clicked_at) = $1`,
			day.Format("2006-01-02"),
		).Scan(&count)

		stats.Last7Days[i] = DailyVisit{
			Date:   day.Format("2006-01-02"),
			Visits: count,
		}
	}

	return stats, nil
}

func GetDashboardStatsByUser(db *pgxpool.Pool, userID int) (DashboardStats, error) {
	var stats DashboardStats

	err := db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM shortlinks WHERE user_id=$1`, userID,
	).Scan(&stats.TotalLinks)
	if err != nil {
		return stats, err
	}

	err = db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM shortlink_clicks 
		 WHERE shortlink_id IN (SELECT id FROM shortlinks WHERE user_id=$1)`,
		userID,
	).Scan(&stats.TotalVisits)
	if err != nil {
		return stats, err
	}

	if stats.TotalLinks > 0 {
		stats.AvgClickRate = float64(stats.TotalVisits) / float64(stats.TotalLinks)
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -7)

	var thisWeek, lastWeek int
	err = db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM shortlink_clicks 
		 WHERE shortlink_id IN (SELECT id FROM shortlinks WHERE user_id=$1)
		 AND clicked_at >= $2`, userID, weekStart,
	).Scan(&thisWeek)
	if err != nil {
		thisWeek = 0
	}

	lastWeekStart := weekStart.AddDate(0, 0, -7)
	lastWeekEnd := weekStart
	err = db.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM shortlink_clicks 
		 WHERE shortlink_id IN (SELECT id FROM shortlinks WHERE user_id=$1)
		 AND clicked_at >= $2 AND clicked_at < $3`, userID, lastWeekStart, lastWeekEnd,
	).Scan(&lastWeek)
	if err != nil {
		lastWeek = 0
	}

	if lastWeek > 0 {
		stats.VisitsGrowth = float64(thisWeek-lastWeek) / float64(lastWeek) * 100
	} else if thisWeek > 0 {
		stats.VisitsGrowth = 100
	}

	stats.Last7Days = make([]DailyVisit, 7)
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -6+i)  
		var count int
		_ = db.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM shortlink_clicks 
			 WHERE shortlink_id IN (SELECT id FROM shortlinks WHERE user_id=$1)
			 AND DATE(clicked_at) = $2`, userID, day.Format("2006-01-02"),
		).Scan(&count)

		stats.Last7Days[i] = DailyVisit{
			Date:   day.Format("Jan 02"), 
			Visits: count,
		}
	}

	return stats, nil
}
