package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddMarketStartupLink adds an optional startup association to markets.
// Existing markets stay unassociated.
func MigrateAddMarketStartupLink(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.Startup{}); err != nil {
		return err
	}
	return db.AutoMigrate(&models.Market{})
}

func init() {
	migration.Register("20261119130000", func(db *gorm.DB) error {
		return MigrateAddMarketStartupLink(db)
	})
}
