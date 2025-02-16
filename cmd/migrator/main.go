package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/linemk/avito-shop/internal/config"
)

// buildMigrateDSN собирает строку подключения (DSN) из отдельных параметров
func buildMigrateDSN(dbCfg config.DatabaseConfig, migrationTable string, dbPassword string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&x-migrations-table=%s",
		dbCfg.User, dbPassword, dbCfg.Host, dbCfg.Port, dbCfg.Name, migrationTable,
	)
}

// buildQueryDSN собирает DSN для обычных SQL запросов
func buildQueryDSN(dbCfg config.DatabaseConfig, dbPassword string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbCfg.User, dbPassword, dbCfg.Host, dbCfg.Port, dbCfg.Name,
	)
}

func main() {
	var migrationsPathFlag string
	flag.StringVar(&migrationsPathFlag, "migrations-path", "", "path to migration files")
	flag.Parse()

	cfg := config.MustLoad()

	migrationsPath := cfg.Migrations.Path
	if migrationsPathFlag != "" {
		migrationsPath = migrationsPathFlag
	}

	dbPassword := ""
	if dbPassword = GetEnv("DB_PASSWORD", ""); dbPassword == "" {
		log.Fatal("DB_PASSWORD environment variable is required")
	}

	migrationTableName := "migrations"

	dsnForMigrate := buildMigrateDSN(cfg.Database, migrationTableName, dbPassword)
	log.Printf("Using DSN for migrate: %s", dsnForMigrate)

	// Создаем объект мигратора
	m, err := migrate.New(
		"file://"+migrationsPath,
		dsnForMigrate,
	)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations to apply")
		} else {
			log.Fatalf("migration failed: %v", err)
		}
	} else {
		log.Println("Migrations applied successfully")
	}

	dsnForQuery := buildQueryDSN(cfg.Database, dbPassword)

	db, err := sql.Open("postgres", dsnForQuery)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
		ORDER BY table_name
	`)
	if err != nil {
		log.Fatalf("failed to query tables: %v", err)
	}
	defer rows.Close()

	fmt.Println("Current tables in the database:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("failed to scan row: %v", err)
		}
		fmt.Println(" -", tableName)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("error reading rows: %v", err)
	}
}

func GetEnv(key, defaultValue string) string {
	if value, exists := lookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Обертка для os.LookupEnv, чтобы можно было легко подменить в тестах
func lookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}
