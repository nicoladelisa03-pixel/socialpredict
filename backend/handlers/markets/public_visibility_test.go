package marketshandlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"socialpredict/handlers"
	dmarkets "socialpredict/internal/domain/markets"
	rmarkets "socialpredict/internal/repository/markets"
	"socialpredict/models/modelstesting"

	"github.com/gorilla/mux"
)

func TestGetDetailsHidesProposedMarketFromAnonymousAndRegularUsers(t *testing.T) {
	db := modelstesting.NewFakeDB(t)
	creator := modelstesting.GenerateUser("creator", 0)
	if err := db.Create(&creator).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	secret := "Secret proposed title"
	market := modelstesting.GenerateMarket(77, "creator")
	market.QuestionTitle = secret
	market.Description = "secret description"
	market.LifecycleStatus = string(dmarkets.MarketLifecycleProposed)
	if err := db.Create(&market).Error; err != nil {
		t.Fatalf("create market: %v", err)
	}

	service := dmarkets.NewService(rmarkets.NewGormRepository(db), nil, nil, dmarkets.Config{})
	handler := NewHandler(service, nil, nil)

	for _, header := range []string{"", "Bearer regular-user"} {
		req := httptest.NewRequest(http.MethodGet, "/v0/markets/"+strconv.FormatInt(market.ID, 10), nil)
		req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(market.ID, 10)})
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		handler.GetDetails(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), secret) || strings.Contains(rec.Body.String(), "secret description") {
			t.Fatalf("hidden market body leaked details: %s", rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), string(handlers.ReasonMarketNotFound)) {
			t.Fatalf("body = %s, want MARKET_NOT_FOUND", rec.Body.String())
		}
	}

	published := modelstesting.GenerateMarket(78, "creator")
	published.QuestionTitle = "Visible approved title"
	if err := db.Create(&published).Error; err != nil {
		t.Fatalf("create published market: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/v0/markets/"+strconv.FormatInt(published.ID, 10), nil)
	req = mux.SetURLVars(req, map[string]string{"id": strconv.FormatInt(published.ID, 10)})
	rec := httptest.NewRecorder()
	handler.GetDetails(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("published status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Visible approved title") {
		t.Fatalf("published body = %s", rec.Body.String())
	}
}
