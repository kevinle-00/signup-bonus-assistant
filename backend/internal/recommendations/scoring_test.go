package recommendations

import (
	"testing"
	"time"
)

func TestScoreCandidateIneligibleScoresZero(t *testing.T) {
	got := ScoreCandidateAt(
		RecommendationInput{OptimisationGoal: OptimisationGoalMaxNetValue},
		CardOffer{},
		EligibilityResult{Eligible: false, Status: EligibilityIneligible},
		ValueBreakdown{NetEstimatedValueCents: 100000},
		SpendRequirementResult{Difficulty: SpendDifficultyEasy},
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	)

	if got.Score != 0 {
		t.Fatalf("Score = %d, want 0", got.Score)
	}
}

func TestScoreCandidateQantasGoalRewardsQantasCard(t *testing.T) {
	input := RecommendationInput{
		OptimisationGoal:    OptimisationGoalQantas,
		AnnualFeePreference: AnnualFeePreferenceFlexible,
	}
	eligibility := EligibilityResult{Eligible: true, Status: EligibilityHighConfidence}
	value := ValueBreakdown{NetEstimatedValueCents: 80000}
	spend := SpendRequirementResult{Difficulty: SpendDifficultyEasy, Achievable: true}
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	qantas := ScoreCandidateAt(input, CardOffer{RewardType: RewardTypeQantasPoints}, eligibility, value, spend, now)
	velocity := ScoreCandidateAt(input, CardOffer{RewardType: RewardTypeVelocityPoints}, eligibility, value, spend, now)

	if qantas.Score <= velocity.Score {
		t.Fatalf("Qantas score = %d, Velocity score = %d; want Qantas higher", qantas.Score, velocity.Score)
	}
}

func TestScoreCandidateLowEffortFavoursEasySpendAndHighConfidence(t *testing.T) {
	input := RecommendationInput{
		OptimisationGoal:    OptimisationGoalLowEffort,
		AnnualFeePreference: AnnualFeePreferencePreferLow,
		MaxAnnualFeeCents:   20000,
	}
	value := ValueBreakdown{NetEstimatedValueCents: 50000}
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	easy := ScoreCandidateAt(
		input,
		CardOffer{AnnualFeeCents: 10000},
		EligibilityResult{Eligible: true, Status: EligibilityHighConfidence},
		value,
		SpendRequirementResult{Difficulty: SpendDifficultyEasy, Achievable: true},
		now,
	)
	tight := ScoreCandidateAt(
		input,
		CardOffer{AnnualFeeCents: 10000},
		EligibilityResult{Eligible: true, Status: EligibilityLowConfidence},
		value,
		SpendRequirementResult{Difficulty: SpendDifficultyTight, Achievable: true},
		now,
	)

	if easy.Score <= tight.Score {
		t.Fatalf("Easy score = %d, tight score = %d; want easy higher", easy.Score, tight.Score)
	}
}

func TestScoreCandidatePreferLowAnnualFeePenalisesHighFee(t *testing.T) {
	input := RecommendationInput{
		OptimisationGoal:    OptimisationGoalMaxNetValue,
		AnnualFeePreference: AnnualFeePreferencePreferLow,
		MaxAnnualFeeCents:   20000,
	}
	eligibility := EligibilityResult{Eligible: true, Status: EligibilityHighConfidence}
	value := ValueBreakdown{NetEstimatedValueCents: 50000}
	spend := SpendRequirementResult{Difficulty: SpendDifficultyEasy, Achievable: true}
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	lowFee := ScoreCandidateAt(input, CardOffer{AnnualFeeCents: 15000}, eligibility, value, spend, now)
	highFee := ScoreCandidateAt(input, CardOffer{AnnualFeeCents: 50000}, eligibility, value, spend, now)

	if lowFee.Score <= highFee.Score {
		t.Fatalf("Low-fee score = %d, high-fee score = %d; want low-fee higher", lowFee.Score, highFee.Score)
	}
}

func TestScoreCandidateUrgencyAddsPointsForExpiringOffer(t *testing.T) {
	input := RecommendationInput{
		OptimisationGoal:    OptimisationGoalMaxNetValue,
		AnnualFeePreference: AnnualFeePreferenceFlexible,
	}
	eligibility := EligibilityResult{Eligible: true, Status: EligibilityHighConfidence}
	value := ValueBreakdown{NetEstimatedValueCents: 50000}
	spend := SpendRequirementResult{Difficulty: SpendDifficultyEasy, Achievable: true}
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	expiresSoon := now.AddDate(0, 0, 10)
	expiresLater := now.AddDate(0, 0, 90)

	soon := ScoreCandidateAt(input, CardOffer{OfferExpiresAt: &expiresSoon}, eligibility, value, spend, now)
	later := ScoreCandidateAt(input, CardOffer{OfferExpiresAt: &expiresLater}, eligibility, value, spend, now)

	if soon.Score <= later.Score {
		t.Fatalf("Soon score = %d, later score = %d; want soon higher", soon.Score, later.Score)
	}
}
