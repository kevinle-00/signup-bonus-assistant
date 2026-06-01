package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kevinle-00/signup-bonus-assistant/backend/internal/recommendations"
)

func TestCreateRecommendationReturnsRoadmap(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{offers: []recommendations.CardOffer{
		testAPIOffer("Best", 100000, 300000, 99_00),
		testAPIOffer("Alternative", 80000, 300000, 99_00),
	}})
	handler.now = fixedAPINow

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations", strings.NewReader(`{
		"optimisationGoal": "qantas_points",
		"monthlySpendCents": 200000,
		"expectedLargePurchasesNext90DaysCents": 0,
		"annualFeePreference": "flexible",
		"acceptsAmex": true
	}`))

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var roadmap recommendations.RecommendationRoadmap
	if err := json.Unmarshal(recorder.Body.Bytes(), &roadmap); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !roadmap.HasRecommendation {
		t.Fatal("HasRecommendation = false, want true")
	}
	if roadmap.BestRecommendation == nil || roadmap.BestRecommendation.Offer.CardName != "Best" {
		t.Fatalf("BestRecommendation = %+v, want Best", roadmap.BestRecommendation)
	}
	if len(roadmap.ActionChecklist) == 0 {
		t.Fatal("ActionChecklist is empty")
	}
	if len(roadmap.Alternatives) != 1 {
		t.Fatalf("len(Alternatives) = %d, want 1", len(roadmap.Alternatives))
	}
}

func TestCreateRecommendationRejectsInvalidJSON(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations", strings.NewReader(`{"optimisationGoal":`))

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestCreateRecommendationRejectsInvalidInput(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations", strings.NewReader(`{
		"optimisationGoal": "qantas_points",
		"monthlySpendCents": -1,
		"annualFeePreference": "flexible",
		"acceptsAmex": true
	}`))

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestCreateRecommendationRejectsUnsupportedEnums(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations", strings.NewReader(`{
		"optimisationGoal": "points_everywhere",
		"monthlySpendCents": 200000,
		"annualFeePreference": "flexible",
		"acceptsAmex": true
	}`))

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestCreateRecommendationHandlesRepositoryError(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{err: errors.New("boom")})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations", strings.NewReader(`{
		"optimisationGoal": "qantas_points",
		"monthlySpendCents": 200000,
		"annualFeePreference": "flexible",
		"acceptsAmex": true
	}`))

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestHealth(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

type fakeCardOfferRepository struct {
	offers []recommendations.CardOffer
	err    error
}

func (r fakeCardOfferRepository) ListActiveCardOffers(context.Context) ([]recommendations.CardOffer, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.offers, nil
}

func testAPIOffer(name string, points int, minimumSpendCents int, annualFeeCents int) recommendations.CardOffer {
	return recommendations.CardOffer{
		Issuer:            "Issuer",
		CardName:          name,
		RewardProgram:     "Qantas Frequent Flyer",
		RewardType:        recommendations.RewardTypeQantasPoints,
		Network:           recommendations.CardNetworkVisa,
		SignupBonusPoints: points,
		MinimumSpendCents: minimumSpendCents,
		SpendWindowDays:   90,
		AnnualFeeCents:    annualFeeCents,
	}
}

func fixedAPINow() time.Time {
	return time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
}
