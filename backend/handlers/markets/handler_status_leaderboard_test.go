package marketshandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/security"

	"github.com/gorilla/mux"
)

func TestListByStatusHandler_Smoke(t *testing.T) {
	svc := &MockService{}
	svc.ListByStatusFn = func(ctx context.Context, status string, p dmarkets.Page) ([]*dmarkets.Market, error) {
		if status != "active" {
			t.Fatalf("expected status active, got %s", status)
		}
		if p.Limit != 50 {
			t.Fatalf("expected limit 50, got %d", p.Limit)
		}
		now := time.Now()
		return []*dmarkets.Market{{
			ID:                 101,
			QuestionTitle:      "Sample",
			Description:        "desc",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(24 * time.Hour),
			CreatorUsername:    "creator",
			YesLabel:           "YES",
			NoLabel:            "NO",
			Status:             "active",
			CreatedAt:          now,
			UpdatedAt:          now,
		}}, nil
	}

	handler := NewHandler(svc, nil, security.NewSecurityService())
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/status/active?limit=50", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/status/{status}", handler.ListByStatus)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Markets []json.RawMessage `json:"markets"`
		Total   int               `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Markets) != 1 {
		t.Fatalf("expected 1 market, got %d", len(resp.Markets))
	}
}

func TestMarketLeaderboardHandler_Smoke(t *testing.T) {
	svc := &MockService{}
	svc.MarketLeaderboardFn = func(context.Context, int64, dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
		t.Fatal("public leaderboard must not load named rows")
		return nil, nil
	}
	svc.PublicActivityFn = func(ctx context.Context, marketID int64) (dmarkets.PublicTradingActivity, error) {
		if marketID != 77 {
			t.Fatalf("expected marketID 77, got %d", marketID)
		}
		return dmarkets.PublicTradingActivity{BetCount: 3, YesVolume: 20, NoVolume: 5, ParticipantCount: 2}, nil
	}

	handler := NewHandler(svc, nil, security.NewSecurityService())
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/77/leaderboard?limit=25", nil)
	req.Header.Set("Authorization", "Bearer regular-user")
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/{id}/leaderboard", handler.MarketLeaderboard)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if strings.Contains(rr.Body.String(), "username") || strings.Contains(rr.Body.String(), "profit") {
		t.Fatalf("public leaderboard exposed a trader: %s", rr.Body.String())
	}

	var resp handlers.SuccessEnvelope[dmarkets.PublicTradingActivity]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.OK || resp.Result.BetCount != 3 || resp.Result.YesVolume != 20 || resp.Result.NoVolume != 5 || resp.Result.ParticipantCount != 2 {
		t.Fatalf("unexpected public activity: %+v", resp.Result)
	}
}

func TestMarketLeaderboardHandler_FailureEnvelope(t *testing.T) {
	svc := &MockService{}
	svc.PublicActivityFn = func(context.Context, int64) (dmarkets.PublicTradingActivity, error) {
		return dmarkets.PublicTradingActivity{}, errors.New("boom")
	}

	handler := NewHandler(svc, nil, security.NewSecurityService())
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/77/leaderboard", nil)
	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/v0/markets/{id}/leaderboard", handler.MarketLeaderboard)

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	var resp handlers.FailureEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal failure: %v", err)
	}
	if resp.Reason != string(handlers.ReasonInternalError) {
		t.Fatalf("expected reason %q, got %q", handlers.ReasonInternalError, resp.Reason)
	}
}

func TestListUserOwnedMarketsHandlerRequiresLogin(t *testing.T) {
	handler := ListUserOwnedMarketsHandler(&MockService{}, &contractAuthMock{
		err: &authsvc.AuthError{Kind: authsvc.ErrorKindInvalidToken, Message: "invalid token"},
	})
	req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/owned-markets", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListUserOwnedMarketsHandlerUsesOwnedByFilter(t *testing.T) {
	now := time.Now()
	svc := &MockService{}
	svc.ListLifecycleFn = func(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
		if filters.OwnedBy != "alice" {
			t.Fatalf("OwnedBy = %q, want alice", filters.OwnedBy)
		}
		if filters.Status != dmarkets.MarketStatusAll {
			t.Fatalf("Status = %q, want all", filters.Status)
		}
		if filters.Limit != 20 || filters.Offset != 5 {
			t.Fatalf("pagination = %d/%d, want 20/5", filters.Limit, filters.Offset)
		}
		return []*dmarkets.Market{{
			ID:                 45,
			QuestionTitle:      "Owned Market",
			Description:        "desc",
			OutcomeType:        "BINARY",
			ResolutionDateTime: now.Add(24 * time.Hour),
			CreatorUsername:    "bob",
			StewardUsername:    "alice",
			YesLabel:           "YES",
			NoLabel:            "NO",
			Status:             dmarkets.MarketStatusActive,
			LifecycleStatus:    dmarkets.MarketLifecyclePublished,
			CreatedAt:          now,
			UpdatedAt:          now,
		}}, nil
	}

	handler := ListUserOwnedMarketsHandler(svc, &contractAuthMock{user: &dusers.User{Username: "viewer"}})
	req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/owned-markets?limit=20&offset=5", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp handlers.SuccessEnvelope[userOwnedMarketsResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.OK || resp.Result.Total != 1 || resp.Result.Markets[0].StewardUsername != "alice" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestListUserOwnedMarketsHandlerAttachesMarketGroupMetadata(t *testing.T) {
	now := time.Now().UTC()
	group := &dmarkets.MarketGroup{
		ID:                 88,
		QuestionTitle:      "Favorite Tree",
		Description:        "Grouped favorite tree market",
		GroupType:          dmarkets.MarketGroupTypeMultipleChoiceBinary,
		LifecycleStatus:    dmarkets.MarketLifecyclePublished,
		ProposalCost:       10,
		CreatorUsername:    "testuser01",
		StewardUsername:    "testuser01",
		ApprovedBy:         "auto-approval",
		ApprovedAt:         &now,
		ResolutionDateTime: now.Add(24 * time.Hour),
		CreatedAt:          now,
		UpdatedAt:          now,
		Members: []dmarkets.MarketGroupMember{
			{MarketID: 16, AnswerLabel: "Birch", DisplayOrder: 0},
			{MarketID: 17, AnswerLabel: "Pine", DisplayOrder: 1},
			{MarketID: 18, AnswerLabel: "Oak", DisplayOrder: 2},
		},
	}
	svc := &MockService{}
	svc.ListLifecycleFn = func(ctx context.Context, filters dmarkets.ListFilters) ([]*dmarkets.Market, error) {
		if filters.OwnedBy != "testuser01" {
			t.Fatalf("OwnedBy = %q, want testuser01", filters.OwnedBy)
		}
		return []*dmarkets.Market{
			{
				ID:                 16,
				QuestionTitle:      "Favorite Tree - Birch",
				Description:        "Answer choice child market",
				OutcomeType:        "BINARY",
				ResolutionDateTime: now.Add(24 * time.Hour),
				CreatorUsername:    "testuser01",
				StewardUsername:    "testuser01",
				YesLabel:           "YES",
				NoLabel:            "NO",
				Status:             dmarkets.MarketStatusActive,
				LifecycleStatus:    dmarkets.MarketLifecyclePublished,
				CreatedAt:          now,
				UpdatedAt:          now,
			},
		}, nil
	}
	svc.MarketGroupLookupFn = func(ctx context.Context, marketID int64) (*dmarkets.MarketGroup, error) {
		if marketID != 16 {
			t.Fatalf("marketID = %d, want 16", marketID)
		}
		return group, nil
	}

	handler := ListUserOwnedMarketsHandler(svc, &contractAuthMock{user: &dusers.User{Username: "viewer"}})
	req := httptest.NewRequest(http.MethodGet, "/v0/users/testuser01/owned-markets", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "testuser01"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp handlers.SuccessEnvelope[userOwnedMarketsResponse]
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.OK || resp.Result.Total != 1 || len(resp.Result.Markets) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	market := resp.Result.Markets[0]
	if market.MarketGroup == nil {
		t.Fatalf("expected owned market response to include grouped-market metadata")
	}
	if market.MarketGroup.ID != group.ID ||
		market.MarketGroup.QuestionTitle != "Favorite Tree" ||
		market.MarketGroup.AnswerLabel != "Birch" ||
		market.MarketGroup.AnswerCount != 3 ||
		market.MarketGroup.DisplayOrder != 0 {
		t.Fatalf("unexpected market group link: %+v", market.MarketGroup)
	}
}
