package startuphandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	dstartups "socialpredict/internal/domain/startups"
	authsvc "socialpredict/internal/service/auth"
)

type fundingRequest struct {
	RoundLabel        string `json:"roundLabel"`
	AnnouncementDate  string `json:"announcementDate"`
	AmountRaisedMinor *int64 `json:"amountRaisedMinor"`
	ValuationMinor    *int64 `json:"valuationMinor"`
	ValuationBasis    string `json:"valuationBasis"`
	Currency          string `json:"currency"`
	EvidenceURL       string `json:"evidenceUrl"`
	SourceLabel       string `json:"sourceLabel"`
	Note              string `json:"note"`
}

type fundingRoundResponse struct {
	ID                int64  `json:"id"`
	RoundLabel        string `json:"roundLabel"`
	AnnouncementDate  string `json:"announcementDate"`
	AmountRaisedMinor *int64 `json:"amountRaisedMinor"`
	ValuationMinor    *int64 `json:"valuationMinor"`
	ValuationBasis    string `json:"valuationBasis"`
	Currency          string `json:"currency"`
	EvidenceURL       string `json:"evidenceUrl"`
	SourceLabel       string `json:"sourceLabel"`
	Note              string `json:"note,omitempty"`
}

type fundingHistoryResponse struct {
	Rounds                  []fundingRoundResponse `json:"rounds"`
	LatestReportedValuation *fundingRoundResponse  `json:"latestReportedValuation"`
}

type fundingManager interface {
	CreateFundingRound(ctx context.Context, slug string, input dstartups.FundingInput) (*dstartups.FundingRound, error)
	UpdateFundingRound(ctx context.Context, slug string, id int64, input dstartups.FundingInput) (*dstartups.FundingRound, error)
	ListFundingHistory(ctx context.Context, slug string) (*dstartups.FundingHistory, error)
}

// ListFundingHandler returns a startup's sourced funding history. Public read.
func ListFundingHandler(svc fundingManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		history, err := svc.ListFundingHistory(r.Context(), mux.Vars(r)["slug"])
		if err != nil {
			writeStartupError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, fundingHistoryPayload(history))
	}
}

// CreateFundingHandler stores one sourced funding round. Only an admin may call it.
func CreateFundingHandler(svc fundingManager, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if !requireAdmin(w, r, auth) {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		input, ok := decodeFunding(w, r)
		if !ok {
			return
		}
		round, err := svc.CreateFundingRound(r.Context(), mux.Vars(r)["slug"], input)
		if err != nil {
			writeStartupError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusCreated, fundingRoundPayload(*round))
	}
}

// UpdateFundingHandler corrects one funding round. Only an admin may call it.
func UpdateFundingHandler(svc fundingManager, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if !requireAdmin(w, r, auth) {
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		entryID, err := strconv.ParseInt(mux.Vars(r)["entryId"], 10, 64)
		if err != nil || entryID <= 0 {
			_ = handlers.WriteFailureWithDetails(w, http.StatusNotFound, handlers.ReasonNotFound, "Funding round not found.", nil)
			return
		}
		input, ok := decodeFunding(w, r)
		if !ok {
			return
		}
		round, err := svc.UpdateFundingRound(r.Context(), mux.Vars(r)["slug"], entryID, input)
		if err != nil {
			writeStartupError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, fundingRoundPayload(*round))
	}
}

func decodeFunding(w http.ResponseWriter, r *http.Request) (dstartups.FundingInput, bool) {
	var body fundingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return dstartups.FundingInput{}, false
	}
	return dstartups.FundingInput{
		RoundLabel:        body.RoundLabel,
		AnnouncementDate:  body.AnnouncementDate,
		AmountRaisedMinor: body.AmountRaisedMinor,
		ValuationMinor:    body.ValuationMinor,
		ValuationBasis:    body.ValuationBasis,
		Currency:          body.Currency,
		EvidenceURL:       body.EvidenceURL,
		SourceLabel:       body.SourceLabel,
		Note:              body.Note,
	}, true
}

func fundingHistoryPayload(history *dstartups.FundingHistory) fundingHistoryResponse {
	rounds := []fundingRoundResponse{}
	if history != nil {
		for _, round := range history.Rounds {
			rounds = append(rounds, fundingRoundPayload(round))
		}
	}
	var latest *fundingRoundResponse
	if history != nil && history.LatestReportedValuation != nil {
		payload := fundingRoundPayload(*history.LatestReportedValuation)
		latest = &payload
	}
	return fundingHistoryResponse{Rounds: rounds, LatestReportedValuation: latest}
}

func fundingRoundPayload(round dstartups.FundingRound) fundingRoundResponse {
	return fundingRoundResponse{
		ID:                round.ID,
		RoundLabel:        round.RoundLabel,
		AnnouncementDate:  round.AnnouncementDate,
		AmountRaisedMinor: round.AmountRaisedMinor,
		ValuationMinor:    round.ValuationMinor,
		ValuationBasis:    round.ValuationBasis,
		Currency:          round.Currency,
		EvidenceURL:       round.EvidenceURL,
		SourceLabel:       round.SourceLabel,
		Note:              round.Note,
	}
}
