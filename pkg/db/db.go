package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func envOrDefault(key, d string) string {
	v := os.Getenv(key)
	if v == "" {
		return d
	}
	return v
}

func NewGormDBFromEnv() (*gorm.DB, error) {
	host := envOrDefault("DB_HOST", "localhost")
	port := envOrDefault("DB_PORT", "5432")
	user := envOrDefault("DB_USER", "todo")
	password := envOrDefault("DB_PASSWORD", "todo")
	dbname := envOrDefault("DB_NAME", "todo_db")
	ssl := envOrDefault("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, ssl)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// configure sql.DB connection pool
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	// sensible defaults
	maxIdleConns := 10
	maxOpenConns := 100
	connMaxLifetime := time.Minute * 30

	if v := os.Getenv("DB_MAX_IDLE_CONNS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			maxIdleConns = i
		}
	}
	if v := os.Getenv("DB_MAX_OPEN_CONNS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			maxOpenConns = i
		}
	}
	if v := os.Getenv("DB_CONN_MAX_LIFETIME_MIN"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			connMaxLifetime = time.Duration(i) * time.Minute
		}
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// quick ping check with timeout
	tctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(tctx); err != nil {
		log.Printf("db ping failed: %v", err)
		return nil, err
	}

	return gormDB, nil
}
