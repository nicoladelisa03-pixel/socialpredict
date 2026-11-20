package metricshandlers

import (
	"context"
	"net/http"

	"socialpredict/handlers"
	analytics "socialpredict/internal/domain/analytics"
	"socialpredict/internal/domain/readmodels"
)

type GlobalLeaderboardResponse struct {
	ParticipantCount int                `json:"participantCount"`
	Freshness        *FreshnessResponse `json:"freshness,omitempty"`
}

// GetGlobalLeaderboardHandler returns an application reporting handler for the global leaderboard.
func GetGlobalLeaderboardHandler(svc GlobalLeaderboardService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count, freshness, err := globalLeaderboardSummary(r.Context(), svc)
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}

		response := GlobalLeaderboardResponse{ParticipantCount: count}
		if freshness != nil {
			converted := freshnessResponseFromDomain(*freshness)
			response.Freshness = &converted
		}
		if err := handlers.WriteResult(w, http.StatusOK, response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func globalLeaderboardSummary(ctx context.Context, svc GlobalLeaderboardService) (int, *readmodels.Freshness, error) {
	readSvc, ok := svc.(GlobalLeaderboardReadModelService)
	if !ok {
		snapshot, err := svc.ComputeGlobalLeaderboardSnapshot(ctx)
		if err != nil {
			return 0, nil, err
		}
		return leaderboardSnapshotCount(snapshot), nil, nil
	}

	readModel, err := readSvc.GetGlobalLeaderboardReadModel(ctx, 0, 0)
	if err == nil && readModel != nil {
		return leaderboardReadModelCount(readModel), &readModel.Freshness, nil
	}

	refreshed, refreshErr := readSvc.RefreshGlobalLeaderboardSnapshot(ctx)
	if refreshErr == nil && refreshed != nil {
		return leaderboardReadModelCount(refreshed), &refreshed.Freshness, nil
	}
	if readModel != nil {
		return leaderboardReadModelCount(readModel), &readModel.Freshness, nil
	}

	snapshot, computeErr := svc.ComputeGlobalLeaderboardSnapshot(ctx)
	if computeErr != nil {
		return 0, nil, computeErr
	}
	return leaderboardSnapshotCount(snapshot), nil, nil
}

func leaderboardSnapshotCount(snapshot *analytics.GlobalLeaderboardSnapshot) int {
	if snapshot == nil {
		return 0
	}
	return len(snapshot.Result())
}

func leaderboardReadModelCount(model *analytics.GlobalLeaderboardReadModel) int {
	if model == nil {
		return 0
	}
	if model.ParticipantCount > 0 || len(model.Entries) == 0 {
		return model.ParticipantCount
	}
	return len(model.Entries)
}
