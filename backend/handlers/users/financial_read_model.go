package usershandlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	analytics "socialpredict/internal/domain/analytics"
	authsvc "socialpredict/internal/service/auth"
)

type UserFinancialReadModelService interface {
	GetUserFinancialMetricReadModel(context.Context, string) (*analytics.UserFinancialMetricReadModel, error)
}

type freshnessResponse struct {
	GeneratedAt            time.Time `json:"generatedAt"`
	Source                 string    `json:"source"`
	TargetFreshnessSeconds int       `json:"targetFreshnessSeconds"`
	TransactionSafeRead    bool      `json:"transactionSafeRead"`
}

type userFinancialReadModelResponse struct {
	Username      string            `json:"username"`
	Financial     map[string]int64  `json:"financial"`
	PositionCount int               `json:"positionCount"`
	Freshness     freshnessResponse `json:"freshness"`
}

// GetUserFinancialReadModelHandler returns the caller's own financial read model.
// Logged-out visitors and other accounts, including admins, cannot read it.
func GetUserFinancialReadModelHandler(svc UserFinancialReadModelService, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		username := mux.Vars(r)["username"]
		if username == "" {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		if !authhttp.RequireOwnAccount(w, r, auth, username) {
			return
		}

		readModel, err := svc.GetUserFinancialMetricReadModel(r.Context(), username)
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		if readModel == nil {
			_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonNotFound)
			return
		}

		freshness := readModel.Freshness
		response := userFinancialReadModelResponse{
			Username:      readModel.Snapshot.Username,
			Financial:     financialReadModelSnapshotToMap(&readModel.Snapshot.Financial),
			PositionCount: readModel.Snapshot.PositionCount,
			Freshness: freshnessResponse{
				GeneratedAt:            freshness.GeneratedAt,
				Source:                 freshness.Source,
				TargetFreshnessSeconds: freshness.TargetFreshnessSeconds,
				TransactionSafeRead:    freshness.TransactionSafeRead,
			},
		}
		if err := handlers.WriteResult(w, http.StatusOK, response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func financialReadModelSnapshotToMap(snapshot *analytics.FinancialSnapshot) map[string]int64 {
	if snapshot == nil {
		return map[string]int64{}
	}
	return map[string]int64{
		"accountBalance":        snapshot.AccountBalance,
		"maximumDebtAllowed":    snapshot.MaximumDebtAllowed,
		"amountInPlay":          snapshot.AmountInPlay,
		"amountBorrowed":        snapshot.AmountBorrowed,
		"retainedEarnings":      snapshot.RetainedEarnings,
		"equity":                snapshot.Equity,
		"tradingProfits":        snapshot.TradingProfits,
		"workProfits":           snapshot.WorkProfits,
		"unrealizedWorkIncome":  snapshot.UnrealizedWorkIncome,
		"unrealizedWorkProfits": snapshot.UnrealizedWorkProfits,
		"totalProfits":          snapshot.TotalProfits,
		"amountInPlayActive":    snapshot.AmountInPlayActive,
		"totalSpent":            snapshot.TotalSpent,
		"totalSpentInPlay":      snapshot.TotalSpentInPlay,
		"realizedProfits":       snapshot.RealizedProfits,
		"potentialProfits":      snapshot.PotentialProfits,
		"realizedValue":         snapshot.RealizedValue,
		"potentialValue":        snapshot.PotentialValue,
	}
}
