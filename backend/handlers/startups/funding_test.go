package startuphandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"

	"socialpredict/handlers"
	dstartups "socialpredict/internal/domain/startups"
	dusers "socialpredict/internal/domain/users"
	startuprepo "socialpredict/internal/repository/startups"
	authsvc "socialpredict/internal/service/auth"
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestFundingHTTPReadsAndAdminWrites(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	svc := dstartups.NewService(startuprepo.NewGormRepository(db))
	admin := authMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}}
	denied := authMock{err: &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired, Message: "admin privileges required"}}
	anonymous := authMock{err: &authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken, Message: "Authorization header is required"}}

	created, err := svc.Create(context.Background(), dstartups.CreateRequest{Name: "Northwind", Slug: "northwind"})
	if err != nil {
		t.Fatalf("startup: %v", err)
	}
	if _, err := svc.Create(context.Background(), dstartups.CreateRequest{Name: "Other", Slug: "other"}); err != nil {
		t.Fatalf("other startup: %v", err)
	}

	empty := httptest.NewRecorder()
	ListFundingHandler(svc).ServeHTTP(empty, fundingRequestWithVars(http.MethodGet, "/v0/startups/northwind/funding", "", map[string]string{"slug": "northwind"}))
	if empty.Code != http.StatusOK {
		t.Fatalf("empty status = %d body=%s", empty.Code, empty.Body.String())
	}
	var emptyBody handlers.SuccessEnvelope[fundingHistoryResponse]
	if err := json.Unmarshal(empty.Body.Bytes(), &emptyBody); err != nil {
		t.Fatalf("decode empty: %v", err)
	}
	if !emptyBody.OK || len(emptyBody.Result.Rounds) != 0 || emptyBody.Result.LatestReportedValuation != nil {
		t.Fatalf("empty history: %+v", emptyBody.Result)
	}

	missing := httptest.NewRecorder()
	ListFundingHandler(svc).ServeHTTP(missing, fundingRequestWithVars(http.MethodGet, "/v0/startups/missing/funding", "", map[string]string{"slug": "missing"}))
	assertFailure(t, missing, http.StatusNotFound, handlers.ReasonNotFound)

	forbidden := httptest.NewRecorder()
	CreateFundingHandler(svc, denied).ServeHTTP(forbidden, fundingRequestWithVars(http.MethodPost, "/v0/admin/startups/northwind/funding", validFundingJSON(nil, nil), map[string]string{"slug": "northwind"}))
	assertFailure(t, forbidden, http.StatusForbidden, handlers.ReasonAuthorizationDenied)

	unauthorized := httptest.NewRecorder()
	CreateFundingHandler(svc, anonymous).ServeHTTP(unauthorized, fundingRequestWithVars(http.MethodPost, "/v0/admin/startups/northwind/funding", validFundingJSON(nil, nil), map[string]string{"slug": "northwind"}))
	assertFailure(t, unauthorized, http.StatusUnauthorized, handlers.ReasonInvalidToken)

	invalid := httptest.NewRecorder()
	CreateFundingHandler(svc, admin).ServeHTTP(invalid, fundingRequestWithVars(http.MethodPost, "/v0/admin/startups/northwind/funding", `{"roundLabel":"Seed","announcementDate":"nope","currency":"USD","evidenceUrl":"https://example.com","sourceLabel":"Example"}`, map[string]string{"slug": "northwind"}))
	assertFailure(t, invalid, http.StatusBadRequest, handlers.ReasonValidationFailed)

	zero := httptest.NewRecorder()
	CreateFundingHandler(svc, admin).ServeHTTP(zero, fundingRequestWithVars(http.MethodPost, "/v0/admin/startups/northwind/funding", `{"roundLabel":"Seed","announcementDate":"2024-01-02","amountRaisedMinor":0,"currency":"USD","evidenceUrl":"https://example.com","sourceLabel":"Example"}`, map[string]string{"slug": "northwind"}))
	assertFailure(t, zero, http.StatusBadRequest, handlers.ReasonValidationFailed)

	createdRec := httptest.NewRecorder()
	CreateFundingHandler(svc, admin).ServeHTTP(createdRec, fundingRequestWithVars(http.MethodPost, "/v0/admin/startups/northwind/funding", validFundingJSON(int64Ptr(250050), nil), map[string]string{"slug": "northwind"}))
	if createdRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", createdRec.Code, createdRec.Body.String())
	}
	var createdBody handlers.SuccessEnvelope[fundingRoundResponse]
	if err := json.Unmarshal(createdRec.Body.Bytes(), &createdBody); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if createdBody.Result.ID == 0 || createdBody.Result.AmountRaisedMinor == nil || *createdBody.Result.AmountRaisedMinor != 250050 || createdBody.Result.ValuationMinor != nil {
		t.Fatalf("created round: %+v", createdBody.Result)
	}

	later := httptest.NewRecorder()
	CreateFundingHandler(svc, admin).ServeHTTP(later, fundingRequestWithVars(http.MethodPost, "/v0/admin/startups/northwind/funding", `{
		"roundLabel":"Series A",
		"announcementDate":"2025-06-01",
		"valuationMinor":900000,
		"valuationBasis":"post-money",
		"currency":"USD",
		"evidenceUrl":"https://example.com/series-a",
		"sourceLabel":"Example"
	}`, map[string]string{"slug": "northwind"}))
	if later.Code != http.StatusCreated {
		t.Fatalf("later status = %d body=%s", later.Code, later.Body.String())
	}
	var laterBody handlers.SuccessEnvelope[fundingRoundResponse]
	if err := json.Unmarshal(later.Body.Bytes(), &laterBody); err != nil {
		t.Fatalf("decode later: %v", err)
	}

	list := httptest.NewRecorder()
	ListFundingHandler(svc).ServeHTTP(list, fundingRequestWithVars(http.MethodGet, "/v0/startups/northwind/funding", "", map[string]string{"slug": "northwind"}))
	var listed handlers.SuccessEnvelope[fundingHistoryResponse]
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Result.Rounds) != 2 || listed.Result.Rounds[0].RoundLabel != "Seed" || listed.Result.Rounds[1].RoundLabel != "Series A" {
		t.Fatalf("order: %+v", listed.Result.Rounds)
	}
	if listed.Result.LatestReportedValuation == nil || listed.Result.LatestReportedValuation.ID != laterBody.Result.ID || listed.Result.LatestReportedValuation.ValuationBasis != "post-money" {
		t.Fatalf("latest: %+v", listed.Result.LatestReportedValuation)
	}

	updated := httptest.NewRecorder()
	seedID := strconv.FormatInt(createdBody.Result.ID, 10)
	UpdateFundingHandler(svc, admin).ServeHTTP(updated, fundingRequestWithVars(http.MethodPatch, "/v0/admin/startups/northwind/funding/"+seedID, `{
		"roundLabel":"Seed extension",
		"announcementDate":"2024-02-02",
		"amountRaisedMinor":null,
		"valuationMinor":null,
		"currency":"EUR",
		"evidenceUrl":"http://example.com/correction",
		"sourceLabel":"Corrected source",
		"note":"figure withdrawn"
	}`, map[string]string{"slug": "northwind", "entryId": seedID}))
	if updated.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", updated.Code, updated.Body.String())
	}

	var stored models.FundingRound
	if err := db.First(&stored, createdBody.Result.ID).Error; err != nil {
		t.Fatalf("reload round: %v", err)
	}
	if stored.RoundLabel != "Seed extension" || stored.AmountRaisedMinor != nil || stored.ValuationMinor != nil || stored.Currency != "EUR" || stored.Note != "figure withdrawn" {
		t.Fatalf("stored round: %+v", stored)
	}
	if stored.StartupID != created.ID {
		t.Fatalf("startup id changed: %d", stored.StartupID)
	}

	missingEntry := httptest.NewRecorder()
	UpdateFundingHandler(svc, admin).ServeHTTP(missingEntry, fundingRequestWithVars(http.MethodPatch, "/v0/admin/startups/northwind/funding/99", validFundingJSON(nil, nil), map[string]string{"slug": "northwind", "entryId": "99"}))
	assertFailure(t, missingEntry, http.StatusNotFound, handlers.ReasonNotFound)

	foreign := httptest.NewRecorder()
	laterID := strconv.FormatInt(laterBody.Result.ID, 10)
	UpdateFundingHandler(svc, admin).ServeHTTP(foreign, fundingRequestWithVars(http.MethodPatch, "/v0/admin/startups/other/funding/"+laterID, validFundingJSON(nil, nil), map[string]string{"slug": "other", "entryId": laterID}))
	assertFailure(t, foreign, http.StatusNotFound, handlers.ReasonNotFound)

	patchDenied := httptest.NewRecorder()
	UpdateFundingHandler(svc, denied).ServeHTTP(patchDenied, fundingRequestWithVars(http.MethodPatch, "/v0/admin/startups/northwind/funding/1", validFundingJSON(nil, nil), map[string]string{"slug": "northwind", "entryId": "1"}))
	assertFailure(t, patchDenied, http.StatusForbidden, handlers.ReasonAuthorizationDenied)

	patchAnon := httptest.NewRecorder()
	UpdateFundingHandler(svc, anonymous).ServeHTTP(patchAnon, fundingRequestWithVars(http.MethodPatch, "/v0/admin/startups/northwind/funding/1", validFundingJSON(nil, nil), map[string]string{"slug": "northwind", "entryId": "1"}))
	assertFailure(t, patchAnon, http.StatusUnauthorized, handlers.ReasonInvalidToken)
}

func fundingRequestWithVars(method, path, body string, vars map[string]string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	return mux.SetURLVars(req, vars)
}

func validFundingJSON(amount, valuation *int64) string {
	payload := map[string]any{
		"roundLabel":       "Seed",
		"announcementDate": "2024-01-02",
		"currency":         "USD",
		"evidenceUrl":      "https://example.com/seed",
		"sourceLabel":      "Example",
	}
	if amount != nil {
		payload["amountRaisedMinor"] = *amount
	}
	if valuation != nil {
		payload["valuationMinor"] = *valuation
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func int64Ptr(value int64) *int64 {
	return &value
}
