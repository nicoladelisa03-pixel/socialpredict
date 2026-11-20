package betshandlers

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

	"github.com/gorilla/mux"
)

// marketServiceStub satisfies dmarkets.ServiceInterface for tests.
type marketServiceStub struct {
	getMarketBetsFunc     func(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error)
	getMarketBetsPageFunc func(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.BetDisplayInfo, error)
}

func (m marketServiceStub) CreateMarket(context.Context, dmarkets.MarketCreateRequest, string) (*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) SetCustomLabels(context.Context, int64, string, string) error {
	panic("not implemented")
}
func (m marketServiceStub) GetMarket(context.Context, int64) (*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) ListMarkets(context.Context, dmarkets.ListFilters) ([]*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) SearchMarkets(context.Context, string, dmarkets.SearchFilters) (*dmarkets.SearchResults, error) {
	panic("not implemented")
}
func (m marketServiceStub) ResolveMarket(context.Context, int64, string, string) error {
	panic("not implemented")
}
func (m marketServiceStub) ListByStatus(context.Context, string, dmarkets.Page) ([]*dmarkets.Market, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketLeaderboard(context.Context, int64, dmarkets.Page) ([]*dmarkets.LeaderboardRow, error) {
	panic("not implemented")
}
func (m marketServiceStub) ProjectProbability(context.Context, dmarkets.ProbabilityProjectionRequest) (*dmarkets.ProbabilityProjection, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketDetails(context.Context, int64) (*dmarkets.MarketOverview, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetMarketBets(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
	if m.getMarketBetsFunc == nil {
		panic("GetMarketBets called without stub")
	}
	return m.getMarketBetsFunc(ctx, marketID)
}
func (m marketServiceStub) GetMarketBetsPage(ctx context.Context, marketID int64, p dmarkets.Page) ([]*dmarkets.BetDisplayInfo, error) {
	if m.getMarketBetsPageFunc == nil {
		panic("GetMarketBetsPage called without stub")
	}
	return m.getMarketBetsPageFunc(ctx, marketID, p)
}
func (m marketServiceStub) GetMarketPositions(context.Context, int64) (dmarkets.MarketPositions, error) {
	return dmarkets.MarketPositions{
		{Username: "alice", YesSharesOwned: 4},
	}, nil
}
func (m marketServiceStub) GetMarketPositionsPage(context.Context, int64, dmarkets.Page) (dmarkets.MarketPositions, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetUserPositionInMarket(context.Context, int64, string) (*dmarkets.UserPosition, error) {
	panic("not implemented")
}
func (m marketServiceStub) CalculateMarketVolume(context.Context, int64) (int64, error) {
	panic("not implemented")
}
func (m marketServiceStub) GetPublicMarket(context.Context, int64) (*dmarkets.PublicMarket, error) {
	panic("not implemented")
}

func TestMarketBetsHandlerWithService(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name           string
		method         string
		vars           map[string]string
		stub           marketServiceStub
		wantStatusCode int
		wantReason     string
		verifyBody     bool
	}{
		{
			name:           "non GET method rejected",
			method:         http.MethodPost,
			vars:           map[string]string{"marketId": "1"},
			stub:           marketServiceStub{},
			wantStatusCode: http.StatusMethodNotAllowed,
			wantReason:     string(handlers.ReasonMethodNotAllowed),
		},
		{
			name:           "missing market id",
			method:         http.MethodGet,
			wantStatusCode: http.StatusBadRequest,
			stub:           marketServiceStub{},
			wantReason:     string(handlers.ReasonInvalidRequest),
		},
		{
			name:           "invalid market id value",
			method:         http.MethodGet,
			vars:           map[string]string{"marketId": "abc"},
			stub:           marketServiceStub{},
			wantStatusCode: http.StatusBadRequest,
			wantReason:     string(handlers.ReasonInvalidRequest),
		},
		{
			name:   "market not found",
			method: http.MethodGet,
			vars:   map[string]string{"marketId": "42"},
			stub: marketServiceStub{
				getMarketBetsFunc: func(context.Context, int64) ([]*dmarkets.BetDisplayInfo, error) {
					return nil, dmarkets.ErrMarketNotFound
				},
			},
			wantStatusCode: http.StatusNotFound,
			wantReason:     string(handlers.ReasonMarketNotFound),
		},
		{
			name:   "invalid input from service",
			method: http.MethodGet,
			vars:   map[string]string{"marketId": "42"},
			stub: marketServiceStub{
				getMarketBetsFunc: func(context.Context, int64) ([]*dmarkets.BetDisplayInfo, error) {
					return nil, dmarkets.ErrInvalidInput
				},
			},
			wantStatusCode: http.StatusBadRequest,
			wantReason:     string(handlers.ReasonInvalidRequest),
		},
		{
			name:   "internal error bubbled up",
			method: http.MethodGet,
			vars:   map[string]string{"marketId": "42"},
			stub: marketServiceStub{
				getMarketBetsFunc: func(context.Context, int64) ([]*dmarkets.BetDisplayInfo, error) {
					return nil, errors.New("boom")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
			wantReason:     string(handlers.ReasonInternalError),
		},
		{
			name:   "successful response",
			method: http.MethodGet,
			vars:   map[string]string{"marketId": "7"},
			stub: marketServiceStub{
				getMarketBetsFunc: func(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
					if marketID != 7 {
						t.Fatalf("expected marketID 7, got %d", marketID)
					}
					return []*dmarkets.BetDisplayInfo{
						{
							Username:    "alice",
							Outcome:     "YES",
							Amount:      100,
							Probability: 0.55,
							PlacedAt:    now,
						},
					}, nil
				},
			},
			wantStatusCode: http.StatusOK,
			verifyBody:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := MarketBetsHandlerWithService(tt.stub)
			req := httptest.NewRequest(tt.method, "/v0/markets/bets/1", nil)
			if tt.vars != nil {
				req = mux.SetURLVars(req, tt.vars)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("expected status %d, got %d (body: %s)", tt.wantStatusCode, rr.Code, rr.Body.String())
			}

			if tt.verifyBody {
				if strings.Contains(rr.Body.String(), "alice") || strings.Contains(rr.Body.String(), "username") {
					t.Fatalf("public bet activity named a trader: %s", rr.Body.String())
				}
				var envelope handlers.SuccessEnvelope[dmarkets.PublicTradingActivity]
				if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if !envelope.OK {
					t.Fatalf("expected ok envelope, got %+v", envelope)
				}
				got := envelope.Result
				if got.BetCount != 1 || got.YesVolume != 100 || got.NoVolume != 0 || got.ParticipantCount != 1 {
					t.Fatalf("unexpected public activity: %+v", got)
				}
			} else if tt.wantReason != "" {
				var envelope handlers.FailureEnvelope
				if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
					t.Fatalf("failed to decode failure envelope: %v", err)
				}
				if envelope.OK || envelope.Reason != tt.wantReason {
					t.Fatalf("expected reason %q, got %+v", tt.wantReason, envelope)
				}
			}
		})
	}
}

func TestMarketBetsHandlerWithService_IgnoresNamedPage(t *testing.T) {
	handler := MarketBetsHandlerWithService(marketServiceStub{
		getMarketBetsFunc: func(ctx context.Context, marketID int64) ([]*dmarkets.BetDisplayInfo, error) {
			return []*dmarkets.BetDisplayInfo{
				{Username: "alice", Outcome: "YES", Amount: 10},
				{Username: "bob", Outcome: "NO", Amount: 4},
			}, nil
		},
		getMarketBetsPageFunc: func(context.Context, int64, dmarkets.Page) ([]*dmarkets.BetDisplayInfo, error) {
			t.Fatal("public bet activity must not return a named page")
			return nil, nil
		},
	})

	for _, header := range []string{"", "Bearer regular-user"} {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/bets/7?limit=1&offset=0", nil)
		req = mux.SetURLVars(req, map[string]string{"marketId": "7"})
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "alice") || strings.Contains(rec.Body.String(), "bob") || strings.Contains(rec.Body.String(), "username") {
			t.Fatalf("public bet activity named a trader: %s", rec.Body.String())
		}
		var envelope handlers.SuccessEnvelope[dmarkets.PublicTradingActivity]
		if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if envelope.Result.BetCount != 2 || envelope.Result.YesVolume != 10 || envelope.Result.NoVolume != 4 || envelope.Result.ParticipantCount != 1 {
			t.Fatalf("unexpected public activity: %+v", envelope.Result)
		}
	}
}
