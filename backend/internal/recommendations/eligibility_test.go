package recommendations

import (
	"testing"
	"time"
)

func TestEvaluateEligibilityCurrentlyHeldSameCard(t *testing.T) {
	got := EvaluateEligibility(
		RecommendationInput{
			AcceptsAmex: true,
			CardHistory: []UserCardHistoryItem{
				{Issuer: "NAB", CardName: "NAB Qantas Rewards Signature Card", CurrentlyHeld: true},
			},
		},
		CardOffer{Issuer: "NAB", CardName: "NAB Qantas Rewards Signature Card"},
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	)

	if got.Eligible {
		t.Fatal("Eligible = true, want false")
	}
	if got.Status != EligibilityIneligible {
		t.Fatalf("Status = %q, want %q", got.Status, EligibilityIneligible)
	}
}

func TestEvaluateEligibilityRejectsAmexWhenUserDoesNotAcceptAmex(t *testing.T) {
	got := EvaluateEligibility(
		RecommendationInput{AcceptsAmex: false},
		CardOffer{Issuer: "American Express", CardName: "The American Express Platinum Card", Network: CardNetworkAmex},
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	)

	if got.Eligible {
		t.Fatal("Eligible = true, want false")
	}
	if got.Status != EligibilityIneligible {
		t.Fatalf("Status = %q, want %q", got.Status, EligibilityIneligible)
	}
}

func TestEvaluateEligibilityRejectsStrictAnnualFeeMaximum(t *testing.T) {
	got := EvaluateEligibility(
		RecommendationInput{
			AcceptsAmex:         true,
			AnnualFeePreference: AnnualFeePreferenceStrictMax,
			MaxAnnualFeeCents:   20000,
		},
		CardOffer{AnnualFeeCents: 42000},
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	)

	if got.Eligible {
		t.Fatal("Eligible = true, want false")
	}
	if got.Status != EligibilityIneligible {
		t.Fatalf("Status = %q, want %q", got.Status, EligibilityIneligible)
	}
}

func TestEvaluateEligibilityDowngradesRecentIssuerHistory(t *testing.T) {
	windowDays := 730
	closedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	got := EvaluateEligibility(
		RecommendationInput{
			AcceptsAmex: true,
			CardHistory: []UserCardHistoryItem{
				{Issuer: "ANZ", CardName: "Old ANZ Rewards Card", ClosedAt: &closedAt},
			},
		},
		CardOffer{
			Issuer: "ANZ",
			EligibilityRules: []EligibilityRule{
				{Type: "not_held_recently", Description: "Not available if held recently.", WindowDays: &windowDays},
			},
		},
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	)

	if !got.Eligible {
		t.Fatal("Eligible = false, want true with low confidence")
	}
	if got.Status != EligibilityLowConfidence {
		t.Fatalf("Status = %q, want %q", got.Status, EligibilityLowConfidence)
	}
	if len(got.Warnings) == 0 {
		t.Fatal("Warnings is empty, want recent-cardholder warning")
	}
}

func TestEvaluateEligibilityNewCardholdersOnlyHonoursWindow(t *testing.T) {
	windowDays := 365
	recentClose := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldClose := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	offer := CardOffer{
		Issuer: "ANZ",
		EligibilityRules: []EligibilityRule{
			{Type: "new_cardholders_only", Description: "New cardholders only.", WindowDays: &windowDays},
		},
	}

	// Recent ANZ history within the 12-month window → downgrade to medium.
	recent := EvaluateEligibility(
		RecommendationInput{
			AcceptsAmex: true,
			CardHistory: []UserCardHistoryItem{{Issuer: "ANZ", CardName: "ANZ Old Card", ClosedAt: &recentClose}},
		},
		offer,
		now,
	)
	if recent.Status != EligibilityMediumConfidence {
		t.Fatalf("Recent issuer history Status = %q, want %q", recent.Status, EligibilityMediumConfidence)
	}

	// 8-year-old ANZ history is well outside the window → no downgrade.
	old := EvaluateEligibility(
		RecommendationInput{
			AcceptsAmex: true,
			CardHistory: []UserCardHistoryItem{{Issuer: "ANZ", CardName: "ANZ Old Card", ClosedAt: &oldClose}},
		},
		offer,
		now,
	)
	if old.Status != EligibilityHighConfidence {
		t.Fatalf("Old issuer history Status = %q, want %q (rule should ignore history outside its window)", old.Status, EligibilityHighConfidence)
	}
}

func TestEvaluateEligibilityNewAmexCardMembersOnlyHonoursWindow(t *testing.T) {
	windowDays := 540
	oldClose := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	offer := CardOffer{
		Issuer:  "American Express",
		Network: CardNetworkAmex,
		EligibilityRules: []EligibilityRule{
			{Type: "new_amex_card_members_only", Description: "New Amex Card Members only.", WindowDays: &windowDays},
		},
	}

	got := EvaluateEligibility(
		RecommendationInput{
			AcceptsAmex: true,
			CardHistory: []UserCardHistoryItem{{Issuer: "American Express", CardName: "Old Amex", ClosedAt: &oldClose}},
		},
		offer,
		now,
	)

	if got.Status != EligibilityHighConfidence {
		t.Fatalf("Status = %q, want %q (8-year-old Amex card should be outside the 18-month new-member window)", got.Status, EligibilityHighConfidence)
	}
}

func TestEvaluateEligibilityManualReview(t *testing.T) {
	got := EvaluateEligibility(
		RecommendationInput{AcceptsAmex: true},
		CardOffer{
			Issuer: "HSBC",
			EligibilityRules: []EligibilityRule{
				{Type: "manual_review", Description: "Partner eligibility may vary."},
			},
		},
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	)

	if !got.Eligible {
		t.Fatal("Eligible = false, want true")
	}
	if got.Status != EligibilityManualReview {
		t.Fatalf("Status = %q, want %q", got.Status, EligibilityManualReview)
	}
}
