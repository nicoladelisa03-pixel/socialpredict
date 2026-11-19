package startuphandlers

import (
	"bytes"
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
	"socialpredict/models"
	"socialpredict/models/modelstesting"
)

func TestMarketStartupLinkVisibilityAndCompatibility(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	published := modelstesting.GenerateMarket(11, creator.Username)
	published.QuestionTitle = "Published question"
	published.InitialProbability = 0.42
	resolved := modelstesting.GenerateMarket(12, creator.Username)
	resolved.QuestionTitle = "Resolved question"
	resolved.IsResolved = true
	resolved.ResolutionResult = "YES"
	resolved.LifecycleStatus = "resolved"
	proposed := modelstesting.GenerateMarket(13, creator.Username)
	proposed.QuestionTitle = "Proposed question"
	proposed.LifecycleStatus = "proposed"
	for _, market := range []*models.Market{&published, &resolved, &proposed} {
		if err := db.Create(market).Error; err != nil {
			t.Fatalf("seed market %d: %v", market.ID, err)
		}
	}

	svc := dstartups.NewService(startuprepo.NewGormRepository(db))
	admin := authMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}}
	denied := authMock{err: &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired, Message: "admin privileges required"}}
	anonymous := authMock{err: &authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken, Message: "Authorization header is required"}}
	link := SetMarketStartupHandler(svc, admin)

	first := createStartup(t, svc, "Northwind", "northwind")
	second := createStartup(t, svc, "Other", "other")

	patchMarket(t, link, "11", `{"startupSlug":"northwind"}`, http.StatusOK)
	patchMarket(t, link, "12", `{"startupSlug":"Northwind"}`, http.StatusOK)
	patchMarket(t, link, "13", `{"startupSlug":"northwind"}`, http.StatusOK)

	listed := listStartupMarkets(t, svc, "northwind")
	if listed.Total != 2 || len(listed.Markets) != 2 {
		t.Fatalf("public markets = %+v", listed)
	}
	if listed.Markets[0].QuestionTitle != "Published question" || listed.Markets[1].QuestionTitle != "Resolved question" {
		t.Fatalf("unexpected public order: %+v", listed.Markets)
	}
	if listed.Markets[0].InitialProbability != 0.42 || listed.Markets[1].IsResolved != true || listed.Markets[1].ResolutionResult != "YES" {
		t.Fatalf("public market fields changed: %+v", listed.Markets)
	}
	for _, market := range listed.Markets {
		if market.QuestionTitle == "Proposed question" {
			t.Fatalf("unpublished market leaked: %+v", market)
		}
	}

	patchMarket(t, link, "11", `{"startupSlug":"other"}`, http.StatusOK)
	if got := listStartupMarkets(t, svc, "northwind"); got.Total != 1 || got.Markets[0].ID != resolved.ID {
		t.Fatalf("after change, northwind markets = %+v", got)
	}
	if got := listStartupMarkets(t, svc, "other"); got.Total != 1 || got.Markets[0].ID != published.ID {
		t.Fatalf("after change, other markets = %+v", got)
	}

	cleared := httptest.NewRecorder()
	link.ServeHTTP(cleared, mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/markets/12/startup", bytes.NewBufferString(`{"startupSlug":""}`)), map[string]string{"id": "12"}))
	if cleared.Code != http.StatusOK {
		t.Fatalf("clear status = %d body=%s", cleared.Code, cleared.Body.String())
	}
	if got := listStartupMarkets(t, svc, "northwind"); got.Total != 0 {
		t.Fatalf("cleared markets = %+v", got)
	}

	forbidden := httptest.NewRecorder()
	SetMarketStartupHandler(svc, denied).ServeHTTP(forbidden, mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/markets/11/startup", bytes.NewBufferString(`{"startupSlug":"other"}`)), map[string]string{"id": "11"}))
	assertFailure(t, forbidden, http.StatusForbidden, handlers.ReasonAuthorizationDenied)
	unauthorized := httptest.NewRecorder()
	SetMarketStartupHandler(svc, anonymous).ServeHTTP(unauthorized, mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/markets/11/startup", bytes.NewBufferString(`{"startupSlug":"other"}`)), map[string]string{"id": "11"}))
	assertFailure(t, unauthorized, http.StatusUnauthorized, handlers.ReasonInvalidToken)
	patchMarket(t, link, "abc", `{"startupSlug":"other"}`, http.StatusBadRequest)
	patchMarket(t, link, "99", `{"startupSlug":"other"}`, http.StatusNotFound)
	patchMarket(t, link, "11", `{"startupSlug":"Bad Slug"}`, http.StatusBadRequest)
	patchMarket(t, link, "11", `{"startupSlug":"missing"}`, http.StatusNotFound)

	var stored models.Market
	if err := db.First(&stored, published.ID).Error; err != nil {
		t.Fatalf("reload market: %v", err)
	}
	if stored.StartupID == nil || *stored.StartupID != second.ID || stored.InitialProbability != 0.42 || stored.QuestionTitle != "Published question" {
		t.Fatalf("market trading fields changed: %+v", stored)
	}
	if first.ID == 0 {
		t.Fatalf("startup was not saved")
	}
}

func TestAdminMarketAssociationsStayOffPublicRoutes(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	svc := dstartups.NewService(startuprepo.NewGormRepository(db))
	startup := createStartup(t, svc, "Northwind", "northwind")
	proposed := modelstesting.GenerateMarket(21, creator.Username)
	proposed.QuestionTitle = "Hidden proposed title"
	proposed.LifecycleStatus = "proposed"
	published := modelstesting.GenerateMarket(22, creator.Username)
	published.QuestionTitle = "Visible published title"
	if err := db.Create(&proposed).Error; err != nil {
		t.Fatalf("seed proposed market: %v", err)
	}
	if err := db.Create(&published).Error; err != nil {
		t.Fatalf("seed published market: %v", err)
	}
	if err := db.Model(&models.Market{}).Where("id = ?", proposed.ID).Update("startup_id", startup.ID).Error; err != nil {
		t.Fatalf("link proposed market: %v", err)
	}

	admin := authMock{admin: &dusers.User{Username: "admin", UserType: string(dusers.UserTypeAdmin)}}
	handler := ListMarketAssociationsHandler(svc, admin)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v0/admin/startup-markets", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("admin list status = %d body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[marketAssociationListResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode associations: %v", err)
	}
	foundProposed := false
	foundPublished := false
	for _, market := range envelope.Result.Markets {
		if market.MarketID == proposed.ID {
			foundProposed = true
			if market.StartupSlug != "northwind" || market.QuestionTitle != "Hidden proposed title" {
				t.Fatalf("proposed association = %+v", market)
			}
		}
		if market.MarketID == published.ID {
			foundPublished = true
			if market.StartupSlug != "" {
				t.Fatalf("published market unexpectedly linked: %+v", market)
			}
		}
	}
	if !foundProposed || !foundPublished {
		t.Fatalf("associations = %+v", envelope.Result.Markets)
	}

	public := listStartupMarkets(t, svc, "northwind")
	for _, market := range public.Markets {
		if market.ID == proposed.ID || market.QuestionTitle == "Hidden proposed title" {
			t.Fatalf("public startup markets leaked a proposed market: %+v", public.Markets)
		}
	}

	denied := authMock{err: &authsvc.AuthError{Kind: authsvc.ErrorKindAdminRequired, Message: "admin privileges required"}}
	forbidden := httptest.NewRecorder()
	ListMarketAssociationsHandler(svc, denied).ServeHTTP(forbidden, httptest.NewRequest(http.MethodGet, "/v0/admin/startup-markets", nil))
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("non-admin status = %d", forbidden.Code)
	}
	anonymous := authMock{err: &authsvc.AuthError{Kind: authsvc.ErrorKindMissingToken, Message: "Authorization header is required"}}
	unauthorized := httptest.NewRecorder()
	ListMarketAssociationsHandler(svc, anonymous).ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/v0/admin/startup-markets", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous status = %d", unauthorized.Code)
	}
}

