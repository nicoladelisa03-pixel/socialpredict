package startuphandlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	marketdto "socialpredict/handlers/markets/dto"
	dstartups "socialpredict/internal/domain/startups"
	authsvc "socialpredict/internal/service/auth"
)

type marketStartupRequest struct {
	StartupSlug *string `json:"startupSlug"`
}

type marketStartupResponse struct {
	MarketID    int64  `json:"marketId"`
	StartupSlug string `json:"startupSlug"`
}

type startupMarketsResponse struct {
	Markets []marketdto.PublicMarketResponse `json:"markets"`
	Total   int                              `json:"total"`
}

type marketAssociationResponse struct {
	MarketID        int64  `json:"marketId"`
	QuestionTitle   string `json:"questionTitle"`
	LifecycleStatus string `json:"lifecycleStatus"`
	StartupSlug     string `json:"startupSlug"`
}

type marketAssociationListResponse struct {
	Markets []marketAssociationResponse `json:"markets"`
	Total   int                         `json:"total"`
}

type marketLinker interface {
	SetMarketStartup(ctx context.Context, marketID int64, startupSlug string) (dstartups.MarketStartupLink, error)
	ListLinkedMarkets(ctx context.Context, slug string) ([]dstartups.LinkedMarket, error)
	ListMarketAssociations(ctx context.Context) ([]dstartups.MarketAssociation, error)
}

// SetMarketStartupHandler attaches, changes, or clears a market's startup. Only an admin may call it.
func SetMarketStartupHandler(svc marketLinker, auth authsvc.Authenticator) http.HandlerFunc {
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
		marketID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
		if err != nil || marketID <= 0 {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
			return
		}
		var req marketStartupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StartupSlug == nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		link, err := svc.SetMarketStartup(r.Context(), marketID, *req.StartupSlug)
		if err != nil {
			writeStartupError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketStartupResponse{
			MarketID:    link.MarketID,
			StartupSlug: link.StartupSlug,
		})
	}
}

// ListMarketAssociationsHandler returns every market and its startup slug.
// Only an admin may call it. Public market routes stay unchanged.
func ListMarketAssociationsHandler(svc marketLinker, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
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
		rows, err := svc.ListMarketAssociations(r.Context())
		if err != nil {
			writeStartupError(w, err)
			return
		}
		responses := make([]marketAssociationResponse, 0, len(rows))
		for _, row := range rows {
			responses = append(responses, marketAssociationResponse{
				MarketID:        row.MarketID,
				QuestionTitle:   row.QuestionTitle,
				LifecycleStatus: row.LifecycleStatus,
				StartupSlug:     row.StartupSlug,
			})
		}
		_ = handlers.WriteResult(w, http.StatusOK, marketAssociationListResponse{
			Markets: responses,
			Total:   len(responses),
		})
	}
}

// ListMarketsHandler returns the public markets linked to a startup.
func ListMarketsHandler(svc marketLinker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		markets, err := svc.ListLinkedMarkets(r.Context(), mux.Vars(r)["slug"])
		if err != nil {
			writeStartupError(w, err)
			return
		}
		responses := make([]marketdto.PublicMarketResponse, 0, len(markets))
		for _, market := range markets {
			responses = append(responses, publicMarketResponse(market))
		}
		_ = handlers.WriteResult(w, http.StatusOK, startupMarketsResponse{
			Markets: responses,
			Total:   len(responses),
		})
	}
}

func publicMarketResponse(market dstartups.LinkedMarket) marketdto.PublicMarketResponse {
	return marketdto.PublicMarketResponse{
		ID:                      market.ID,
		QuestionTitle:           market.QuestionTitle,
		Description:             market.Description,
		OutcomeType:             market.OutcomeType,
		ResolutionDateTime:      market.ResolutionDateTime,
		FinalResolutionDateTime: market.FinalResolutionDateTime,
		UTCOffset:               market.UTCOffset,
		IsResolved:              market.IsResolved,
		ResolutionResult:        market.ResolutionResult,
		InitialProbability:      market.InitialProbability,
		CreatorUsername:         market.CreatorUsername,
		StewardUsername:         market.StewardUsername,
		CreatedAt:               market.CreatedAt,
		YesLabel:                market.YesLabel,
		NoLabel:                 market.NoLabel,
	}
}
