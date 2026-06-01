package recommendations

import (
	"fmt"
	"time"
)

func BuildActionChecklist(candidate RecommendationCandidate, now time.Time) []ActionChecklistItem {
	offer := candidate.Offer
	items := []ActionChecklistItem{
		{
			Kind:        ActionChecklistReviewTerms,
			Title:       "Verify the current public offer terms",
			Description: "Confirm the bonus, annual fee, minimum spend, exclusions, and eligibility terms on the issuer site before applying.",
			DueAt:       offer.OfferExpiresAt,
		},
	}

	if len(candidate.Warnings) > 0 || candidate.Eligibility.Status == EligibilityMediumConfidence {
		items = append(items, ActionChecklistItem{
			Kind:        ActionChecklistReviewRisk,
			Title:       "Review eligibility and spend cautions",
			Description: checklistSentence(candidate.Warnings, "Review the cautions attached to this card before applying."),
		})
	}

	items = append(items, ActionChecklistItem{
		Kind:        ActionChecklistApply,
		Title:       fmt.Sprintf("Apply for %s", offer.CardName),
		Description: "Apply only if the current issuer terms still match this recommendation and you are comfortable with the eligibility notes.",
		DueAt:       offer.OfferExpiresAt,
	})

	if offer.MinimumSpendCents > 0 && offer.SpendWindowDays > 0 {
		spendDueAt := now.AddDate(0, 0, offer.SpendWindowDays)
		items = append(items, ActionChecklistItem{
			Kind:        ActionChecklistMeetSpend,
			Title:       fmt.Sprintf("Spend %s on eligible purchases", formatCents(offer.MinimumSpendCents)),
			Description: fmt.Sprintf("Meet the minimum spend within %d days of approval. The date shown assumes approval today; adjust it once the card is approved.", offer.SpendWindowDays),
			DueAt:       &spendDueAt,
		})
	}

	items = append(items, ActionChecklistItem{
		Kind:        ActionChecklistTrackBonus,
		Title:       "Track bonus posting",
		Description: "Keep evidence of eligible spend and check that the bonus points or cashback post after the issuer's stated processing period.",
	})

	annualFeeReviewAt := now.AddDate(0, 11, 0)
	items = append(items, ActionChecklistItem{
		Kind:        ActionChecklistAnnualFee,
		Title:       "Review before the next annual fee",
		Description: fmt.Sprintf("The model includes a first-year annual fee of %s. Review whether to keep the card before the next annual fee is due.", formatCents(offer.AnnualFeeCents)),
		DueAt:       &annualFeeReviewAt,
	})

	if offer.LaterBonusCondition != nil && *offer.LaterBonusCondition != "" {
		items = append(items, ActionChecklistItem{
			Kind:        ActionChecklistLaterBonus,
			Title:       "Review later bonus condition separately",
			Description: fmt.Sprintf("Later bonus condition: %s. This later bonus is not included in the MVP year-one value.", *offer.LaterBonusCondition),
		})
	}

	return items
}

func checklistSentence(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}
