package recommendations

import (
	"strings"
	"time"
)

func EvaluateEligibility(input RecommendationInput, offer CardOffer, now time.Time) EligibilityResult {
	reasons := []string{}
	warnings := []string{}
	status := EligibilityHighConfidence

	if currentlyHoldsSameCard(input.CardHistory, offer) {
		return EligibilityResult{
			Eligible: false,
			Status:   EligibilityIneligible,
			Reasons:  []string{"Your card history says you currently hold this card."},
		}
	}
	reasons = append(reasons, "You do not appear to currently hold this card.")

	if offer.Network == CardNetworkAmex && !input.AcceptsAmex {
		return EligibilityResult{
			Eligible: false,
			Status:   EligibilityIneligible,
			Reasons:  []string{"This is an American Express card and you said you do not want Amex cards."},
		}
	}

	if input.AnnualFeePreference == AnnualFeePreferenceStrictMax && input.MaxAnnualFeeCents > 0 && offer.AnnualFeeCents > input.MaxAnnualFeeCents {
		return EligibilityResult{
			Eligible: false,
			Status:   EligibilityIneligible,
			Reasons:  []string{"The annual fee is above your strict maximum."},
		}
	}

	for _, rule := range offer.EligibilityRules {
		ruleStatus, ruleWarnings := evaluateEligibilityRule(input, offer, rule, now)
		warnings = append(warnings, ruleWarnings...)
		status = lowerEligibilityStatus(status, ruleStatus)
	}

	if status == EligibilityHighConfidence {
		reasons = append(reasons, "No recent matching card history was found.")
	}

	return EligibilityResult{
		Eligible: status != EligibilityIneligible,
		Status:   status,
		Reasons:  reasons,
		Warnings: warnings,
	}
}

func evaluateEligibilityRule(input RecommendationInput, offer CardOffer, rule EligibilityRule, now time.Time) (EligibilityStatus, []string) {
	switch rule.Type {
	case "manual_review":
		return EligibilityManualReview, []string{rule.Description}
	case "new_amex_card_members_only":
		// Honour the rule's WindowDays so a card held 5+ years ago does not
		// permanently disqualify the user from "new card member" offers.
		// Australian Amex new-member rules are typically 18 months (540 days).
		if recentlyHeldIssuer(input.CardHistory, "American Express", rule.WindowDays, now) ||
			recentlyHeldIssuer(input.CardHistory, "Amex", rule.WindowDays, now) {
			return EligibilityLowConfidence, []string{cardHistoryWarning(rule.Description)}
		}
	case "new_cardholders_only":
		// Honour the rule's WindowDays. Without a window, any historical card
		// from this issuer (even closed a decade ago) would trigger the
		// warning, which produces noise on almost every recommendation for
		// long-term Australian banking customers.
		if recentlyHeldMatchingIssuer(input.CardHistory, offer, rule.WindowDays, now) {
			return EligibilityMediumConfidence, []string{cardHistoryWarning(rule.Description)}
		}
	case "not_current_cardholder":
		if currentlyHoldsSameIssuer(input.CardHistory, offer) {
			return EligibilityLowConfidence, []string{cardHistoryWarning(rule.Description)}
		}
		if recentlyHeldMatchingIssuer(input.CardHistory, offer, rule.WindowDays, now) {
			return EligibilityLowConfidence, []string{cardHistoryWarning(rule.Description)}
		}
	case "not_held_recently":
		if recentlyHeldMatchingIssuer(input.CardHistory, offer, rule.WindowDays, now) {
			return EligibilityLowConfidence, []string{cardHistoryWarning(rule.Description)}
		}
	default:
		return EligibilityManualReview, []string{rule.Description}
	}

	return EligibilityHighConfidence, nil
}

func currentlyHoldsSameCard(history []UserCardHistoryItem, offer CardOffer) bool {
	for _, item := range history {
		if item.CurrentlyHeld && sameIssuer(item.Issuer, offer.Issuer) && sameText(item.CardName, offer.CardName) {
			return true
		}
	}
	return false
}

func currentlyHoldsSameIssuer(history []UserCardHistoryItem, offer CardOffer) bool {
	for _, item := range history {
		if item.CurrentlyHeld && sameIssuer(item.Issuer, offer.Issuer) {
			return true
		}
	}
	return false
}

func recentlyHeldMatchingIssuer(history []UserCardHistoryItem, offer CardOffer, windowDays *int, now time.Time) bool {
	return recentlyHeldIssuer(history, offer.Issuer, windowDays, now)
}

// recentlyHeldIssuer reports whether the user currently holds or recently
// closed a card from the given issuer. A nil windowDays means the rule has no
// time bound, in which case only currently-held cards count — historical
// closed cards do not, because unbounded "ever held" rules produce too many
// false positives for long-term banking customers.
func recentlyHeldIssuer(history []UserCardHistoryItem, issuer string, windowDays *int, now time.Time) bool {
	var cutoff time.Time
	if windowDays != nil {
		cutoff = now.AddDate(0, 0, -*windowDays)
	}

	for _, item := range history {
		if !sameIssuer(item.Issuer, issuer) {
			continue
		}
		if item.CurrentlyHeld {
			return true
		}
		if windowDays != nil && item.ClosedAt != nil && item.ClosedAt.After(cutoff) {
			return true
		}
	}
	return false
}

func cardHistoryWarning(description string) string {
	if description == "" {
		return "Your card history may affect eligibility for this offer."
	}
	return "Your card history may affect eligibility: " + description
}

func lowerEligibilityStatus(current EligibilityStatus, next EligibilityStatus) EligibilityStatus {
	if eligibilityRank(next) > eligibilityRank(current) {
		return next
	}
	return current
}

func eligibilityRank(status EligibilityStatus) int {
	switch status {
	case EligibilityHighConfidence:
		return 0
	case EligibilityMediumConfidence:
		return 1
	case EligibilityManualReview:
		return 2
	case EligibilityLowConfidence:
		return 3
	case EligibilityIneligible:
		return 4
	default:
		return 2
	}
}

func sameText(left string, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func sameIssuer(left string, right string) bool {
	leftKey := normaliseIssuer(left)
	rightKey := normaliseIssuer(right)
	return leftKey != "" && leftKey == rightKey
}

func normaliseIssuer(value string) string {
	cleaned := strings.ToLower(strings.TrimSpace(value))
	cleaned = strings.NewReplacer(
		".", "",
		"&", "and",
		"-", " ",
		"_", " ",
	).Replace(cleaned)
	compact := strings.ReplaceAll(strings.Join(strings.Fields(cleaned), " "), " ", "")

	switch compact {
	case "amex", "americanexpress":
		return "americanexpress"
	case "anz", "australiaandnewzealandbank", "australiaandnewzealandbankinggroup":
		return "anz"
	case "nab", "nationalaustraliabank":
		return "nab"
	case "cba", "commbank", "commonwealthbank", "commonwealthbankofaustralia":
		return "commonwealthbank"
	case "westpac", "westpacbankingcorporation":
		return "westpac"
	case "stgeorge", "saintgeorge", "banksa", "bankofmelbourne":
		return "westpacregionalbrands"
	case "bankwest":
		return "bankwest"
	case "qantasmoney", "qantaspremier":
		return "qantasmoney"
	case "virginmoney":
		return "virginmoney"
	case "hsbc":
		return "hsbc"
	case "suncorp", "suncorpbank":
		return "suncorpbank"
	case "bendigo", "bendigobank":
		return "bendigobank"
	case "citi", "citibank":
		return "citi"
	case "macquarie", "macquariebank":
		return "macquariebank"
	default:
		return compact
	}
}
