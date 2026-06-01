package recommendations

import (
	"cmp"
	"slices"
	"time"
)

const maxAlternatives = 3

func Recommend(input RecommendationInput, offers []CardOffer, now time.Time) (RecommendationResult, error) {
	recommendable := []RecommendationCandidate{}
	caution := []RecommendationCandidate{}
	summary := RecommendationSummary{CardsConsidered: len(offers)}

	for _, offer := range offers {
		candidate, err := buildCandidate(input, offer, now)
		if err != nil {
			return RecommendationResult{}, err
		}

		if candidate.Eligibility.Eligible {
			summary.EligibleCards++
		}
		if candidate.ValueBreakdown.NetEstimatedValueCents > summary.HighestNetValueCents {
			summary.HighestNetValueCents = candidate.ValueBreakdown.NetEstimatedValueCents
		}

		if isRecommendable(candidate) {
			recommendable = append(recommendable, candidate)
		} else {
			caution = append(caution, candidate)
		}
	}

	sortCandidates(recommendable)
	sortCandidates(caution)
	assignRanks(recommendable)

	result := RecommendationResult{
		Summary:                  summary,
		IneligibleOrCautionCards: caution,
	}

	if len(recommendable) == 0 {
		return result, nil
	}

	best := recommendable[0]
	result.BestRecommendation = &best
	result.Summary.EstimatedYearOneValueCents = best.ValueBreakdown.NetEstimatedValueCents

	if len(recommendable) > 1 {
		alternativeCount := min(maxAlternatives, len(recommendable)-1)
		result.Alternatives = recommendable[1 : 1+alternativeCount]
	}

	return result, nil
}

func buildCandidate(input RecommendationInput, offer CardOffer, now time.Time) (RecommendationCandidate, error) {
	eligibility := EvaluateEligibility(input, offer, now)
	value, err := CalculateValue(offer)
	if err != nil {
		return RecommendationCandidate{}, err
	}
	spend := AssessSpendRequirement(input, offer)
	score := ScoreCandidateAt(input, offer, eligibility, value, spend, now)

	return RecommendationCandidate{
		Offer:            offer,
		Score:            score.Score,
		Eligibility:      eligibility,
		ValueBreakdown:   value,
		SpendRequirement: spend,
		Reasons:          consolidateReasons(score, eligibility, spend),
		Warnings:         consolidateWarnings(score, eligibility, spend),
	}, nil
}

func isRecommendable(candidate RecommendationCandidate) bool {
	if !candidate.Eligibility.Eligible || candidate.Score <= 0 {
		return false
	}
	if candidate.SpendRequirement.Difficulty == SpendDifficultyUnlikely {
		return false
	}
	return candidate.Eligibility.Status == EligibilityHighConfidence || candidate.Eligibility.Status == EligibilityMediumConfidence
}

func sortCandidates(candidates []RecommendationCandidate) {
	slices.SortFunc(candidates, func(a RecommendationCandidate, b RecommendationCandidate) int {
		if result := cmp.Compare(b.Score, a.Score); result != 0 {
			return result
		}
		if result := cmp.Compare(b.ValueBreakdown.NetEstimatedValueCents, a.ValueBreakdown.NetEstimatedValueCents); result != 0 {
			return result
		}
		if result := cmp.Compare(eligibilityRank(a.Eligibility.Status), eligibilityRank(b.Eligibility.Status)); result != 0 {
			return result
		}
		if result := cmp.Compare(spendDifficultyRank(a.SpendRequirement.Difficulty), spendDifficultyRank(b.SpendRequirement.Difficulty)); result != 0 {
			return result
		}
		if result := cmp.Compare(a.Offer.AnnualFeeCents, b.Offer.AnnualFeeCents); result != 0 {
			return result
		}
		if result := cmp.Compare(a.Offer.Issuer, b.Offer.Issuer); result != 0 {
			return result
		}
		return cmp.Compare(a.Offer.CardName, b.Offer.CardName)
	})
}

func assignRanks(candidates []RecommendationCandidate) {
	for index := range candidates {
		candidates[index].Rank = index + 1
	}
}

func spendDifficultyRank(difficulty SpendDifficulty) int {
	switch difficulty {
	case SpendDifficultyEasy:
		return 0
	case SpendDifficultyAchievable:
		return 1
	case SpendDifficultyTight:
		return 2
	case SpendDifficultyUnlikely:
		return 3
	default:
		return 2
	}
}

func consolidateReasons(score ScoreResult, eligibility EligibilityResult, spend SpendRequirementResult) []string {
	reasons := []string{}
	reasons = appendUnique(reasons, score.Reasons...)
	reasons = appendUnique(reasons, eligibility.Reasons...)
	if spend.Reason != "" {
		reasons = appendUnique(reasons, spend.Reason)
	}
	return reasons
}

func consolidateWarnings(score ScoreResult, eligibility EligibilityResult, spend SpendRequirementResult) []string {
	warnings := []string{}
	warnings = appendUnique(warnings, eligibility.Warnings...)
	warnings = appendUnique(warnings, score.Warnings...)
	if eligibility.Status == EligibilityLowConfidence {
		warnings = appendUnique(warnings, "Eligibility confidence is low for this offer.")
	}
	if eligibility.Status == EligibilityManualReview {
		warnings = appendUnique(warnings, "Some eligibility terms require manual review.")
	}
	if spend.Difficulty == SpendDifficultyUnlikely {
		warnings = appendUnique(warnings, "Minimum spend appears unlikely based on projected spend.")
	}
	return warnings
}

func appendUnique(values []string, additions ...string) []string {
	for _, addition := range additions {
		if addition == "" || slices.Contains(values, addition) {
			continue
		}
		values = append(values, addition)
	}
	return values
}
