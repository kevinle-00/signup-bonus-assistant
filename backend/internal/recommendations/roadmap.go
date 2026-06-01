package recommendations

import "time"

func BuildRoadmap(result RecommendationResult, now time.Time) RecommendationRoadmap {
	roadmap := RecommendationRoadmap{
		Summary:                  result.Summary,
		Alternatives:             result.Alternatives,
		IneligibleOrCautionCards: result.IneligibleOrCautionCards,
	}

	if result.BestRecommendation == nil {
		roadmap.NoRecommendationReasons = noRecommendationReasons(result)
		return roadmap
	}

	best := *result.BestRecommendation
	roadmap.HasRecommendation = true
	roadmap.BestRecommendation = &best
	roadmap.ActionChecklist = BuildActionChecklist(best, now)
	roadmap.Reasons = appendUnique(nil, best.Reasons...)
	roadmap.Warnings = appendUnique(nil, best.Warnings...)

	return roadmap
}

func noRecommendationReasons(result RecommendationResult) []string {
	reasons := []string{"No card is safe enough to recommend from the current inputs."}
	if len(result.IneligibleOrCautionCards) > 0 {
		reasons = append(reasons, "Some cards were retained as caution cards so their risks can still be reviewed.")
	}
	return reasons
}
