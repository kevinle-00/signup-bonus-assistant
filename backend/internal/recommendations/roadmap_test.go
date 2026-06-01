package recommendations

import "testing"

func TestBuildRoadmapFromRecommendationResult(t *testing.T) {
	result, err := Recommend(defaultInput(), []CardOffer{
		testOffer("Best", RewardTypeQantasPoints, 100000, 300000, 99_00),
		testOffer("Alternative", RewardTypeQantasPoints, 90000, 300000, 99_00),
	}, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}

	roadmap := BuildRoadmap(result, testNow())

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
	if roadmap.Summary.EstimatedYearOneValueCents != result.Summary.EstimatedYearOneValueCents {
		t.Fatalf("roadmap summary value = %d, want %d", roadmap.Summary.EstimatedYearOneValueCents, result.Summary.EstimatedYearOneValueCents)
	}
}

func TestBuildRoadmapWithoutRecommendation(t *testing.T) {
	input := defaultInput()
	input.MonthlySpendCents = 10_000

	result, err := Recommend(input, []CardOffer{
		testOffer("Too Hard", RewardTypeQantasPoints, 100000, 1_000_000, 99_00),
	}, testNow())
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}

	roadmap := BuildRoadmap(result, testNow())

	if roadmap.HasRecommendation {
		t.Fatal("HasRecommendation = true, want false")
	}
	if roadmap.BestRecommendation != nil {
		t.Fatalf("BestRecommendation = %+v, want nil", roadmap.BestRecommendation)
	}
	if len(roadmap.NoRecommendationReasons) == 0 {
		t.Fatal("NoRecommendationReasons is empty")
	}
}

func TestRecommendationServiceStyleInMemoryFlow(t *testing.T) {
	now := testNow()
	condition := "Keep card open for over 12 months."
	input := RecommendationInput{
		OptimisationGoal:                      OptimisationGoalQantas,
		MonthlySpendCents:                     250_000,
		ExpectedLargePurchasesNext90DaysCents: 100_000,
		AnnualFeePreference:                   AnnualFeePreferenceFlexible,
		AcceptsAmex:                           false,
	}
	offers := []CardOffer{
		{
			Issuer:            "A Bank",
			CardName:          "Qantas Strong",
			RewardType:        RewardTypeQantasPoints,
			Network:           CardNetworkVisa,
			SignupBonusPoints: 110000,
			MinimumSpendCents: 500000,
			SpendWindowDays:   90,
			AnnualFeeCents:    300_00,
			EligibilityRules:  []EligibilityRule{{Type: "not_held_recently", Description: "No recent A Bank rewards cards.", WindowDays: intPtr(730)}},
		},
		{
			Issuer:              "B Bank",
			CardName:            "Qantas Alternative",
			RewardType:          RewardTypeQantasPoints,
			Network:             CardNetworkVisa,
			SignupBonusPoints:   85000,
			MinimumSpendCents:   300000,
			SpendWindowDays:     90,
			AnnualFeeCents:      199_00,
			LaterBonusCondition: &condition,
		},
		{
			Issuer:            "American Express",
			CardName:          "Amex High Value",
			RewardType:        RewardTypeFlexiblePoints,
			Network:           CardNetworkAmex,
			SignupBonusPoints: 200000,
			MinimumSpendCents: 300000,
			SpendWindowDays:   90,
			AnnualFeeCents:    1450_00,
		},
	}

	result, err := Recommend(input, offers, now)
	if err != nil {
		t.Fatalf("Recommend() error = %v", err)
	}
	roadmap := BuildRoadmap(result, now)

	if !roadmap.HasRecommendation {
		t.Fatal("HasRecommendation = false, want true")
	}
	if roadmap.BestRecommendation == nil || roadmap.BestRecommendation.Offer.CardName != "Qantas Strong" {
		t.Fatalf("BestRecommendation = %+v, want Qantas Strong", roadmap.BestRecommendation)
	}
	if roadmap.Summary.EstimatedYearOneValueCents != roadmap.BestRecommendation.ValueBreakdown.NetEstimatedValueCents {
		t.Fatalf("EstimatedYearOneValueCents = %d, want best net value %d", roadmap.Summary.EstimatedYearOneValueCents, roadmap.BestRecommendation.ValueBreakdown.NetEstimatedValueCents)
	}
	if len(roadmap.Alternatives) != 1 {
		t.Fatalf("len(Alternatives) = %d, want 1", len(roadmap.Alternatives))
	}
	if len(roadmap.IneligibleOrCautionCards) != 1 {
		t.Fatalf("len(IneligibleOrCautionCards) = %d, want 1", len(roadmap.IneligibleOrCautionCards))
	}
	findChecklistItem(t, roadmap.ActionChecklist, ActionChecklistMeetSpend)
	findChecklistItem(t, roadmap.ActionChecklist, ActionChecklistAnnualFee)
}

func intPtr(value int) *int {
	return &value
}
