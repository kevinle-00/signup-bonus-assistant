package recommendations

import "fmt"

func AssessSpendRequirement(input RecommendationInput, offer CardOffer) SpendRequirementResult {
	spendWindowMonths := ceilDiv(offer.SpendWindowDays, 30)
	projectedSpendCents := input.MonthlySpendCents*spendWindowMonths + input.ExpectedLargePurchasesNext90DaysCents

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
	case ratioPercent >= 125:
		return SpendDifficultyEasy
	case ratioPercent >= 100:
		return SpendDifficultyAchievable
	case ratioPercent >= 75:
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
