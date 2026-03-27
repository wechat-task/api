package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Migration struct {
	Version int
	Name    string
	UpSQL   string
}

func getMigrations() ([]Migration, error) {
	var migrations []Migration

	files, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migration files: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		// Extract version from filename (e.g., 000001_create_users_table.up.sql)
		parts := strings.Split(file.Name(), "_")
		if len(parts) < 2 {
			continue
		}

		var version int
		_, err := fmt.Sscanf(parts[0], "%d", &version)
		if err != nil {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + file.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		name := strings.TrimSuffix(file.Name(), ".up.sql")
		name = strings.TrimPrefix(name, parts[0]+"_")

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			UpSQL:   string(content),
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func RunMigrations(db *gorm.DB) error {
	// Get SQL database connection
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql database: %w", err)
	}

	migrations, err := getMigrations()
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	// Create schema_migrations table to track applied migrations
	if err := createSchemaMigrationsTable(sqlDB); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Get applied migrations
	appliedVersions, err := getAppliedMigrations(sqlDB)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if _, applied := appliedVersions[migration.Version]; applied {
			continue // Already applied
		}

		if err := runMigration(sqlDB, &migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
		}
	}

	return nil
}

func createSchemaMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	query := `SELECT version FROM schema_migrations ORDER BY version;`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func runMigration(db *sql.DB, migration *Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration
	recordQuery := `INSERT INTO schema_migrations (version, name) VALUES ($1, $2);`
	if _, err := tx.Exec(recordQuery, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}
