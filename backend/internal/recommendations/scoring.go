package recommendations

import "time"

type scoreWeights struct {
	Value            int
	Spend            int
	Goal             int
	Eligibility      int
	AnnualFeeComfort int
}

// urgencyTiebreakerBonusMax is the maximum number of points added to a card's
// score when its offer is close to expiring. Urgency is intentionally not part
// of the weighted sum because a soon-to-expire mediocre offer should not
// outrank a strong, longer-window offer. It is applied as a small post-hoc
// bonus so that when two cards have similar weighted scores, the one that
// expires sooner wins the tiebreak.
const urgencyTiebreakerBonusMax = 3

// ScoreCandidate ranks a card by practical recommendation quality, not just raw
// value. A high-value card should not win if the user is unlikely to meet the
// spend requirement or has weak eligibility confidence.
func ScoreCandidate(input RecommendationInput, offer CardOffer, eligibility EligibilityResult, value ValueBreakdown, spend SpendRequirementResult) ScoreResult {
	return ScoreCandidateAt(input, offer, eligibility, value, spend, time.Now())
}

// ScoreCandidateAt accepts a clock value so expiry-sensitive scoring remains
// deterministic in tests.
func ScoreCandidateAt(input RecommendationInput, offer CardOffer, eligibility EligibilityResult, value ValueBreakdown, spend SpendRequirementResult, now time.Time) ScoreResult {
	if !eligibility.Eligible || eligibility.Status == EligibilityIneligible {
		return ScoreResult{
			Score:    0,
			Reasons:  []string{"This card is not eligible for recommendation based on the current inputs."},
			Warnings: eligibility.Warnings,
		}
	}

	weights := weightsForOptimisationGoal(input.OptimisationGoal)
	score := 0
	score += valueScore(value, weights.Value)
	score += spendScore(spend, weights.Spend)
	score += goalScore(input, offer, value, spend, eligibility, weights.Goal)
	score += eligibilityScore(eligibility, weights.Eligibility)
	score += annualFeeComfortScore(input, offer, weights.AnnualFeeComfort)
	// Urgency is applied last and capped small so it can only break ties
	// between otherwise comparable cards, not promote a weak offer over a
	// strong one simply because it expires sooner.
	score += urgencyTiebreakerBonus(offer, now)

	reasons := scoreReasons(input, offer, value, spend, eligibility)
	warnings := append([]string{}, eligibility.Warnings...)
	if spend.Difficulty == SpendDifficultyUnlikely {
		warnings = append(warnings, "The spend requirement appears unlikely based on projected spend.")
	}

	return ScoreResult{
		Score:    clamp(score, 0, 100),
		Reasons:  reasons,
		Warnings: warnings,
	}
}

// OptimisationGoal changes the user's definition of "best". For example,
// low_effort favours easy spend requirements and eligibility confidence, while
// max_net_value puts more weight on estimated value. Weights for each goal
// should sum to 100 so the weighted portion of the score is a clear
// percentage; urgency is applied separately as a small tiebreaker bonus.
func weightsForOptimisationGoal(goal OptimisationGoal) scoreWeights {
	switch goal {
	case OptimisationGoalMaxNetValue:
		return scoreWeights{Value: 55, Spend: 15, Goal: 5, Eligibility: 15, AnnualFeeComfort: 10}
	case OptimisationGoalQantas, OptimisationGoalVelocity, OptimisationGoalCashback:
		return scoreWeights{Value: 30, Spend: 20, Goal: 30, Eligibility: 15, AnnualFeeComfort: 5}
	case OptimisationGoalLowEffort:
		return scoreWeights{Value: 25, Spend: 30, Goal: 5, Eligibility: 25, AnnualFeeComfort: 15}
	default:
		return scoreWeights{Value: 45, Spend: 20, Goal: 15, Eligibility: 15, AnnualFeeComfort: 5}
	}
}

func valueScore(value ValueBreakdown, weight int) int {
	if value.NetEstimatedValueCents <= 0 {
		return 0
	}
	return proportionalScore(value.NetEstimatedValueCents, 100000, weight)
}

func spendScore(spend SpendRequirementResult, weight int) int {
	switch spend.Difficulty {
	case SpendDifficultyEasy:
		return weight
	case SpendDifficultyAchievable:
		return weight * 80 / 100
	case SpendDifficultyTight:
		return weight * 40 / 100
	default:
		return 0
	}
}

