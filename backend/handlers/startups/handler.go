package startuphandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/authhttp"
	dstartups "socialpredict/internal/domain/startups"
	authsvc "socialpredict/internal/service/auth"
)

type startupManager interface {
	List(ctx context.Context) ([]dstartups.Startup, error)
	GetBySlug(ctx context.Context, slug string) (*dstartups.Startup, error)
	Create(ctx context.Context, req dstartups.CreateRequest) (*dstartups.Startup, error)
	Update(ctx context.Context, slug string, req dstartups.UpdateRequest) (*dstartups.Startup, error)
}

type startupRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	WebsiteURL  string `json:"websiteUrl"`
}

type startupResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	WebsiteURL  string    `json:"websiteUrl,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type startupListResponse struct {
	Startups []startupResponse `json:"startups"`
	Total    int               `json:"total"`
}

// ListHandler returns every startup. Public read is provisional for this MVP.
func ListHandler(svc startupManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		startups, err := svc.List(r.Context())
		if err != nil {
			writeStartupError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, startupListResponse{
			Startups: startupResponses(startups),
			Total:    len(startups),
		})
	}
}

// GetHandler returns one startup by slug. Public read is provisional for this MVP.
func GetHandler(svc startupManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = handlers.WriteFailure(w, http.StatusMethodNotAllowed, handlers.ReasonMethodNotAllowed)
			return
		}
		if svc == nil {
			_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
			return
		}
		startup, err := svc.GetBySlug(r.Context(), mux.Vars(r)["slug"])
		if err != nil {
			writeStartupError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, startupResponseFromDomain(*startup))
	}
}

// CreateHandler stores a startup. Only an admin may call it.
func CreateHandler(svc startupManager, auth authsvc.Authenticator) http.HandlerFunc {
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
		req, ok := decodeStartupRequest(w, r)
		if !ok {
			return
		}
		startup, err := svc.Create(r.Context(), dstartups.CreateRequest{
			Name:        req.Name,
			Slug:        req.Slug,
			Description: req.Description,
			WebsiteURL:  req.WebsiteURL,
		})
		if err != nil {
			writeStartupError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusCreated, startupResponseFromDomain(*startup))
	}
}

// UpdateHandler edits a startup's name, description, and website. The slug stays the same.
func UpdateHandler(svc startupManager, auth authsvc.Authenticator) http.HandlerFunc {
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
		req, ok := decodeStartupRequest(w, r)
		if !ok {
			return
		}
		startup, err := svc.Update(r.Context(), mux.Vars(r)["slug"], dstartups.UpdateRequest{
			Name:        req.Name,
			Slug:        req.Slug,
			Description: req.Description,
			WebsiteURL:  req.WebsiteURL,
		})
		if err != nil {
			writeStartupError(w, err)
			return
		}
		_ = handlers.WriteResult(w, http.StatusOK, startupResponseFromDomain(*startup))
	}
}

func requireAdmin(w http.ResponseWriter, r *http.Request, auth authsvc.Authenticator) bool {
	if auth == nil {
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
		return false
	}
	if _, authErr := auth.RequireAdmin(r); authErr != nil {
		_ = authhttp.WriteFailure(w, authErr)
		return false
	}
	return true
}

func decodeStartupRequest(w http.ResponseWriter, r *http.Request) (startupRequest, bool) {
	var req startupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonInvalidRequest)
		return startupRequest{}, false
	}
	return req, true
}

func writeStartupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dstartups.ErrInvalidInput):
		_ = handlers.WriteFailure(w, http.StatusBadRequest, handlers.ReasonValidationFailed)
	case errors.Is(err, dstartups.ErrSlugTaken):
		_ = handlers.WriteFailureWithDetails(w, http.StatusConflict, handlers.ReasonValidationFailed, "A startup with this slug already exists.", map[string]any{"field": "slug"})
	case errors.Is(err, dstartups.ErrNotFound):
		_ = handlers.WriteFailureWithDetails(w, http.StatusNotFound, handlers.ReasonNotFound, "Startup not found.", nil)
	case errors.Is(err, dstartups.ErrMarketNotFound):
		_ = handlers.WriteFailureWithDetails(w, http.StatusNotFound, handlers.ReasonNotFound, "Market not found.", nil)
	case errors.Is(err, dstartups.ErrFundingNotFound):
		_ = handlers.WriteFailureWithDetails(w, http.StatusNotFound, handlers.ReasonNotFound, "Funding round not found.", nil)
	default:
		_ = handlers.WriteFailure(w, http.StatusInternalServerError, handlers.ReasonInternalError)
	}
}

func startupResponses(startups []dstartups.Startup) []startupResponse {
	responses := make([]startupResponse, 0, len(startups))
	for _, startup := range startups {
		responses = append(responses, startupResponseFromDomain(startup))
	}
	return responses
}

func startupResponseFromDomain(startup dstartups.Startup) startupResponse {
	return startupResponse{
		ID:          startup.ID,
		Name:        startup.Name,
		Slug:        startup.Slug,
		Description: startup.Description,
		WebsiteURL:  startup.WebsiteURL,
		CreatedAt:   startup.CreatedAt,
		UpdatedAt:   startup.UpdatedAt,
	}
}
