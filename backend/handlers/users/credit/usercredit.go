package usercredit

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

// GetUserCreditHandler returns an HTTP handler for the caller's own available credit.
func GetUserCreditHandler(svc dusers.ServiceInterface, maximumDebtAllowed int64, auth authsvc.Authenticator) http.HandlerFunc {
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

		credit, err := svc.GetUserCredit(r.Context(), username, maximumDebtAllowed)
		if err != nil {
			if err == dusers.ErrUserNotFound {
				// Maintain legacy behavior: treat missing users as zero-account and return max debt.
				credit = maximumDebtAllowed
			} else {
				_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
				return
			}
		}

		response := dto.UserCreditResponse{Credit: credit}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		}
	}
}
