package seed

import (
	"testing"

	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestSeedCuratedStartups_FreshDatabaseIsIdempotent(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := db.AutoMigrate(&models.Startup{}); err != nil {
		t.Fatalf("migrate startups: %v", err)
	}

	if err := SeedCuratedStartups(db); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := SeedCuratedStartups(db); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	var rows []models.Startup
	if err := db.Order("slug asc").Find(&rows).Error; err != nil {
		t.Fatalf("list startups: %v", err)
	}
	want := []string{"canva", "databricks", "discord", "granola", "mercor", "notion", "openai", "stripe"}
	if len(rows) != len(want) {
		t.Fatalf("startup count = %d, want %d", len(rows), len(want))
	}
	for i, row := range rows {
		if row.Slug != want[i] {
			t.Fatalf("slug %d = %s, want %s", i, row.Slug, want[i])
		}
		if row.Name == "" || row.Description == "" || len(row.WebsiteURL) < 8 || row.WebsiteURL[:8] != "https://" {
			t.Fatalf("startup %s is missing public fields", row.Slug)
		}
	}
	var blocked int64
	if err := db.Unscoped().Model(&models.Startup{}).Where("slug IN ?", []string{"anthropic", "spacex", "step3-verify", "step8-admin"}).Count(&blocked).Error; err != nil {
		t.Fatalf("count blocked slugs: %v", err)
	}
	if blocked != 0 {
		t.Fatalf("fresh seed included %d test or excluded companies", blocked)
	}

	var mercor models.Startup
	if err := db.Where("slug = ?", "mercor").First(&mercor).Error; err != nil {
		t.Fatalf("load mercor: %v", err)
	}
	var rounds []models.FundingRound
	if err := db.Where("startup_id = ?", mercor.ID).Find(&rounds).Error; err != nil {
		t.Fatalf("list mercor funding: %v", err)
	}
	if len(rounds) != 1 {
		t.Fatalf("mercor funding count = %d, want 1", len(rounds))
	}
	round := rounds[0]
	if round.RoundLabel != "Series D" || round.AnnouncementDate != "2026-11-10" || round.ValuationBasis != "unspecified" || round.Currency != "USD" {
		t.Fatalf("unexpected mercor round: %+v", round)
	}
	if round.AmountRaisedMinor == nil || *round.AmountRaisedMinor != 75_000_000_000 || round.ValuationMinor == nil || *round.ValuationMinor != 2_000_000_000_000 {
		t.Fatalf("mercor money = %v / %v", round.AmountRaisedMinor, round.ValuationMinor)
	}
	if round.EvidenceURL != "https://www.mercor.com/blog/series-d/" || round.SourceLabel != "Mercor" {
		t.Fatalf("mercor source = %s %s", round.EvidenceURL, round.SourceLabel)
	}
	var otherFunding int64
	if err := db.Model(&models.FundingRound{}).Where("startup_id <> ?", mercor.ID).Count(&otherFunding).Error; err != nil {
		t.Fatalf("count other funding: %v", err)
	}
	if otherFunding != 0 {
		t.Fatalf("other companies have %d funding rounds", otherFunding)
	}
}

func TestSeedCuratedStartups_PreservesExistingRows(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := db.AutoMigrate(&models.Startup{}); err != nil {
		t.Fatalf("migrate startups: %v", err)
	}
	existing := []models.Startup{
		{Name: "Stripe Labs", Slug: "stripe", Description: "Edited locally", WebsiteURL: "https://stripe.com/about"},
		{Name: "Step 3 Verified", Slug: "step3-verify", Description: "Local test", WebsiteURL: "https://example.com/verified"},
		{Name: "Anthropic", Slug: "anthropic", Description: "Leave this row alone", WebsiteURL: "https://www.anthropic.com"},
		{Name: "Mercor Local", Slug: "mercor", Description: "Corrected description", WebsiteURL: "https://www.mercor.com/about"},
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("seed existing rows: %v", err)
	}
	var mercor models.Startup
	if err := db.Where("slug = ?", "mercor").First(&mercor).Error; err != nil {
		t.Fatalf("load mercor: %v", err)
	}
	correctedAmount := int64(100)
	corrected := models.FundingRound{
		StartupID:         mercor.ID,
		RoundLabel:        "Series D",
		AnnouncementDate:  "2026-11-10",
		AmountRaisedMinor: &correctedAmount,
		ValuationBasis:    "unspecified",
		Currency:          "USD",
		EvidenceURL:       "https://www.mercor.com/blog/series-d/",
		SourceLabel:       "Corrected source",
		Note:              "leave this correction",
	}
	if err := db.Create(&corrected).Error; err != nil {
		t.Fatalf("seed corrected funding: %v", err)
	}

	if err := SeedCuratedStartups(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var stripe models.Startup
	if err := db.Where("slug = ?", "stripe").First(&stripe).Error; err != nil {
		t.Fatalf("load stripe: %v", err)
	}
	if stripe.Name != "Stripe Labs" || stripe.Description != "Edited locally" || stripe.WebsiteURL != "https://stripe.com/about" {
		t.Fatalf("existing stripe was overwritten: %+v", stripe)
	}

	var extras int64
	if err := db.Model(&models.Startup{}).Where("slug IN ?", []string{"step3-verify", "anthropic"}).Count(&extras).Error; err != nil {
		t.Fatalf("count preserved rows: %v", err)
	}
	if extras != 2 {
		t.Fatalf("preserved rows = %d, want 2", extras)
	}

	var total int64
	if err := db.Model(&models.Startup{}).Count(&total).Error; err != nil {
		t.Fatalf("count startups: %v", err)
	}
	if total != 10 {
		t.Fatalf("startup count = %d, want 10", total)
	}

	var storedMercor models.Startup
	if err := db.Where("slug = ?", "mercor").First(&storedMercor).Error; err != nil {
		t.Fatalf("reload mercor: %v", err)
	}
	if storedMercor.Name != "Mercor Local" || storedMercor.Description != "Corrected description" {
		t.Fatalf("existing mercor was overwritten: %+v", storedMercor)
	}
	var storedRounds []models.FundingRound
	if err := db.Where("startup_id = ?", storedMercor.ID).Find(&storedRounds).Error; err != nil {
		t.Fatalf("reload funding: %v", err)
	}
	if len(storedRounds) != 1 || storedRounds[0].AmountRaisedMinor == nil || *storedRounds[0].AmountRaisedMinor != 100 || storedRounds[0].Note != "leave this correction" || storedRounds[0].SourceLabel != "Corrected source" {
		t.Fatalf("corrected funding changed: %+v", storedRounds)
	}
}

func TestSeedCuratedStartups_DoesNotRestoreSoftDeletedCompany(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	hidden := models.Startup{Name: "Mercor", Slug: "mercor", Description: "Hidden", WebsiteURL: "https://www.mercor.com"}
	if err := db.Create(&hidden).Error; err != nil {
		t.Fatalf("create mercor: %v", err)
	}
	if err := db.Delete(&hidden).Error; err != nil {
		t.Fatalf("soft delete mercor: %v", err)
	}

	if err := SeedCuratedStartups(db); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var visible int64
	if err := db.Model(&models.Startup{}).Where("slug = ?", "mercor").Count(&visible).Error; err != nil {
		t.Fatalf("count visible mercor: %v", err)
	}
	if visible != 0 {
		t.Fatalf("soft-deleted mercor was restored")
	}
	var funding int64
	if err := db.Unscoped().Model(&models.FundingRound{}).Where("startup_id = ?", hidden.ID).Count(&funding).Error; err != nil {
		t.Fatalf("count funding: %v", err)
	}
	if funding != 0 {
		t.Fatalf("funding was attached to a soft-deleted startup")
	}
}

func TestSeedCuratedFunding_LeavesACorrectedAnnouncementURL(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	if err := SeedCuratedStartups(db); err != nil {
		t.Fatalf("first seed: %v", err)
	}
	if err := db.Model(&models.FundingRound{}).Where("evidence_url = ?", "https://www.mercor.com/blog/series-d/").Update("evidence_url", "https://www.mercor.com/blog/series-d-correction/").Error; err != nil {
		t.Fatalf("change url: %v", err)
	}
	if err := SeedCuratedStartups(db); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	var rounds []models.FundingRound
	if err := db.Find(&rounds).Error; err != nil {
		t.Fatalf("list funding: %v", err)
	}
	if len(rounds) != 1 || rounds[0].EvidenceURL != "https://www.mercor.com/blog/series-d-correction/" {
		t.Fatalf("corrected announcement was duplicated: %+v", rounds)
	}
}
