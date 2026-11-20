package migrations_test

import (
	"testing"

	"socialpredict/migration/migrations"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMigrateAddStartupsCreatesStartupTable(t *testing.T) {
	db := modelstesting.NewTestDB(t)

	if err := migrations.MigrateAddStartups(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	if !db.Migrator().HasTable(&models.Startup{}) {
		t.Fatalf("expected startups table")
	}
	if !db.Migrator().HasIndex(&models.Startup{}, "idx_startups_slug") {
		t.Fatalf("expected unique slug index")
	}
}

func TestMigrateAddStartupsIsIdempotent(t *testing.T) {
	db := modelstesting.NewTestDB(t)
	if err := migrations.MigrateAddStartups(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := migrations.MigrateAddStartups(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
}
