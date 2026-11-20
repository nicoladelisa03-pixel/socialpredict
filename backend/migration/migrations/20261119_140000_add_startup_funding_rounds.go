package migrations

import (
	"socialpredict/migration"
	"socialpredict/models"

	"gorm.io/gorm"
)

// MigrateAddStartupFundingRounds adds sourced funding history for startups.
func MigrateAddStartupFundingRounds(db *gorm.DB) error {
	return db.AutoMigrate(&models.FundingRound{})
}

func init() {
	migration.Register("20261119140000", func(db *gorm.DB) error {
		return MigrateAddStartupFundingRounds(db)
	})
}