func createStartup(t *testing.T, svc *dstartups.Service, name, slug string) *dstartups.Startup {
	t.Helper()
	startup, err := svc.Create(t.Context(), dstartups.CreateRequest{Name: name, Slug: slug})
	if err != nil {
		t.Fatalf("create startup: %v", err)
	}
	return startup
}

func patchMarket(t *testing.T, handler http.Handler, id, body string, want int) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := mux.SetURLVars(httptest.NewRequest(http.MethodPatch, "/v0/admin/markets/"+id+"/startup", bytes.NewBufferString(body)), map[string]string{"id": id})
	handler.ServeHTTP(rec, req)
	if rec.Code != want {
		t.Fatalf("patch %s status = %d, want %d body=%s", id, rec.Code, want, rec.Body.String())
	}
}

func listStartupMarkets(t *testing.T, svc *dstartups.Service, slug string) startupMarketsResponse {
	t.Helper()
	rec := httptest.NewRecorder()
	ListMarketsHandler(svc).ServeHTTP(rec, mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/v0/startups/"+slug+"/markets", nil), map[string]string{"slug": slug}))
	if rec.Code != http.StatusOK {
		t.Fatalf("list markets status = %d body=%s", rec.Code, rec.Body.String())
	}
	var envelope handlers.SuccessEnvelope[startupMarketsResponse]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode markets: %v", err)
	}
	return envelope.Result
}
