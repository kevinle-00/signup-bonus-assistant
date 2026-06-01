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
	cardOffers repositories.CardOfferRepository
	now        func() time.Time
}

func NewHandler(cardOffers repositories.CardOfferRepository) *Handler {
	return &Handler{
		cardOffers: cardOffers,
		now:        time.Now,
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
		writeError(w, http.StatusInternalServerError, "load card offers")
		return
	}

	writeJSON(w, http.StatusOK, offers)
}

func (h *Handler) handleCreateRecommendation(w http.ResponseWriter, r *http.Request) {
	var input recommendations.RecommendationInput
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request JSON: %v", err))
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid request JSON: multiple JSON values are not supported")
		return
	}
	if err := validateRecommendationInput(input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	offers, err := h.cardOffers.ListActiveCardOffers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "load card offers")
		return
	}

	now := h.now().UTC()
	result, err := recommendations.Recommend(input, offers, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "build recommendation")
		return
	}
	roadmap := recommendations.BuildRoadmap(result, now)

	writeJSON(w, http.StatusOK, roadmap)
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

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
