package recommendations

import "time"

type RewardType string

const (
	RewardTypeQantasPoints   RewardType = "qantas_points"
	RewardTypeVelocityPoints RewardType = "velocity_points"
	RewardTypeCashback       RewardType = "cashback"
	RewardTypeFlexiblePoints RewardType = "flexible_points"
	RewardTypeBankPoints     RewardType = "bank_points"
	RewardTypeTravelPerks    RewardType = "travel_perks"
)

type CardNetwork string

const (
	CardNetworkVisa       CardNetwork = "visa"
	CardNetworkMastercard CardNetwork = "mastercard"
	CardNetworkAmex       CardNetwork = "amex"
)

type AnnualFeePreference string

const (
	AnnualFeePreferenceStrictMax AnnualFeePreference = "strict_max"
	AnnualFeePreferencePreferLow AnnualFeePreference = "prefer_low"
	AnnualFeePreferenceFlexible  AnnualFeePreference = "flexible"
)

type OptimisationGoal string

const (
	OptimisationGoalMaxNetValue OptimisationGoal = "max_net_value"
	OptimisationGoalQantas      OptimisationGoal = "qantas_points"
	OptimisationGoalVelocity    OptimisationGoal = "velocity_points"
	OptimisationGoalCashback    OptimisationGoal = "cashback"
	OptimisationGoalLowEffort   OptimisationGoal = "low_effort"
)

type SpendingCategory string

const (
	SpendingCategoryGroceries      SpendingCategory = "groceries"
	SpendingCategoryDining         SpendingCategory = "dining"
	SpendingCategoryTravel         SpendingCategory = "travel"
	SpendingCategoryBills          SpendingCategory = "bills"
	SpendingCategoryOnlineShopping SpendingCategory = "online_shopping"
	SpendingCategoryFuel           SpendingCategory = "fuel"
	SpendingCategoryOther          SpendingCategory = "other"
)

type CardOffer struct {
	ID                           string
	Issuer                       string
	CardName                     string
	RewardProgram                string
	RewardType                   RewardType
	Network                      CardNetwork
	SignupBonusPoints            int
	EstimatedBonusValueCents     int
	CashbackCents                int
	MinimumSpendCents            int
	SpendWindowDays              int
	AnnualFeeCents               int
	TravelCreditCents            int
	LaterBonusPoints             int
	LaterBonusCondition          *string
	LaterBonusIncludedInMVPValue bool
	OfferExpiresAt               *time.Time
	EligibilityRules             []EligibilityRule
	TermsSummary                 []string
}

type EligibilityRule struct {
	Type        string
	Description string
	WindowDays  *int
}

type RecommendationInput struct {
	OptimisationGoal                      OptimisationGoal
	MonthlySpendCents                     int
	ExpectedLargePurchasesNext90DaysCents int
	SpendingCategories                    []SpendingCategory
	AnnualFeePreference                   AnnualFeePreference
	MaxAnnualFeeCents                     int
	AcceptsAmex                           bool
	CardHistory                           []UserCardHistoryItem
}

type UserCardHistoryItem struct {
	Issuer        string
	CardName      string
	OpenedAt      *time.Time
	ClosedAt      *time.Time
	CurrentlyHeld bool
}

type ValueBreakdown struct {
	SignupBonusValueCents         int
	RequiredSpendPointsValueCents int
	TravelCreditValueCents        int
	AnnualFeeCents                int
	NetEstimatedValueCents        int
}

type SpendDifficulty string

const (
	SpendDifficultyEasy       SpendDifficulty = "easy"
	SpendDifficultyAchievable SpendDifficulty = "achievable"
	SpendDifficultyTight      SpendDifficulty = "tight"
	SpendDifficultyUnlikely   SpendDifficulty = "unlikely"
)

type SpendRequirementResult struct {
	MinimumSpendCents       int
	SpendWindowDays         int
	ProjectedUserSpendCents int
	Achievable              bool
	Difficulty              SpendDifficulty
	Reason                  string
}

type EligibilityStatus string

const (
	EligibilityHighConfidence   EligibilityStatus = "high_confidence"
	EligibilityMediumConfidence EligibilityStatus = "medium_confidence"
	EligibilityLowConfidence    EligibilityStatus = "low_confidence"
	EligibilityIneligible       EligibilityStatus = "ineligible"
	EligibilityManualReview     EligibilityStatus = "manual_review"
)

type EligibilityResult struct {
	Eligible bool
	Status   EligibilityStatus
	Reasons  []string
	Warnings []string
}

type ScoreResult struct {
	Score    int
	Reasons  []string
	Warnings []string
}

type RecommendationCandidate struct {
	Offer            CardOffer
	Rank             int
	Score            int
	Eligibility      EligibilityResult
	ValueBreakdown   ValueBreakdown
	SpendRequirement SpendRequirementResult
	Reasons          []string
	Warnings         []string
}

type RecommendationSummary struct {
	EstimatedYearOneValueCents int
	CardsConsidered            int
	EligibleCards              int
	HighestNetValueCents       int
}

type RecommendationResult struct {
	Summary                  RecommendationSummary
	BestRecommendation       *RecommendationCandidate
	Alternatives             []RecommendationCandidate
	IneligibleOrCautionCards []RecommendationCandidate
}

type ActionChecklistItemKind string

const (
	ActionChecklistReviewTerms ActionChecklistItemKind = "review_terms"
	ActionChecklistReviewRisk  ActionChecklistItemKind = "review_risk"
	ActionChecklistApply       ActionChecklistItemKind = "apply"
	ActionChecklistMeetSpend   ActionChecklistItemKind = "meet_spend"
	ActionChecklistTrackBonus  ActionChecklistItemKind = "track_bonus"
	ActionChecklistAnnualFee   ActionChecklistItemKind = "annual_fee_review"
	ActionChecklistLaterBonus  ActionChecklistItemKind = "later_bonus_review"
)

type ActionChecklistItem struct {
	Kind        ActionChecklistItemKind
	Title       string
	Description string
	DueAt       *time.Time
}

type RecommendationRoadmap struct {
	HasRecommendation        bool
	Summary                  RecommendationSummary
	BestRecommendation       *RecommendationCandidate
	Alternatives             []RecommendationCandidate
	IneligibleOrCautionCards []RecommendationCandidate
	ActionChecklist          []ActionChecklistItem
	Reasons                  []string
	Warnings                 []string
	NoRecommendationReasons  []string
}
