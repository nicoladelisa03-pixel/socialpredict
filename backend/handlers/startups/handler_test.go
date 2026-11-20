package startuphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	dstartups "socialpredict/internal/domain/startups"
	dusers "socialpredict/internal/domain/users"
	startuprepo "socialpredict/internal/repository/startups"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/models/modelstesting"
)

type authMock struct {
	admin *dusers.User
	err   *authsvc.AuthError
}

func (m authMock) CurrentUser(*http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.admin, m.err
}
func (m authMock) RequireUser(*http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.admin, m.err
}
func (m authMock) RequireAdmin(*http.Request) (*dusers.User, *authsvc.AuthError) {
	return m.admin, m.err
}

func TestStartupHTTPCreateListGetAndUpdate(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	svc := dstartups.NewService(startuprepo.NewGormRepository(db))
	admin := authMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}}
	denied := authMock{err: &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired, Message: "admin privileges required"}}
	anonymous := authMock{err: &authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken, Message: "Authorization header is required"}}

	create := CreateHandler(svc, admin)
	body := bytes.NewBufferString(`{"name":"Northwind","slug":"Northwind","description":"A test company","websiteUrl":"https://northwind.example"}`)
	rec := httptest.NewRecorder()
	create.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v0/admin/startups", body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}

	duplicate := httptest.NewRecorder()
	create.ServeHTTP(duplicate, httptest.NewRequest(http.MethodPost, "/v0/admin/startups", bytes.NewBufferString(`{"name":"Other","slug":"northwind"}`)))
	assertFailure(t, duplicate, http.StatusConflict, handlers.ReasonValidationFailed)

	invalid := httptest.NewRecorder()
	create.ServeHTTP(invalid, httptest.NewRequest(http.MethodPost, "/v0/admin/startups", bytes.NewBufferString(`{"name":"Bad","slug":"Bad Slug"}`)))
	assertFailure(t, invalid, http.StatusBadRequest, handlers.ReasonValidationFailed)

	forbidden := httptest.NewRecorder()
	CreateHandler(svc, denied).ServeHTTP(forbidden, httptest.NewRequest(http.MethodPost, "/v0/admin/startups", bytes.NewBufferString(`{"name":"Nope","slug":"nope"}`)))
	assertFailure(t, forbidden, http.StatusForbidden, handlers.ReasonAuthorizationDenied)

	unauthorized := httptest.NewRecorder()
	CreateHandler(svc, anonymous).ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, "/v0/admin/startups", bytes.NewBufferString(`{"name":"Nope","slug":"nope"}`)))
	assertFailure(t, unauthorized, http.StatusUnauthorized, handlers.ReasonInvalidToken)

	listRec := httptest.NewRecorder()
	ListHandler(svc).ServeHTTP(listRec, httptest.NewRequest(http.MethodGet, "/v0/startups", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listRec.Code, listRec.Body.String())
	}
	var listed handlers.SuccessEnvelope[startupListResponse]
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if !listed.OK || listed.Result.Total != 1 || listed.Result.Startups[0].Slug != "northwind" {
		t.Fatalf("unexpected list: %+v", listed.Result)
	}

	getRec := httptest.NewRecorder()
	GetHandler(svc).ServeHTTP(getRec, mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/startups/northwind", nil), map[string]string{"slug": "northwind"}))
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", getRec.Code, getRec.Body.String())
	}

	missing := httptest.NewRecorder()
	GetHandler(svc).ServeHTTP(missing, mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/startups/missing", nil), map[string]string{"slug": "missing"}))
	assertFailure(t, missing, http.StatusNotFound, handlers.ReasonNotFound)

	update := httptest.NewRecorder()
	UpdateHandler(svc, admin).ServeHTTP(update, mux.SetURLVars(
		httptest.NewRequest(http.MethodPatch, "/v0/admin/startups/northwind", bytes.NewBufferString(`{"name":"Northwind Labs","description":"Updated","websiteUrl":"https://labs.example"}`)),
		map[string]string{"slug": "northwind"},
	))
	if update.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", update.Code, update.Body.String())
	}
	var updated handlers.SuccessEnvelope[startupResponse]
	if err := json.Unmarshal(update.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.Result.Slug != "northwind" || updated.Result.Name != "Northwind Labs" || updated.Result.WebsiteURL != "https://labs.example" {
		t.Fatalf("unexpected update: %+v", updated.Result)
	}

	slugChange := httptest.NewRecorder()
	UpdateHandler(svc, denied).ServeHTTP(slugChange, mux.SetURLVars(
		httptest.NewRequest(http.MethodPatch, "/v0/admin/startups/northwind", bytes.NewBufferString(`{"name":"Other","slug":"other"}`)),
		map[string]string{"slug": "northwind"},
	))
	assertFailure(t, slugChange, http.StatusForbidden, handlers.ReasonAuthorizationDenied)

	rejectedSlug := httptest.NewRecorder()
	UpdateHandler(svc, admin).ServeHTTP(rejectedSlug, mux.SetURLVars(
		httptest.NewRequest(http.MethodPatch, "/v0/admin/startups/northwind", bytes.NewBufferString(`{"name":"Other","slug":"other"}`)),
		map[string]string{"slug": "northwind"},
	))
	assertFailure(t, rejectedSlug, http.StatusBadRequest, handlers.ReasonValidationFailed)

	stored, err := svc.GetBySlug(context.Background(), "northwind")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if stored.Slug != "northwind" || stored.Name != "Northwind Labs" {
		t.Fatalf("stored startup changed unexpectedly: %+v", stored)
	}
}

func assertFailure(t *testing.T, rec *httptest.ResponseRecorder, status int, reason handlers.FailureReason) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, status, rec.Body.String())
	}
	var envelope handlers.FailureEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode failure: %v", err)
	}
	if envelope.OK || envelope.Reason != string(reason) {
		t.Fatalf("unexpected failure: %+v", envelope)
	}
}
