package usershandlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	dusers "socialpredict/internal/domain/users"
	authsvc "socialpredict/internal/service/auth"
)

type financialAuthMock struct {
	user *dusers.User
	err  *authsvc.AuthError
}

func (m financialAuthMock) CurrentUser(*http.Request) (*dusers.User, *authsvc.AuthError) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m financialAuthMock) RequireUser(r *http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.CurrentUser(r)
}

func (m financialAuthMock) RequireAdmin(*http.Request) (*dusers.User, *authsvc.AuthError) {
	return nil, &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired}
}

func ownFinancialAccount(username string) financialAuthMock {
	return financialAuthMock{user: &dusers.User{Username: username, UserType: "REGULAR"}}
}

type financialServiceMock struct {
	snapshot map[string]int64
	err      error
	calls    int
}

func (m *financialServiceMock) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

func (m *financialServiceMock) ApplyTransaction(context.Context, string, int64, string) error {
	return nil
}

func (m *financialServiceMock) GetUser(context.Context, string) (*dusers.User, error) {
	return nil, nil
}

func (m *financialServiceMock) GetUserCredit(context.Context, string, int64) (int64, error) {
	return 0, nil
}

func (m *financialServiceMock) GetUserPortfolio(context.Context, string) (*dusers.Portfolio, error) {
	return nil, nil
}

func (m *financialServiceMock) GetUserFinancials(context.Context, string) (map[string]int64, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.snapshot, nil
}

func (m *financialServiceMock) ListUserMarkets(context.Context, int64) ([]*dusers.UserMarket, error) {
	return nil, nil
}

func (m *financialServiceMock) UpdateDescription(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *financialServiceMock) UpdateDisplayName(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *financialServiceMock) UpdateEmoji(context.Context, string, string) (*dusers.User, error) {
	return nil, nil
}

func (m *financialServiceMock) UpdatePersonalLinks(context.Context, string, dusers.PersonalLinks) (*dusers.User, error) {
	return nil, nil
}

func (m *financialServiceMock) GetPrivateProfile(context.Context, string) (*dusers.PrivateProfile, error) {
	return nil, nil
}

func (m *financialServiceMock) ChangePassword(context.Context, string, string, string) error {
	return nil
}

func TestGetUserFinancialHandlerSuccess(t *testing.T) {
	mock := &financialServiceMock{snapshot: map[string]int64{"accountBalance": 500}}
	handler := GetUserFinancialHandler(mock, ownFinancialAccount("alice"))

	req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/financial", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var wrapper map[string]map[string]int64
	if err := json.Unmarshal(rec.Body.Bytes(), &wrapper); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if wrapper["financial"]["accountBalance"] != 500 {
		t.Fatalf("expected accountBalance 500, got %d", wrapper["financial"]["accountBalance"])
	}
}

func TestGetUserFinancialHandlerUserNotFound(t *testing.T) {
	handler := GetUserFinancialHandler(&financialServiceMock{err: dusers.ErrUserNotFound}, ownFinancialAccount("missing"))
	req := httptest.NewRequest(http.MethodGet, "/v0/users/missing/financial", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "missing"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	requireFinancialFailureReason(t, rec, handlers.ReasonUserNotFound)
}

func TestGetUserFinancialHandlerInternalError(t *testing.T) {
	handler := GetUserFinancialHandler(&financialServiceMock{err: errors.New("boom")}, ownFinancialAccount("alice"))
	req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/financial", nil)
	req = mux.SetURLVars(req, map[string]string{"username": "alice"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	requireFinancialFailureReason(t, rec, handlers.ReasonInternalError)
}

func TestGetUserFinancialHandlerInvalidMethod(t *testing.T) {
	handler := GetUserFinancialHandler(&financialServiceMock{}, ownFinancialAccount("alice"))
	req := httptest.NewRequest(http.MethodPost, "/v0/users/alice/financial", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
	requireFinancialFailureReason(t, rec, handlers.ReasonMethodNotAllowed)
}

func TestGetUserFinancialHandlerRequiresOwnAccount(t *testing.T) {
	mock := &financialServiceMock{snapshot: map[string]int64{"accountBalance": 500}}

	t.Run("anonymous", func(t *testing.T) {
		handler := GetUserFinancialHandler(mock, financialAuthMock{
			err: &authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken},
		})
		req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/financial", nil)
		req = mux.SetURLVars(req, map[string]string{"username": "alice"})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
		requireFinancialFailureReason(t, rec, handlers.ReasonInvalidToken)
	})

	t.Run("other user", func(t *testing.T) {
		handler := GetUserFinancialHandler(mock, ownFinancialAccount("bob"))
		req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/financial", nil)
		req = mux.SetURLVars(req, map[string]string{"username": "alice"})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
		requireFinancialFailureReason(t, rec, handlers.ReasonAuthorizationDenied)
	})

	t.Run("admin cannot read another account", func(t *testing.T) {
		handler := GetUserFinancialHandler(mock, financialAuthMock{
			user: &dusers.User{Username: "admin", UserType: "ADMIN"},
		})
		req := httptest.NewRequest(http.MethodGet, "/v0/users/alice/financial", nil)
		req = mux.SetURLVars(req, map[string]string{"username": "alice"})
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
		requireFinancialFailureReason(t, rec, handlers.ReasonAuthorizationDenied)
		if strings.Contains(rec.Body.String(), "accountBalance") || strings.Contains(rec.Body.String(), "500") {
			t.Fatalf("denial leaked financial data: %s", rec.Body.String())
		}
	})

	if mock.calls != 0 {
		t.Fatalf("denied reads called the financial service %d times", mock.calls)
	}
}

func requireFinancialFailureReason(t *testing.T, rec *httptest.ResponseRecorder, reason handlers.FailureReason) {
	t.Helper()

	var body handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal failure response: %v", err)
	}
	if body.OK || body.Reason != string(reason) {
		t.Fatalf("expected failure reason %q, got %+v", reason, body)
	}
}
