package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddStartupFundingRoundsPreservesStartups(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddStartups(db); err != nil {
		t.Fatalf("startup migration: %v", err)
	}
	startup := models.Startup{Name: "Kept", Slug: "kept", Description: "already stored"}
	if err := db.Create(&startup).Error; err != nil {
		t.Fatalf("create startup: %v", err)
	}

	if err := migrations.MigrateAddStartupFundingRounds(db); err != nil {
		t.Fatalf("funding migration: %v", err)
	}
	if err := migrations.MigrateAddStartupFundingRounds(db); err != nil {
		t.Fatalf("second funding migration: %v", err)
	}
	if !db.Migrator().HasTable(&models.FundingRound{}) {
		t.Fatalf("expected funding_rounds table")
	}
	if !db.Migrator().HasColumn(&models.FundingRound{}, "AmountRaisedMinor") {
		t.Fatalf("expected nullable amount column")
	}

	var stored models.Startup
	if err := db.Where("slug = ?", "kept").First(&stored).Error; err != nil {
		t.Fatalf("load startup: %v", err)
	}
	if stored.Name != "Kept" || stored.Description != "already stored" {
		t.Fatalf("startup changed: %+v", stored)
	}
}