func goalScore(input RecommendationInput, offer CardOffer, value ValueBreakdown, spend SpendRequirementResult, eligibility EligibilityResult, weight int) int {
	switch input.OptimisationGoal {
	case OptimisationGoalQantas:
		return rewardMatchScore(offer.RewardType == RewardTypeQantasPoints, weight)
	case OptimisationGoalVelocity:
		return rewardMatchScore(offer.RewardType == RewardTypeVelocityPoints, weight)
	case OptimisationGoalCashback:
		return rewardMatchScore(offer.RewardType == RewardTypeCashback || offer.CashbackCents > 0, weight)
	case OptimisationGoalLowEffort:
		if spend.Difficulty == SpendDifficultyEasy && eligibility.Status == EligibilityHighConfidence {
			return weight
		}
		return 0
	case OptimisationGoalMaxNetValue:
		if value.NetEstimatedValueCents > 0 {
			return weight
		}
		return 0
	default:
		return 0
	}
}

func rewardMatchScore(matches bool, weight int) int {
	if matches {
		return weight
	}
	return 0
}

func eligibilityScore(eligibility EligibilityResult, weight int) int {
	switch eligibility.Status {
	case EligibilityHighConfidence:
		return weight
	case EligibilityMediumConfidence:
		return weight * 70 / 100
	case EligibilityManualReview:
		return weight * 60 / 100
	case EligibilityLowConfidence:
		return weight * 35 / 100
	default:
		return 0
	}
}

// urgencyTiebreakerBonus returns a small bonus (capped at
// urgencyTiebreakerBonusMax) when an offer is close to expiring. It is
// intentionally tiny relative to the weighted score: it should only change the
// order of cards whose underlying recommendation quality is already similar.
// Using urgency as a full weighted axis caused soon-to-expire mediocre offers
// to outrank stronger, longer-window offers, which is the opposite of what a
// trustworthy recommendation should do.
func urgencyTiebreakerBonus(offer CardOffer, now time.Time) int {
	if offer.OfferExpiresAt == nil {
		return 0
	}

	daysUntilExpiry := int(offer.OfferExpiresAt.Sub(now).Hours() / 24)
	switch {
	case daysUntilExpiry < 0:
		return 0
	case daysUntilExpiry <= 30:
		return urgencyTiebreakerBonusMax
	case daysUntilExpiry <= 60:
		return urgencyTiebreakerBonusMax / 2
	default:
		return 0
	}
}

// strict_max is handled by eligibility. Scoring only models softer fee comfort
// preferences so valuable high-fee cards are not hidden unless the user asked
// for a strict maximum.
func annualFeeComfortScore(input RecommendationInput, offer CardOffer, weight int) int {
	if weight == 0 {
		return 0
	}

	switch input.AnnualFeePreference {
	case AnnualFeePreferenceFlexible, AnnualFeePreferenceStrictMax:
		return weight
	case AnnualFeePreferencePreferLow:
		if input.MaxAnnualFeeCents <= 0 || offer.AnnualFeeCents <= input.MaxAnnualFeeCents {
			return weight
		}
		if offer.AnnualFeeCents <= input.MaxAnnualFeeCents*150/100 {
			return weight / 2
		}
		return 0
	default:
		return weight / 2
	}
}

func scoreReasons(input RecommendationInput, offer CardOffer, value ValueBreakdown, spend SpendRequirementResult, eligibility EligibilityResult) []string {
	reasons := []string{}
	if value.NetEstimatedValueCents > 0 {
		reasons = append(reasons, "This card has positive estimated net value after annual fees and credits.")
	}
	if spend.Difficulty == SpendDifficultyEasy || spend.Difficulty == SpendDifficultyAchievable {
		reasons = append(reasons, "The minimum spend requirement appears achievable based on projected spend.")
	}
	if matchesOptimisationGoal(input.OptimisationGoal, offer) {
		reasons = append(reasons, "The reward programme matches your optimisation goal.")
	}
	if eligibility.Status == EligibilityHighConfidence {
		reasons = append(reasons, "Eligibility confidence is high based on the supplied card history.")
	}
	return reasons
}

func matchesOptimisationGoal(goal OptimisationGoal, offer CardOffer) bool {
	switch goal {
	case OptimisationGoalQantas:
		return offer.RewardType == RewardTypeQantasPoints
	case OptimisationGoalVelocity:
		return offer.RewardType == RewardTypeVelocityPoints
	case OptimisationGoalCashback:
		return offer.RewardType == RewardTypeCashback || offer.CashbackCents > 0
	default:
		return false
	}
}

func proportionalScore(value int, fullValue int, weight int) int {
	if value >= fullValue {
		return weight
	}
	return value * weight / fullValue
}

func clamp(value int, minimum int, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
