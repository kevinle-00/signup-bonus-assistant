package recommendations

import "fmt"

// Real card spend rarely tracks projected spend 1:1. Most Australian issuers
// exclude government payments, BPAY, ATO, gift cards, balance transfers, and
// refunds from the minimum-spend calculation. A user whose projected spend only
// just clears the minimum is realistically at risk of missing it, so these bands
// build in a safety margin.
const (
	// 150%+ of minimum: comfortable even after typical exclusions and refunds.
	spendEasyRatioPercent = 150
	// 110%+ of minimum: meets the target with a small buffer for normal noise.
	spendAchievableRatioPercent = 110
	// 85%+ of minimum: below target, but close enough that shifted spend or one
	// planned purchase could plausibly push it over.
	spendTightRatioPercent = 85
)

func AssessSpendRequirement(input RecommendationInput, offer CardOffer) SpendRequirementResult {
	spendWindowMonths := ceilDiv(offer.SpendWindowDays, 30)
	// ExpectedLargePurchasesNext90DaysCents is a 90-day figure from onboarding.
	// Prorate it to the card's spend window so a 90-day holiday booking is not
	// credited in full against a 14- or 30-day minimum-spend window. The
	// proration ratio is capped at 1.0 because windows longer than 90 days do
	// not get to project additional large purchases we never asked the user
	// about.
	proratedLargePurchasesCents := input.ExpectedLargePurchasesNext90DaysCents * min(offer.SpendWindowDays, 90) / 90
	projectedSpendCents := input.MonthlySpendCents*spendWindowMonths + proratedLargePurchasesCents

	difficulty := spendDifficulty(projectedSpendCents, offer.MinimumSpendCents)
	achievable := difficulty != SpendDifficultyUnlikely

	return SpendRequirementResult{
		MinimumSpendCents:       offer.MinimumSpendCents,
		SpendWindowDays:         offer.SpendWindowDays,
		ProjectedUserSpendCents: projectedSpendCents,
		Achievable:              achievable,
		Difficulty:              difficulty,
		Reason:                  spendReason(difficulty, projectedSpendCents, offer.MinimumSpendCents),
	}
}

func spendDifficulty(projectedSpendCents int, minimumSpendCents int) SpendDifficulty {
	if minimumSpendCents <= 0 {
		return SpendDifficultyEasy
	}

	ratioPercent := projectedSpendCents * 100 / minimumSpendCents
	switch {
	case ratioPercent >= spendEasyRatioPercent:
		return SpendDifficultyEasy
	case ratioPercent >= spendAchievableRatioPercent:
		return SpendDifficultyAchievable
	case ratioPercent >= spendTightRatioPercent:
		return SpendDifficultyTight
	default:
		return SpendDifficultyUnlikely
	}
}

func spendReason(difficulty SpendDifficulty, projectedSpendCents int, minimumSpendCents int) string {
	switch difficulty {
	case SpendDifficultyEasy:
		return fmt.Sprintf("Projected spend of %s is comfortably above the %s minimum spend requirement.", formatCents(projectedSpendCents), formatCents(minimumSpendCents))
	case SpendDifficultyAchievable:
		return fmt.Sprintf("Projected spend of %s meets the %s minimum spend requirement.", formatCents(projectedSpendCents), formatCents(minimumSpendCents))
	case SpendDifficultyTight:
		return fmt.Sprintf("Projected spend of %s is close to the %s minimum spend requirement, so this may be tight.", formatCents(projectedSpendCents), formatCents(minimumSpendCents))
	default:
		return fmt.Sprintf("Projected spend of %s is below the %s minimum spend requirement.", formatCents(projectedSpendCents), formatCents(minimumSpendCents))
	}
}

func ceilDiv(value int, divisor int) int {
	if divisor <= 0 {
		return 0
	}
	if value <= 0 {
		return 0
	}
	return (value + divisor - 1) / divisor
}

func formatCents(cents int) string {
	dollars := cents / 100
	remainingCents := cents % 100
	return fmt.Sprintf("$%d.%02d", dollars, remainingCents)
}
