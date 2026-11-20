package seed

import (
	"context"
	"errors"
	"fmt"

	dstartups "socialpredict/internal/domain/startups"
	startuprepo "socialpredict/internal/repository/startups"
	"socialpredict/models"

	"gorm.io/gorm"
)

type curatedStartup struct {
	Name        string
	Slug        string
	Description string
	WebsiteURL  string
}

// curatedPrivateStartups is the public roster for a fresh database.
// Descriptions are short restatements of each company's own website.
// Funding is seeded separately and is not copied into these descriptions.
func curatedPrivateStartups() []curatedStartup {
	return []curatedStartup{
		{
			Name:        "Stripe",
			Slug:        "stripe",
			Description: "Stripe provides financial infrastructure for businesses to accept payments, offer billing, and manage revenue.",
			WebsiteURL:  "https://stripe.com",
		},
		{
			Name:        "OpenAI",
			Slug:        "openai",
			Description: "OpenAI is an AI research and deployment company. Its stated mission is to ensure that artificial general intelligence benefits all of humanity.",
			WebsiteURL:  "https://openai.com",
		},
		{
			Name:        "Databricks",
			Slug:        "databricks",
			Description: "Databricks is a data and AI company. It offers a data and AI platform built on a lakehouse architecture, and was founded in 2013 by the creators of Apache Spark.",
			WebsiteURL:  "https://www.databricks.com",
		},
		{
			Name:        "Canva",
			Slug:        "canva",
			Description: "Canva is an online design and publishing tool. The company says it launched in 2013 to help people design and publish.",
			WebsiteURL:  "https://www.canva.com",
		},
		{
			Name:        "Notion",
			Slug:        "notion",
			Description: "Notion is an all-in-one workspace. The company describes it as a place to combine notes, tasks, roadmaps, and other work in customizable building blocks.",
			WebsiteURL:  "https://www.notion.com",
		},
		{
			Name:        "Discord",
			Slug:        "discord",
			Description: "Discord is a communications platform for voice, video, and text. It was built so people can talk with friends while playing games.",
			WebsiteURL:  "https://discord.com",
		},
		{
			Name:        "Mercor",
			Slug:        "mercor",
			Description: "Mercor organizes human expertise for AI work. Its site describes a platform that connects specialists with enterprises for evaluations and AI products.",
			WebsiteURL:  "https://www.mercor.com",
		},
		{
			Name:        "Granola",
			Slug:        "granola",
			Description: "Granola is an AI notepad for meetings. It writes notes and action items from the conversation without adding a meeting bot, and it works with Zoom, Google Meet, and Teams.",
			WebsiteURL:  "https://www.granola.ai",
		},
	}
}

// SeedCuratedStartups inserts any missing curated company. It does not update
// or delete rows that are already stored, including local test startups.
func SeedCuratedStartups(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("seed curated startups: database unavailable")
	}
	for _, item := range curatedPrivateStartups() {
		var count int64
		if err := db.Unscoped().Model(&models.Startup{}).Where("slug = ?", item.Slug).Count(&count).Error; err != nil {
			return fmt.Errorf("check startup %s: %w", item.Slug, err)
		}
		if count > 0 {
			continue
		}
		row := models.Startup{
			Name:        item.Name,
			Slug:        item.Slug,
			Description: item.Description,
			WebsiteURL:  item.WebsiteURL,
		}
		if err := db.Create(&row).Error; err != nil {
			return fmt.Errorf("create startup %s: %w", item.Slug, err)
		}
	}
	return seedCuratedFunding(db)
}

type curatedFundingRound struct {
	Slug              string
	RoundLabel        string
	AnnouncementDate  string
	AmountRaisedMinor int64
	ValuationMinor    int64
	EvidenceURL       string
	SourceLabel       string
	Note              string
}

// curatedFundingRounds is the sourced funding inserted on a fresh database.
// Amounts are integer cents. Mercor's figures come from its 10 November 2026
// Series D announcement and are not a market price.
func curatedFundingRounds() []curatedFundingRound {
	return []curatedFundingRound{
		{
			Slug:              "mercor",
			RoundLabel:        "Series D",
			AnnouncementDate:  "2026-11-10",
			AmountRaisedMinor: 750_000_000 * 100,
			ValuationMinor:    20_000_000_000 * 100,
			EvidenceURL:       "https://www.mercor.com/blog/series-d/",
			SourceLabel:       "Mercor",
			Note:              "Mercor's announcement does not say whether the valuation is pre-money or post-money.",
		},
	}
}

// seedCuratedFunding inserts a sourced round only when that startup is active
// and neither its evidence URL nor the same round label and announcement date
// is already stored. An existing round is left as it was saved.
func seedCuratedFunding(db *gorm.DB) error {
	svc := dstartups.NewService(startuprepo.NewGormRepository(db))
	for _, item := range curatedFundingRounds() {
		var startup models.Startup
		err := db.Where("slug = ?", item.Slug).First(&startup).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return fmt.Errorf("load startup %s for funding: %w", item.Slug, err)
		}
		var count int64
		err = db.Unscoped().Model(&models.FundingRound{}).
			Where(
				"startup_id = ? AND (evidence_url = ? OR (round_label = ? AND announcement_date = ?))",
				startup.ID,
				item.EvidenceURL,
				item.RoundLabel,
				item.AnnouncementDate,
			).
			Count(&count).Error
		if err != nil {
			return fmt.Errorf("check funding for %s: %w", item.Slug, err)
		}
		if count > 0 {
			continue
		}
		raised := item.AmountRaisedMinor
		valuation := item.ValuationMinor
		if _, err := svc.CreateFundingRound(context.Background(), item.Slug, dstartups.FundingInput{
			RoundLabel:        item.RoundLabel,
			AnnouncementDate:  item.AnnouncementDate,
			AmountRaisedMinor: &raised,
			ValuationMinor:    &valuation,
			ValuationBasis:    dstartups.BasisUnspecified,
			Currency:          "USD",
			EvidenceURL:       item.EvidenceURL,
			SourceLabel:       item.SourceLabel,
			Note:              item.Note,
		}); err != nil {
			return fmt.Errorf("create funding for %s: %w", item.Slug, err)
		}
	}
	return nil
}
