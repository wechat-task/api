package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  databaseURL,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Migrate runs SQL-based migrations instead of GORM AutoMigrate
// This avoids accessing information_schema, making it compatible with
// managed database services like Supabase that restrict system table access.
func Migrate(db *gorm.DB) error {
	return RunMigrations(db)
}
