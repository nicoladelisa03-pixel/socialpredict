package positions

import (
	"errors"
	"net/http"
	"strconv"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	dmarkets "socialpredict/internal/domain/markets"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/logger"

	"github.com/gorilla/mux"
)

// MarketPositionsHandlerWithService returns public totals for one market.
// It does not list who holds shares.
func MarketPositionsHandlerWithService(svc dmarkets.ServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		marketID, err := parseMarketID(mux.Vars(r)["marketId"])
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}

		bets, err := svc.GetMarketBets(r.Context(), marketID)
		if err != nil {
			writePositionsError(w, marketID, err)
			return
		}
		positions, err := svc.GetMarketPositions(r.Context(), marketID)
		if err != nil {
			writePositionsError(w, marketID, err)
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, dmarkets.PublicTradingActivityFrom(bets, positions)); err != nil {
			logger.LogError("MarketPositions", "WriteResponse", err)
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

// MarketUserPositionHandlerWithService returns one account's holdings in a market.
// Only that account may read them.
func MarketUserPositionHandlerWithService(svc dmarkets.ServiceInterface, auth authsvc.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}

		marketID, username, err := parseMarketUserParams(mux.Vars(r))
		if err != nil {
			_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
			return
		}
		if !authhttp.RequireOwnAccount(w, r, auth, username) {
			return
		}

		position, err := svc.GetUserPositionInMarket(r.Context(), marketID, username)
		if err != nil {
			writeUserPositionError(w, marketID, username, err)
			return
		}

		if err := handlers.WriteResult(w, http.StatusOK, newUserPositionResponse(position)); err != nil {
			logger.LogError("MarketUserPosition", "WriteResponse", err)
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}

func parseMarketID(marketIDStr string) (int64, error) {
	if marketIDStr == "" {
		return 0, errors.New("Market ID is required")
	}

	marketID, err := strconv.ParseInt(marketIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("Invalid market ID")
	}
	return marketID, nil
}

func writePositionsError(w http.ResponseWriter, marketID int64, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case dmarkets.ErrInvalidInput:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		logger.LogError("MarketPositions", "GetMarketPositions", err)
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func parseMarketUserParams(vars map[string]string) (int64, string, error) {
	marketID, err := parseMarketID(vars["marketId"])
	if err != nil {
		return 0, "", err
	}

	username := vars["username"]
	if username == "" {
		return 0, "", errors.New("Username is required")
	}

	return marketID, username, nil
}

func writeUserPositionError(w http.ResponseWriter, marketID int64, username string, err error) {
	switch err {
	case dmarkets.ErrMarketNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonMarketNotFound)
	case dmarkets.ErrUserNotFound:
		_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
	case dmarkets.ErrInvalidInput:
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	default:
		logger.LogError("MarketUserPosition", "GetUserPositionInMarket", err)
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}
