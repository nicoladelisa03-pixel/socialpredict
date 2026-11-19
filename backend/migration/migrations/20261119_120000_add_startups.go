package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddStartups adds the admin-managed startup profile table.
func MigrateAddStartups(db *gorm.DB) error {
	return db.AutoMigrate(&models.Startup{})
}

func init() {
	migration.Register("20261119120000", func(db *gorm.DB) error {
		return MigrateAddStartups(db)
	})
}
