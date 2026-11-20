package marketshandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"socialpredict/handlers"
	"socialpredict/handlers/markets/dto"
	dmarkets "socialpredict/internal/domain/markets"
	dusers "socialpredict/internal/domain/users"

	"github.com/gorilla/mux"
)

func TestGetMarketGroupIncludesChildProbabilityHistoryAndAmendments(t *testing.T) {
	now := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)
	service := &MockService{
		MarketGroupOverviewFn: func(_ context.Context, groupID int64) (*dmarkets.MarketGroupOverview, error) {
			if groupID != 9 {
				t.Fatalf("expected group id 9, got %d", groupID)
			}
			return &dmarkets.MarketGroupOverview{
				Group: &dmarkets.MarketGroup{
					ID:                 9,
					QuestionTitle:      "Spain vs Canada",
					Description:        "Grouped market",
					GroupType:          dmarkets.MarketGroupTypeMultipleChoiceBinary,
					ProbabilityPolicy:  dmarkets.MarketGroupProbabilityPolicyIndependentBinary,
					ResolutionPolicy:   dmarkets.MarketGroupResolutionPolicyIndependentChildren,
					LifecycleStatus:    dmarkets.MarketLifecyclePublished,
					CreatorUsername:    "moderator",
					StewardUsername:    "moderator",
					ResolutionDateTime: now.Add(24 * time.Hour),
					CreatedAt:          now,
					UpdatedAt:          now,
					Members: []dmarkets.MarketGroupMember{
						{ID: 1, GroupID: 9, MarketID: 101, AnswerLabel: "Spain", DisplayOrder: 0},
					},
				},
				Creator: &dmarkets.CreatorSummary{Username: "moderator"},
				Answers: []dmarkets.MarketGroupAnswerOverview{
					{
						Member: dmarkets.MarketGroupMember{ID: 1, GroupID: 9, MarketID: 101, AnswerLabel: "Spain", DisplayOrder: 0},
						Overview: &dmarkets.MarketOverview{
							Market: &dmarkets.Market{
								ID:                 101,
								QuestionTitle:      "Spain vs Canada - Spain",
								Description:        "Child market",
								OutcomeType:        "BINARY",
								ResolutionDateTime: now.Add(24 * time.Hour),
								CreatorUsername:    "moderator",
								StewardUsername:    "moderator",
								YesLabel:           "YES",
								NoLabel:            "NO",
								Status:             dmarkets.MarketStatusActive,
								LifecycleStatus:    dmarkets.MarketLifecyclePublished,
								CreatedAt:          now,
								UpdatedAt:          now,
								InitialProbability: 0.5,
							},
							Creator:         &dmarkets.CreatorSummary{Username: "moderator"},
							LastProbability: 0.64,
							ProbabilityChanges: []dmarkets.ProbabilityPoint{
								{Probability: 0.50, Timestamp: now},
								{Probability: 0.64, Timestamp: now.Add(time.Minute)},
							},
							DescriptionAmendments: []dmarkets.MarketDescriptionAmendment{
								{
									ID:         77,
									MarketID:   101,
									Version:    2,
									Body:       "Clarifies resolution source.",
									BodyFormat: dmarkets.DescriptionAmendmentFormatMarkdownLite,
									Status:     dmarkets.DescriptionAmendmentStatusApproved,
									CreatedBy:  "moderator",
									CreatedAt:  now,
									UpdatedAt:  now,
								},
							},
						},
					},
				},
			}, nil
		},
	}

	req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/market-groups/9", nil), map[string]string{"id": "9"})
	rec := httptest.NewRecorder()

	NewHandler(service, nil, nil).GetMarketGroup(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response dto.MarketGroupDetailsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(response.Answers))
	}
	answer := response.Answers[0]
	if len(answer.ProbabilityChanges) != 2 {
		t.Fatalf("expected 2 probability points, got %d", len(answer.ProbabilityChanges))
	}
	if got := answer.ProbabilityChanges[1].Probability; got != 0.64 {
		t.Fatalf("expected latest probability 0.64, got %v", got)
	}
	if len(answer.DescriptionAmendments) != 1 {
		t.Fatalf("expected 1 description amendment, got %d", len(answer.DescriptionAmendments))
	}
	if got := answer.DescriptionAmendments[0].Body; got != "Clarifies resolution source." {
		t.Fatalf("expected amendment body to round trip, got %q", got)
	}
}

