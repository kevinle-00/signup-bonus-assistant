package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kevinle-00/signup-bonus-assistant/backend/internal/recommendations"
	"github.com/kevinle-00/signup-bonus-assistant/backend/internal/repositories"
)

type Handler struct {
	cardOffers        repositories.CardOfferRepository
	recommendationRun repositories.RecommendationRunRepository
	now               func() time.Time
}

func NewHandler(cardOffers repositories.CardOfferRepository, recommendationRun repositories.RecommendationRunRepository) *Handler {
	return &Handler{
		cardOffers:        cardOffers,
		recommendationRun: recommendationRun,
		now:               time.Now,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /api/card-offers", h.handleListCardOffers)
	mux.HandleFunc("POST /api/recommendations", h.handleCreateRecommendation)
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleListCardOffers(w http.ResponseWriter, r *http.Request) {
	offers, err := h.cardOffers.ListActiveCardOffers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "card_offers_unavailable", "load card offers")
		return
	}

	writeJSON(w, http.StatusOK, offers)
}

func (h *Handler) handleCreateRecommendation(w http.ResponseWriter, r *http.Request) {
	var input recommendations.RecommendationInput
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("invalid request JSON: %v", err))
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid request JSON: multiple JSON values are not supported")
		return
	}
	if err := validateRecommendationInput(input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	offers, err := h.cardOffers.ListActiveCardOffers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "card_offers_unavailable", "load card offers")
		return
	}

	now := h.now().UTC()
	result, err := recommendations.Recommend(input, offers, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "recommendation_failed", "build recommendation")
		return
	}
	roadmap := recommendations.BuildRoadmap(result, now)
	if err := h.recommendationRun.CreateRecommendationRun(r.Context(), repositories.RecommendationRun{
		InputSnapshot:              buildRecommendationInputSnapshot(input),
		ResultSnapshot:             roadmap,
		BestCardOfferID:            bestCardOfferID(roadmap),
		EstimatedYearOneValueCents: roadmap.Summary.EstimatedYearOneValueCents,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "recommendation_run_persist_failed", "persist recommendation run")
		return
	}

	writeJSON(w, http.StatusOK, roadmap)
}

type recommendationInputSnapshot struct {
	MonthlySpendCents                     int                                 `json:"monthlySpendCents"`
	ExpectedLargePurchasesNext90DaysCents int                                 `json:"expectedLargePurchasesNext90DaysCents"`
	OptimisationGoal                      recommendations.OptimisationGoal    `json:"optimisationGoal"`
	AnnualFeePreference                   recommendations.AnnualFeePreference `json:"annualFeePreference"`
	MaxAnnualFeeCents                     int                                 `json:"maxAnnualFeeCents"`
	AcceptsAmex                           bool                                `json:"acceptsAmex"`
	SpendingCategories                    []recommendations.SpendingCategory  `json:"spendingCategories"`
	CardHistorySummary                    []cardHistorySnapshot               `json:"cardHistorySummary"`
}

type cardHistorySnapshot struct {
	Issuer        string     `json:"issuer"`
	CardName      string     `json:"cardName"`
	OpenedAt      *time.Time `json:"openedAt,omitempty"`
	ClosedAt      *time.Time `json:"closedAt,omitempty"`
	CurrentlyHeld bool       `json:"currentlyHeld"`
}

func buildRecommendationInputSnapshot(input recommendations.RecommendationInput) recommendationInputSnapshot {
	history := make([]cardHistorySnapshot, 0, len(input.CardHistory))
	for _, item := range input.CardHistory {
		history = append(history, cardHistorySnapshot{
			Issuer:        item.Issuer,
			CardName:      item.CardName,
			OpenedAt:      item.OpenedAt,
			ClosedAt:      item.ClosedAt,
			CurrentlyHeld: item.CurrentlyHeld,
		})
	}

	return recommendationInputSnapshot{
		MonthlySpendCents:                     input.MonthlySpendCents,
		ExpectedLargePurchasesNext90DaysCents: input.ExpectedLargePurchasesNext90DaysCents,
		OptimisationGoal:                      input.OptimisationGoal,
		AnnualFeePreference:                   input.AnnualFeePreference,
		MaxAnnualFeeCents:                     input.MaxAnnualFeeCents,
		AcceptsAmex:                           input.AcceptsAmex,
		SpendingCategories:                    input.SpendingCategories,
		CardHistorySummary:                    history,
	}
}

func bestCardOfferID(roadmap recommendations.RecommendationRoadmap) string {
	if roadmap.BestRecommendation == nil {
		return ""
	}
	return roadmap.BestRecommendation.Offer.ID
}

func validateRecommendationInput(input recommendations.RecommendationInput) error {
	if input.OptimisationGoal == "" {
		return errors.New("optimisationGoal is required")
	}
	if !validOptimisationGoal(input.OptimisationGoal) {
		return fmt.Errorf("unsupported optimisationGoal %q", input.OptimisationGoal)
	}
	if input.AnnualFeePreference == "" {
		return errors.New("annualFeePreference is required")
	}
	if !validAnnualFeePreference(input.AnnualFeePreference) {
		return fmt.Errorf("unsupported annualFeePreference %q", input.AnnualFeePreference)
	}
	if input.MonthlySpendCents < 0 {
		return errors.New("monthlySpendCents cannot be negative")
	}
	if input.ExpectedLargePurchasesNext90DaysCents < 0 {
		return errors.New("expectedLargePurchasesNext90DaysCents cannot be negative")
	}
	if input.MaxAnnualFeeCents < 0 {
		return errors.New("maxAnnualFeeCents cannot be negative")
	}
	if input.AnnualFeePreference == recommendations.AnnualFeePreferenceStrictMax && input.MaxAnnualFeeCents == 0 {
		return errors.New("maxAnnualFeeCents is required when annualFeePreference is strict_max")
	}
	return nil
}

func validOptimisationGoal(goal recommendations.OptimisationGoal) bool {
	switch goal {
	case recommendations.OptimisationGoalMaxNetValue,
		recommendations.OptimisationGoalQantas,
		recommendations.OptimisationGoalVelocity,
		recommendations.OptimisationGoalCashback,
		recommendations.OptimisationGoalLowEffort:
		return true
	default:
		return false
	}
}

func validAnnualFeePreference(preference recommendations.AnnualFeePreference) bool {
	switch preference {
	case recommendations.AnnualFeePreferenceStrictMax,
		recommendations.AnnualFeePreferencePreferLow,
		recommendations.AnnualFeePreferenceFlexible:
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{Error: apiError{Code: code, Message: message}})
}
