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
	"github.com/kevinle-00/signup-bonus-assistant/backend/internal/repositories"
)

func TestCreateRecommendationReturnsRoadmap(t *testing.T) {
	runRepository := &fakeRecommendationRunRepository{}
	handler := NewHandler(fakeCardOfferRepository{offers: []recommendations.CardOffer{
		testAPIOffer("Best", 100000, 300000, 99_00),
		testAPIOffer("Alternative", 80000, 300000, 99_00),
	}}, runRepository)
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
	if len(runRepository.runs) != 1 {
		t.Fatalf("persisted runs = %d, want 1", len(runRepository.runs))
	}
	if runRepository.runs[0].EstimatedYearOneValueCents != roadmap.Summary.EstimatedYearOneValueCents {
		t.Fatalf("persisted value = %d, want %d", runRepository.runs[0].EstimatedYearOneValueCents, roadmap.Summary.EstimatedYearOneValueCents)
	}
}

func TestCreateRecommendationRejectsInvalidJSON(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{}, &fakeRecommendationRunRepository{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations", strings.NewReader(`{"optimisationGoal":`))

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, recorder, "invalid_request")
}

func TestCreateRecommendationRejectsInvalidInput(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{}, &fakeRecommendationRunRepository{})

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
	assertErrorCode(t, recorder, "invalid_request")
}

func TestCreateRecommendationRejectsUnsupportedEnums(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{}, &fakeRecommendationRunRepository{})

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
	assertErrorCode(t, recorder, "invalid_request")
}

func TestCreateRecommendationHandlesRepositoryError(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{err: errors.New("boom")}, &fakeRecommendationRunRepository{})

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
	assertErrorCode(t, recorder, "card_offers_unavailable")
}

func TestCreateRecommendationHandlesRecommendationRunError(t *testing.T) {
	handler := NewHandler(
		fakeCardOfferRepository{offers: []recommendations.CardOffer{testAPIOffer("Best", 100000, 300000, 99_00)}},
		&fakeRecommendationRunRepository{err: errors.New("boom")},
	)
	handler.now = fixedAPINow

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
	assertErrorCode(t, recorder, "recommendation_run_persist_failed")
}

func TestListCardOffersReturnsActiveOffers(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{offers: []recommendations.CardOffer{
		testAPIOffer("First", 100000, 300000, 99_00),
		testAPIOffer("Second", 80000, 300000, 149_00),
	}}, &fakeRecommendationRunRepository{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/card-offers", nil)

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var offers []recommendations.CardOffer
	if err := json.Unmarshal(recorder.Body.Bytes(), &offers); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(offers) != 2 {
		t.Fatalf("len(offers) = %d, want 2", len(offers))
	}
	if offers[0].CardName != "First" {
		t.Fatalf("first card = %q, want First", offers[0].CardName)
	}
}

func TestListCardOffersHandlesRepositoryError(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{err: errors.New("boom")}, &fakeRecommendationRunRepository{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/card-offers", nil)

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	assertErrorCode(t, recorder, "card_offers_unavailable")
}

func TestHealth(t *testing.T) {
	handler := NewHandler(fakeCardOfferRepository{}, &fakeRecommendationRunRepository{})

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

type fakeRecommendationRunRepository struct {
	runs []repositories.RecommendationRun
	err  error
}

func (r *fakeRecommendationRunRepository) CreateRecommendationRun(_ context.Context, run repositories.RecommendationRun) error {
	if r.err != nil {
		return r.err
	}
	r.runs = append(r.runs, run)
	return nil
}

func assertErrorCode(t *testing.T, recorder *httptest.ResponseRecorder, want string) {
	t.Helper()
	var response errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if response.Error.Code != want {
		t.Fatalf("error code = %q, want %q; body = %s", response.Error.Code, want, recorder.Body.String())
	}
}

func testAPIOffer(name string, points int, minimumSpendCents int, annualFeeCents int) recommendations.CardOffer {
	return recommendations.CardOffer{
		ID:                "39024f5a-7e50-4709-87c5-6aa56cbf2dff",
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