func TestResolveMarketGroupHandlerAcceptsManualChildResolutions(t *testing.T) {
	var captured dmarkets.MarketGroupResolveRequest
	service := &MockService{
		ResolveMarketGroupFn: func(_ context.Context, groupID int64, req dmarkets.MarketGroupResolveRequest, username string) (*dmarkets.MarketGroup, error) {
			if groupID != 9 || username != "moderator" {
				t.Fatalf("unexpected args groupID=%d username=%q", groupID, username)
			}
			captured = req
			return &dmarkets.MarketGroup{
				ID:              groupID,
				QuestionTitle:   "Grouped market",
				LifecycleStatus: dmarkets.MarketLifecycleResolved,
				Members: []dmarkets.MarketGroupMember{
					{MarketID: 101, AnswerLabel: "Home", DisplayOrder: 0},
					{MarketID: 102, AnswerLabel: "Away", DisplayOrder: 1},
				},
			}, nil
		},
	}
	req := mux.SetURLVars(
		httptest.NewRequest(
			http.MethodPost,
			"/v0/market-groups/9/resolve",
			strings.NewReader(`{"mode":"manual","resolutions":[{"marketId":101,"resolution":"YES"},{"marketId":102,"resolution":"NO"}]}`),
		),
		map[string]string{"id": "9"},
	)
	rec := httptest.NewRecorder()

	NewHandler(service, lifecycleAuthMock{
		user: &dusers.User{Username: "moderator", UserType: string(dusers.UserTypeModerator)},
	}, nil).ResolveMarketGroup(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	expected := dmarkets.MarketGroupResolveRequest{
		Mode: "manual",
		Resolutions: []dmarkets.MarketGroupChildResolution{
			{MarketID: 101, Resolution: "YES"},
			{MarketID: 102, Resolution: "NO"},
		},
	}
	if !reflect.DeepEqual(captured, expected) {
		t.Fatalf("captured = %+v, want %+v", captured, expected)
	}
}

func TestResolveMarketGroupHandlerReportsUnpublishedChild(t *testing.T) {
	service := &MockService{
		ResolveMarketGroupFn: func(_ context.Context, groupID int64, req dmarkets.MarketGroupResolveRequest, username string) (*dmarkets.MarketGroup, error) {
			if groupID != 9 || username != "moderator" {
				t.Fatalf("unexpected args groupID=%d username=%q", groupID, username)
			}
			return nil, &dmarkets.MarketGroupChildNotPublishedError{
				MarketID:        102,
				AnswerLabel:     "Away",
				LifecycleStatus: dmarkets.MarketLifecycleProposed,
			}
		},
	}
	req := mux.SetURLVars(
		httptest.NewRequest(
			http.MethodPost,
			"/v0/market-groups/9/resolve",
			strings.NewReader(`{"mode":"exclusive_yes","winningMarketId":101}`),
		),
		map[string]string{"id": "9"},
	)
	rec := httptest.NewRecorder()

	NewHandler(service, lifecycleAuthMock{
		user: &dusers.User{Username: "moderator", UserType: string(dusers.UserTypeModerator)},
	}, nil).ResolveMarketGroup(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rec.Code, rec.Body.String())
	}
	var response handlers.FailureEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Reason != string(handlers.ReasonMarketGroupChildUnpublished) {
		t.Fatalf("reason = %q, want %q", response.Reason, handlers.ReasonMarketGroupChildUnpublished)
	}
	if response.Message == "" || !strings.Contains(response.Message, "Away") || !strings.Contains(response.Message, "not published") {
		t.Fatalf("unexpected message: %q", response.Message)
	}
	if response.Details["marketId"] != float64(102) ||
		response.Details["answerLabel"] != "Away" ||
		response.Details["lifecycleStatus"] != dmarkets.MarketLifecycleProposed {
		t.Fatalf("unexpected details: %+v", response.Details)
	}
}

func TestMarketGroupActivityHandlersHideParticipantRows(t *testing.T) {
	service := &MockService{
		PublicGroupActivityFn: func(_ context.Context, groupID int64) (dmarkets.PublicTradingActivity, error) {
			if groupID != 9 {
				t.Fatalf("expected group id 9, got %d", groupID)
			}
			return dmarkets.PublicTradingActivity{BetCount: 4, YesVolume: 30, NoVolume: 12, ParticipantCount: 2}, nil
		},
		MarketGroupBetsFn: func(context.Context, int64, dmarkets.Page) (*dmarkets.MarketGroupBetsPage, error) {
			t.Fatal("public group activity must not load named bets")
			return nil, nil
		},
	}
	routes := []string{"bets", "positions", "leaderboard"}
	for _, route := range routes {
		req := mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/market-groups/9/"+route+"?limit=21&offset=20", nil), map[string]string{"id": "9"})
		req.Header.Set("Authorization", "Bearer regular-user")
		rec := httptest.NewRecorder()
		handler := NewHandler(service, nil, nil)
		switch route {
		case "bets":
			handler.MarketGroupBets(rec, req)
		case "positions":
			handler.MarketGroupPositions(rec, req)
		default:
			handler.MarketGroupLeaderboard(rec, req)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d body=%s", route, rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "username") || strings.Contains(rec.Body.String(), "alice") || strings.Contains(rec.Body.String(), "profit") {
			t.Fatalf("%s response named a trader: %s", route, rec.Body.String())
		}
		var response handlers.SuccessEnvelope[struct {
			GroupID          int64 `json:"groupId"`
			BetCount         int   `json:"betCount"`
			YesVolume        int64 `json:"yesVolume"`
			NoVolume         int64 `json:"noVolume"`
			ParticipantCount int   `json:"participantCount"`
		}]
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode %s response: %v", route, err)
		}
		if !response.OK || response.Result.GroupID != 9 || response.Result.BetCount != 4 || response.Result.YesVolume != 30 || response.Result.NoVolume != 12 || response.Result.ParticipantCount != 2 {
			t.Fatalf("unexpected %s activity: %+v", route, response.Result)
		}
	}
}
