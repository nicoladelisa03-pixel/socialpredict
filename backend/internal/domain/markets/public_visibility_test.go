package markets_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	markets "socialpredict/internal/domain/markets"
	"socialpredict/models/modelstesting"
)

func TestPublicReadsHideUnapprovedMarkets(t *testing.T) {
	service, db, _ := setupServiceWithDB(t)
	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	secret := "Secret proposed title"
	proposed := modelstesting.GenerateMarket(41, "creator")
	proposed.QuestionTitle = secret
	proposed.Description = "secret description"
	proposed.LifecycleStatus = string(markets.MarketLifecycleProposed)
	if err := db.Create(&proposed).Error; err != nil {
		t.Fatalf("create proposed market: %v", err)
	}

	published := modelstesting.GenerateMarket(42, "creator")
	published.QuestionTitle = "Public approved title"
	published.ResolutionDateTime = time.Now().Add(-time.Hour)
	if err := db.Create(&published).Error; err != nil {
		t.Fatalf("create published market: %v", err)
	}

	hidden := []struct {
		id        int64
		lifecycle string
	}{
		{43, string(markets.MarketLifecycleRejected)},
		{44, string(markets.MarketLifecycleCancelled)},
		{45, string(markets.MarketLifecycleClosed)},
	}
	for _, item := range hidden {
		market := modelstesting.GenerateMarket(item.id, "creator")
		market.QuestionTitle = "Hidden " + item.lifecycle
		market.LifecycleStatus = item.lifecycle
		if err := db.Create(&market).Error; err != nil {
			t.Fatalf("create %s market: %v", item.lifecycle, err)
		}
	}

	resolved := modelstesting.GenerateMarket(46, "creator")
	resolved.QuestionTitle = "Resolved public title"
	resolved.LifecycleStatus = string(markets.MarketLifecycleResolved)
	resolved.IsResolved = true
	if err := db.Create(&resolved).Error; err != nil {
		t.Fatalf("create resolved market: %v", err)
	}

	ctx := context.Background()
	if (&markets.Market{LifecycleStatus: string(markets.MarketLifecycleProposed)}).IsTradableAt(time.Now()) {
		t.Fatal("proposed market was tradable")
	}

	assertHidden := func(id int64) {
		t.Helper()
		if _, err := service.GetMarketDetails(ctx, id); !errors.Is(err, markets.ErrMarketNotFound) {
			t.Fatalf("GetMarketDetails(%d) error = %v, want ErrMarketNotFound", id, err)
		}
		if _, err := service.GetMarketBetsPage(ctx, id, markets.Page{}); !errors.Is(err, markets.ErrMarketNotFound) {
			t.Fatalf("GetMarketBetsPage(%d) error = %v, want ErrMarketNotFound", id, err)
		}
		if _, err := service.GetMarketPositions(ctx, id); !errors.Is(err, markets.ErrMarketNotFound) {
			t.Fatalf("GetMarketPositions(%d) error = %v, want ErrMarketNotFound", id, err)
		}
		if _, err := service.GetUserPositionInMarket(ctx, id, "creator"); !errors.Is(err, markets.ErrMarketNotFound) {
			t.Fatalf("GetUserPositionInMarket(%d) error = %v, want ErrMarketNotFound", id, err)
		}
		if _, err := service.GetMarketLeaderboard(ctx, id, markets.Page{}); !errors.Is(err, markets.ErrMarketNotFound) {
			t.Fatalf("GetMarketLeaderboard(%d) error = %v, want ErrMarketNotFound", id, err)
		}
		if _, err := service.ProjectProbability(ctx, markets.ProbabilityProjectionRequest{MarketID: id, Amount: 10, Outcome: "YES"}); !errors.Is(err, markets.ErrMarketNotFound) {
			t.Fatalf("ProjectProbability(%d) error = %v, want ErrMarketNotFound", id, err)
		}
		if _, err := service.GetMarketSummaryReadModel(ctx, id); !errors.Is(err, markets.ErrMarketNotFound) {
			t.Fatalf("GetMarketSummaryReadModel(%d) error = %v, want ErrMarketNotFound", id, err)
		}
	}

	assertHidden(proposed.ID)
	for _, item := range hidden {
		assertHidden(item.id)
	}

	review, err := service.GetMarket(ctx, proposed.ID)
	if err != nil {
		t.Fatalf("GetMarket(proposed) error = %v", err)
	}
	if review.QuestionTitle != secret {
		t.Fatalf("review load title = %q", review.QuestionTitle)
	}

	for _, id := range []int64{published.ID, resolved.ID} {
		details, err := service.GetMarketDetails(ctx, id)
		if err != nil {
			t.Fatalf("GetMarketDetails(%d) error = %v", id, err)
		}
		if details.Market == nil || strings.TrimSpace(details.Market.QuestionTitle) == "" {
			t.Fatalf("public market %d missing title", id)
		}
		if _, err := service.GetMarketBetsPage(ctx, id, markets.Page{}); err != nil {
			t.Fatalf("GetMarketBetsPage(%d) error = %v", id, err)
		}
	}

	approved, err := service.ApproveProposedMarket(ctx, proposed.ID, "admin", true)
	if err != nil {
		t.Fatalf("ApproveProposedMarket error = %v", err)
	}
	if approved.QuestionTitle != secret {
		t.Fatalf("approved title = %q", approved.QuestionTitle)
	}
	details, err := service.GetMarketDetails(ctx, proposed.ID)
	if err != nil {
		t.Fatalf("approved market details error = %v", err)
	}
	if details.Market.QuestionTitle != secret {
		t.Fatalf("public title after approval = %q", details.Market.QuestionTitle)
	}
}
