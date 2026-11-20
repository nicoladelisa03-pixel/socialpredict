package usershandlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// GetUserFinancialHandler returns an HTTP handler for the caller's own financial snapshot.
func GetUserFinancialHandler(svc dusers.ServiceInterface, auth authsvc.Authenticator) http.HandlerFunc {
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

		snapshot, err := svc.GetUserFinancials(r.Context(), username)
		if err != nil {
			switch err {
			case dusers.ErrUserNotFound:
				_ = handlers.WriteFailure(w, http.StatusNotFound, handlers.ReasonUserNotFound)
			default:
				_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			}
			return
		}

		if snapshot == nil {
			snapshot = make(map[string]int64)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"financial": snapshot}); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}
