package recommendations

import (
	"strings"
	"testing"
	"time"
)

func TestBuildActionChecklistIncludesSpendDeadline(t *testing.T) {
	now := testNow()
	candidate := RecommendationCandidate{
		Offer: testOffer("Best", RewardTypeQantasPoints, 100000, 500000, 99_00),
	}

	items := BuildActionChecklist(candidate, now)
	spendItem := findChecklistItem(t, items, ActionChecklistMeetSpend)

	if !strings.Contains(spendItem.Title, "$5000.00") {
		t.Fatalf("spend title = %q, want minimum spend amount", spendItem.Title)
	}
	if spendItem.DueAt == nil || !spendItem.DueAt.Equal(now.AddDate(0, 0, 90)) {
		t.Fatalf("spend due date = %v, want %v", spendItem.DueAt, now.AddDate(0, 0, 90))
	}
}

func TestBuildActionChecklistIncludesRiskReviewForWarnings(t *testing.T) {
	candidate := RecommendationCandidate{
		Offer:       testOffer("Best", RewardTypeQantasPoints, 100000, 300000, 99_00),
		Eligibility: EligibilityResult{Status: EligibilityManualReview},
		Warnings:    []string{"Check the personalised offer before applying."},
	}

	items := BuildActionChecklist(candidate, testNow())
	riskItem := findChecklistItem(t, items, ActionChecklistReviewRisk)

	if !strings.Contains(riskItem.Description, "personalised offer") {
		t.Fatalf("risk description = %q, want warning text", riskItem.Description)
	}
}

func TestBuildActionChecklistIncludesLaterBonusAsReviewOnly(t *testing.T) {
	condition := "Keep card open for over 12 months."
	candidate := RecommendationCandidate{
		Offer: CardOffer{
			CardName:            "Best",
			RewardType:          RewardTypeQantasPoints,
			MinimumSpendCents:   300000,
			SpendWindowDays:     90,
			AnnualFeeCents:      99_00,
			LaterBonusCondition: &condition,
		},
	}

	items := BuildActionChecklist(candidate, testNow())
	laterBonusItem := findChecklistItem(t, items, ActionChecklistLaterBonus)

	if !strings.Contains(laterBonusItem.Description, "not included in the MVP year-one value") {
		t.Fatalf("later bonus description = %q, want MVP exclusion note", laterBonusItem.Description)
	}
}

func TestBuildActionChecklistIncludesAnnualFeeReview(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	candidate := RecommendationCandidate{Offer: testOffer("Best", RewardTypeQantasPoints, 100000, 300000, 420_00)}

	items := BuildActionChecklist(candidate, now)
	annualFeeItem := findChecklistItem(t, items, ActionChecklistAnnualFee)

	if !strings.Contains(annualFeeItem.Description, "$420.00") {
		t.Fatalf("annual fee description = %q, want fee amount", annualFeeItem.Description)
	}
	if annualFeeItem.DueAt == nil || !annualFeeItem.DueAt.Equal(now.AddDate(0, 11, 0)) {
		t.Fatalf("annual fee due date = %v, want %v", annualFeeItem.DueAt, now.AddDate(0, 11, 0))
	}
}

func findChecklistItem(t *testing.T, items []ActionChecklistItem, kind ActionChecklistItemKind) ActionChecklistItem {
	t.Helper()
	for _, item := range items {
		if item.Kind == kind {
			return item
		}
	}
	t.Fatalf("checklist item %q not found in %#v", kind, items)
	return ActionChecklistItem{}
}
