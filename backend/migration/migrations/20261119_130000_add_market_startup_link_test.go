package migrations_test

import (
	"testing"
	"time"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddMarketStartupLinkAddsNullableAssociation(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddMarketStartupLink(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if !db.Migrator().HasColumn(&models.Market{}, "StartupID") {
		t.Fatalf("expected markets.startup_id")
	}

	market := models.Market{
		QuestionTitle:      "Unlinked",
		Description:        "Still valid",
		OutcomeType:        "BINARY",
		ResolutionDateTime: time.Now().Add(time.Hour),
		InitialProbability: 0.5,
		CreatorUsername:    "creator",
		LifecycleStatus:    "published",
	}
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create unlinked market: %v", err)
	}
	if market.StartupID != nil {
		t.Fatalf("new market startup id = %v, want nil", market.StartupID)
	}
}

func TestMigrateAddMarketStartupLinkIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddMarketStartupLink(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddMarketStartupLink(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
