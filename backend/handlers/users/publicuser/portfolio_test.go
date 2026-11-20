package publicuser

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	"socialpredict/handlers/users/dto"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

type portfolioAuthMock struct {
	user *dusers.User
	err  *authsvc.AuthError
}

func (m portfolioAuthMock) CurrentUser(*http.Request) (*dusers.User, *authsvc.AuthError) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m portfolioAuthMock) RequireUser(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.CurrentUser(r)
}

func (m portfolioAuthMock) RequireAdmin(*http.Request) (*dusers.User, *authsvc.AuthError) {
	return nil, &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired}
}

func ownPortfolioAccount(username string) portfolioAuthMock {
	return portfolioAuthMock{user: &dusers.User{Username: username}}
}

type portfolioServiceMock struct {
	portfolio *dusers.Portfolio
	err       error
	calls     int
}

func (m *portfolioServiceMock) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

func (m *portfolioServiceMock) ApplyTransaction(context.Context, string, int64, string) error {
	return nil
}

func (m *portfolioServiceMock) GetUser(context.Context, string) (*dusers.User, error) {
	return nil, nil
}

func (m *portfolioServiceMock) GetUserCredit(context.Context, string, int64) (int64, error) {
	return 0, nil
}

func (m *portfolioServiceMock) GetUserPortfolio(context.Context, string) (*dusers.Portfolio, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.portfolio, nil
}

func (m *portfolioServiceMock) GetUserFinancials(context.Context, string) (map[string]int64, error) {
	return nil, nil
}

func (m *portfolioServiceMock) ListUserMarkets(context.Context, int64) ([]*dusers.UserMarket, error) {
	return nil, nil
}

func (m *portfolioServiceMock) UpdateDescription(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *portfolioServiceMock) UpdateDisplayName(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *portfolioServiceMock) UpdateEmoji(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *portfolioServiceMock) UpdatePersonalLinks(context.Context, string, dusers.PersonalLinks) (*dusers.User, error) {
	return nil, nil
}

func (m *portfolioServiceMock) GetPrivateProfile(context.Context, string) (*dusers.PrivateProfile, error) {
	return nil, nil
}

func (m *portfolioServiceMock) ChangePassword(context.Context, string, string, string) error {
	return nil
}

func TestGetPortfolioHandlerSuccess(t *testing.T) {
	portfolio := &dusers.Portfolio{
		Items: []dusers.PortfolioItem{
			{
				MarketID:       1,
				QuestionTitle:  "Test Market",
				YesSharesOwned: 10,
				NoSharesOwned:  5,
				LastBetPlaced:  time.Now(),
			},
		},
		TotalSharesOwned: 15,
	}
	mock := &portfolioServiceMock{portfolio: portfolio}
	handler := GetPortfolioHandler(mock, ownPortfolioAccount("alice"))

	req := httptest.NewRequest(http.MethodGet, "/v0/portfolio/alice", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body dto.PortfolioResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(body.PortfolioItems) != 1 || body.TotalSharesOwned != 15 {
		t.Fatalf("unexpected response: %+v", body)
	}
	if body.PortfolioItems[0].QuestionTitle != "Test Market" {
		t.Fatalf("expected question title 'Test Market', got %q", body.PortfolioItems[0].QuestionTitle)
	}
}

func TestGetPortfolioHandlerInvalidMethod(t *testing.T) {
	handler := GetPortfolioHandler(&portfolioServiceMock{}, ownPortfolioAccount("alice"))
	req := httptest.NewRequest(http.MethodPost, "/v0/portfolio/alice", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
	requirePortfolioFailureReason(t, rec, handlers.ReasonMethodNotAllowed)
}

func TestGetPortfolioHandlerServiceError(t *testing.T) {
	mock := &portfolioServiceMock{err: errors.New("boom")}
	handler := GetPortfolioHandler(mock, ownPortfolioAccount("alice"))

	req := httptest.NewRequest(http.MethodGet, "/v0/portfolio/alice", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	requirePortfolioFailureReason(t, rec, handlers.ReasonInternalError)
}

func TestGetPortfolioHandlerRequiresOwnAccount(t *testing.T) {
	mock := &portfolioServiceMock{portfolio: &dusers.Portfolio{TotalSharesOwned: 15}}
	handler := GetPortfolioHandler(mock, portfolioAuthMock{
		err: &authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken},
	})
	req := httptest.NewRequest(http.MethodGet, "/v0/portfolio/alice", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
	requirePortfolioFailureReason(t, rec, handlers.ReasonInvalidToken)

	handler = GetPortfolioHandler(mock, ownPortfolioAccount("bob"))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
	requirePortfolioFailureReason(t, rec, handlers.ReasonAuthorizationDenied)
	if mock.calls != 0 {
		t.Fatalf("denied portfolio read called the service %d times", mock.calls)
	}
}

func requirePortfolioFailureReason(t *testing.T, rec *httptest.ResponseRecorder, reason handlers.FailureReason) {
	t.Helper()

	var body handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal failure response: %v", err)
	}
	if body.OK || body.Reason != string(reason) {
		t.Fatalf("expected failure reason %q, got %+v", reason, body)
	}
}
