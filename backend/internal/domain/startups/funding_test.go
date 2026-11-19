package startups

import (
	"context"
	"errors"
	"testing"
)

func TestFundingRoundRejectsZeroAndKeepsUnknownNull(t *testing.T) {
	svc := NewService(&memoryRepo{})
	created, err := svc.Create(context.Background(), CreateRequest{Name: "Northwind", Slug: "northwind"})
	if err != nil {
		t.Fatalf("create startup: %v", err)
	}

	raised := int64(1250)
	round, err := svc.CreateFundingRound(context.Background(), created.Slug, FundingInput{
		RoundLabel:        " Seed ",
		AnnouncementDate:  "2024-03-01",
		AmountRaisedMinor: &raised,
		Currency:          "usd",
		EvidenceURL:       "https://northwind.example/seed",
		SourceLabel:       " Northwind ",
		Note:              " announced ",
	})
	if err != nil {
		t.Fatalf("create funding: %v", err)
	}
	if round.AmountRaisedMinor == nil || *round.AmountRaisedMinor != 1250 || round.ValuationMinor != nil {
		t.Fatalf("money fields = amount %v valuation %v", round.AmountRaisedMinor, round.ValuationMinor)
	}
	if round.ValuationBasis != BasisUnspecified || round.Currency != "USD" || round.RoundLabel != "Seed" {
		t.Fatalf("normalized round: %+v", round)
	}

	zero := int64(0)
	if _, err := svc.CreateFundingRound(context.Background(), created.Slug, FundingInput{
		RoundLabel:        "Seed",
		AnnouncementDate:  "2024-03-01",
		AmountRaisedMinor: &zero,
		Currency:          "USD",
		EvidenceURL:       "https://northwind.example/seed",
		SourceLabel:       "Northwind",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("zero amount error = %v", err)
	}
	if _, err := svc.CreateFundingRound(context.Background(), created.Slug, FundingInput{
		RoundLabel:       "Seed",
		AnnouncementDate: "March 1",
		Currency:         "USD",
		EvidenceURL:      "https://northwind.example/seed",
		SourceLabel:      "Northwind",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("bad date error = %v", err)
	}
	if _, err := svc.CreateFundingRound(context.Background(), created.Slug, FundingInput{
		RoundLabel:       "Seed",
		AnnouncementDate: "2024-03-01",
		Currency:         "US",
		EvidenceURL:      "https://northwind.example/seed",
		SourceLabel:      "Northwind",
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("bad currency error = %v", err)
	}
	if _, err := svc.CreateFundingRound(context.Background(), "missing", FundingInput{
		RoundLabel:       "Seed",
		AnnouncementDate: "2024-03-01",
		Currency:         "USD",
		EvidenceURL:      "https://northwind.example/seed",
		SourceLabel:      "Northwind",
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing startup error = %v", err)
	}
}

func TestFundingHistoryOrdersRoundsAndLatestValuation(t *testing.T) {
	svc := NewService(&memoryRepo{})
	created, err := svc.Create(context.Background(), CreateRequest{Name: "Northwind", Slug: "northwind"})
	if err != nil {
		t.Fatalf("create startup: %v", err)
	}
	laterValuation := int64(500000)
	if _, err := svc.CreateFundingRound(context.Background(), created.Slug, FundingInput{
		RoundLabel:       "Series A",
		AnnouncementDate: "2025-01-15",
		ValuationMinor:   &laterValuation,
		ValuationBasis:   BasisPostMoney,
		Currency:         "USD",
		EvidenceURL:      "https://northwind.example/a",
		SourceLabel:      "Northwind",
	}); err != nil {
		t.Fatalf("later round: %v", err)
	}
	earlierRaised := int64(100)
	if _, err := svc.CreateFundingRound(context.Background(), created.Slug, FundingInput{
		RoundLabel:        "Seed",
		AnnouncementDate:  "2024-01-15",
		AmountRaisedMinor: &earlierRaised,
		Currency:          "USD",
		EvidenceURL:       "https://northwind.example/seed",
		SourceLabel:       "Northwind",
	}); err != nil {
		t.Fatalf("earlier round: %v", err)
	}

	history, err := svc.ListFundingHistory(context.Background(), created.Slug)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(history.Rounds) != 2 || history.Rounds[0].RoundLabel != "Seed" || history.Rounds[1].RoundLabel != "Series A" {
		t.Fatalf("order: %+v", history.Rounds)
	}
	if history.LatestReportedValuation == nil || history.LatestReportedValuation.RoundLabel != "Series A" || history.LatestReportedValuation.ValuationBasis != BasisPostMoney {
		t.Fatalf("latest: %+v", history.LatestReportedValuation)
	}

	other, err := svc.Create(context.Background(), CreateRequest{Name: "Other", Slug: "other"})
	if err != nil {
		t.Fatalf("other: %v", err)
	}
	if _, err := svc.UpdateFundingRound(context.Background(), other.Slug, history.Rounds[1].ID, FundingInput{
		RoundLabel:       "Series A",
		AnnouncementDate: "2025-01-15",
		Currency:         "USD",
		EvidenceURL:      "https://northwind.example/a",
		SourceLabel:      "Northwind",
	}); !errors.Is(err, ErrFundingNotFound) {
		t.Fatalf("foreign update error = %v", err)
	}
}
