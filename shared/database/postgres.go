package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/lucas/shared/utils"
)

type PostgreSQLConfig struct {
	Host            string
	Port            string
	Database        string
	Username        string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

var db *sql.DB

func GetPostgreSQLConfig() PostgreSQLConfig {
	return PostgreSQLConfig{
		Host:            utils.GetEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            utils.GetEnvOrDefault("POSTGRES_PORT", "5432"),
		Database:        utils.GetEnvOrDefault("POSTGRES_DB", "microcommerce"),
		Username:        utils.GetEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:        utils.GetEnvOrDefault("POSTGRES_PASSWORD", "password"),
		SSLMode:         utils.GetEnvOrDefault("POSTGRES_SSLMODE", "disable"),
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	}
}

func ConnectPostgreSQL(config PostgreSQLConfig) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode,
	)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to PostgreSQL database: %s", config.Database)
	return nil
}

func GetDB() *sql.DB {
	return db
}

func ClosePostgreSQL() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func RunMigrations(migrationsPath string) error {
	if db == nil {
		return fmt.Errorf("database connection not established")
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Printf("Database migrations completed successfully")
	return nil
}

func HealthCheck() error {
	if db == nil {
		return fmt.Errorf("database connection not established")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}
