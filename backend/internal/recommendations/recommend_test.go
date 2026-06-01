package recommendations

import (
	"testing"
	"time"
)

func TestRecommendSelectsHighestScoringCleanCandidate(t *testing.T) {
	result, err := Recommend(defaultInput(), []CardOffer{
		testOffer("Card A", RewardTypeQantasPoints, 50000, 300000, 99_00),
		testOffer("Card B", RewardTypeQantasPoints, 100000, 300000, 99_00),
	}, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if result.BestRecommendation == nil {
		t.Fatal("BestRecommendation = nil")
	}
	if result.BestRecommendation.Offer.CardName != "Card B" {
		t.Fatalf("BestRecommendation = %q, want Card B", result.BestRecommendation.Offer.CardName)
	}
	if result.BestRecommendation.Rank != 1 {
		t.Fatalf("Rank = %d, want 1", result.BestRecommendation.Rank)
	}
}

func TestRecommendKeepsIneligibleHighValueCardOutOfBest(t *testing.T) {
	input := defaultInput()
	input.CardHistory = []UserCardHistoryItem{{Issuer: "High", CardName: "High Value", CurrentlyHeld: true}}

	result, err := Recommend(input, []CardOffer{
		{Issuer: "High", CardName: "High Value", RewardType: RewardTypeQantasPoints, Network: CardNetworkVisa, SignupBonusPoints: 200000, MinimumSpendCents: 300000, SpendWindowDays: 90},
		testOffer("Clean", RewardTypeQantasPoints, 70000, 300000, 99_00),
	}, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if result.BestRecommendation == nil || result.BestRecommendation.Offer.CardName != "Clean" {
		t.Fatalf("BestRecommendation = %+v, want Clean", result.BestRecommendation)
	}
	if len(result.IneligibleOrCautionCards) == 0 {
		t.Fatal("IneligibleOrCautionCards is empty, want high-value ineligible card")
	}
}

func TestRecommendKeepsUnlikelySpendCardOutOfBest(t *testing.T) {
	input := defaultInput()
	input.MonthlySpendCents = 50_000

	result, err := Recommend(input, []CardOffer{
		testOffer("Too Hard", RewardTypeQantasPoints, 200000, 1_000_000, 99_00),
		testOffer("Achievable", RewardTypeQantasPoints, 50000, 100_000, 99_00),
	}, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if result.BestRecommendation == nil || result.BestRecommendation.Offer.CardName != "Achievable" {
		t.Fatalf("BestRecommendation = %+v, want Achievable", result.BestRecommendation)
	}
}

func TestRecommendReturnsAlternatives(t *testing.T) {
	result, err := Recommend(defaultInput(), []CardOffer{
		testOffer("Best", RewardTypeQantasPoints, 100000, 300000, 99_00),
		testOffer("Alt 1", RewardTypeQantasPoints, 90000, 300000, 99_00),
		testOffer("Alt 2", RewardTypeQantasPoints, 80000, 300000, 99_00),
		testOffer("Alt 3", RewardTypeQantasPoints, 70000, 300000, 99_00),
		testOffer("Alt 4", RewardTypeQantasPoints, 60000, 300000, 99_00),
	}, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if len(result.Alternatives) != 3 {
		t.Fatalf("len(Alternatives) = %d, want 3", len(result.Alternatives))
	}
	if result.Alternatives[0].Rank != 2 {
		t.Fatalf("first alternative rank = %d, want 2", result.Alternatives[0].Rank)
	}
}

func TestRecommendSendsManualReviewToCaution(t *testing.T) {
	result, err := Recommend(defaultInput(), []CardOffer{
		{
			Issuer:            "Manual",
			CardName:          "Manual Review",
			RewardType:        RewardTypeQantasPoints,
			Network:           CardNetworkVisa,
			SignupBonusPoints: 100000,
			MinimumSpendCents: 300000,
			SpendWindowDays:   90,
			EligibilityRules:  []EligibilityRule{{Type: "manual_review", Description: "Check terms."}},
		},
		testOffer("Clean", RewardTypeQantasPoints, 70000, 300000, 99_00),
	}, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if result.BestRecommendation == nil || result.BestRecommendation.Offer.CardName != "Clean" {
		t.Fatalf("BestRecommendation = %+v, want Clean", result.BestRecommendation)
	}
	if len(result.IneligibleOrCautionCards) != 1 {
		t.Fatalf("len(IneligibleOrCautionCards) = %d, want 1", len(result.IneligibleOrCautionCards))
	}
}

func TestRecommendConsolidatesReasonsAndWarnings(t *testing.T) {
	result, err := Recommend(defaultInput(), []CardOffer{
		testOffer("Best", RewardTypeQantasPoints, 100000, 300000, 99_00),
	}, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	candidate := result.BestRecommendation
	if candidate == nil {
		t.Fatal("BestRecommendation = nil")
	}
	if len(candidate.Reasons) < 3 {
		t.Fatalf("len(Reasons) = %d, want at least 3: %#v", len(candidate.Reasons), candidate.Reasons)
	}
	if candidate.ValueBreakdown.NetEstimatedValueCents != result.Summary.EstimatedYearOneValueCents {
		t.Fatalf("EstimatedYearOneValueCents = %d, want best net value %d", result.Summary.EstimatedYearOneValueCents, candidate.ValueBreakdown.NetEstimatedValueCents)
	}
}

func TestRecommendNoOffers(t *testing.T) {
	result, err := Recommend(defaultInput(), nil, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if result.BestRecommendation != nil {
		t.Fatalf("BestRecommendation = %+v, want nil", result.BestRecommendation)
	}
	if result.Summary.CardsConsidered != 0 {
		t.Fatalf("CardsConsidered = %d, want 0", result.Summary.CardsConsidered)
	}
}

func TestRecommendUnknownRewardTypeReturnsError(t *testing.T) {
	_, err := Recommend(defaultInput(), []CardOffer{
		testOffer("Bad", RewardType("unknown"), 100000, 300000, 99_00),
	}, testNow())
	if err == nil {
		t.Fatal("Recommend() error = nil, want error")
	}
}

func TestRecommendDeterministicTieBreak(t *testing.T) {
	result, err := Recommend(defaultInput(), []CardOffer{
		{Issuer: "B Bank", CardName: "Same", RewardType: RewardTypeQantasPoints, Network: CardNetworkVisa, SignupBonusPoints: 100000, MinimumSpendCents: 300000, SpendWindowDays: 90, AnnualFeeCents: 99_00},
		{Issuer: "A Bank", CardName: "Same", RewardType: RewardTypeQantasPoints, Network: CardNetworkVisa, SignupBonusPoints: 100000, MinimumSpendCents: 300000, SpendWindowDays: 90, AnnualFeeCents: 99_00},
	}, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	if result.BestRecommendation == nil || result.BestRecommendation.Offer.Issuer != "A Bank" {
		t.Fatalf("BestRecommendation issuer = %+v, want A Bank", result.BestRecommendation)
	}
}

func defaultInput() RecommendationInput {
	return RecommendationInput{
		OptimisationGoal:    OptimisationGoalQantas,
		MonthlySpendCents:   200_000,
		AnnualFeePreference: AnnualFeePreferenceFlexible,
		AcceptsAmex:         true,
	}
}

func testOffer(name string, rewardType RewardType, points int, minimumSpendCents int, annualFeeCents int) CardOffer {
	return CardOffer{
		Issuer:            "Issuer",
		CardName:          name,
		RewardType:        rewardType,
		Network:           CardNetworkVisa,
		SignupBonusPoints: points,
		MinimumSpendCents: minimumSpendCents,
		SpendWindowDays:   90,
		AnnualFeeCents:    annualFeeCents,
	}
}

func testNow() time.Time {
	return time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
}
