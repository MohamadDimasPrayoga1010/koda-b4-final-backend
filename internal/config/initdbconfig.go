package config

import "github.com/jackc/pgx/v5/pgxpool"

func InitDbConfig() *pgxpool.Pool {
	return InitDB()
}