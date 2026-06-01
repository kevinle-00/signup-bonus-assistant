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
	ID                           string            `json:"id"`
	Issuer                       string            `json:"issuer"`
	CardName                     string            `json:"cardName"`
	RewardProgram                string            `json:"rewardProgram"`
	RewardType                   RewardType        `json:"rewardType"`
	Network                      CardNetwork       `json:"network"`
	SignupBonusPoints            int               `json:"signupBonusPoints"`
	EstimatedBonusValueCents     int               `json:"estimatedBonusValueCents"`
	CashbackCents                int               `json:"cashbackCents"`
	MinimumSpendCents            int               `json:"minimumSpendCents"`
	SpendWindowDays              int               `json:"spendWindowDays"`
	AnnualFeeCents               int               `json:"annualFeeCents"`
	TravelCreditCents            int               `json:"travelCreditCents"`
	LaterBonusPoints             int               `json:"laterBonusPoints"`
	LaterBonusCondition          *string           `json:"laterBonusCondition,omitempty"`
	LaterBonusIncludedInMVPValue bool              `json:"laterBonusIncludedInMVPValue"`
	OfferExpiresAt               *time.Time        `json:"offerExpiresAt,omitempty"`
	EligibilityRules             []EligibilityRule `json:"eligibilityRules"`
	TermsSummary                 []string          `json:"termsSummary"`
}

type EligibilityRule struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	WindowDays  *int   `json:"windowDays,omitempty"`
}

type RecommendationInput struct {
	OptimisationGoal                      OptimisationGoal      `json:"optimisationGoal"`
	MonthlySpendCents                     int                   `json:"monthlySpendCents"`
	ExpectedLargePurchasesNext90DaysCents int                   `json:"expectedLargePurchasesNext90DaysCents"`
	SpendingCategories                    []SpendingCategory    `json:"spendingCategories"`
	AnnualFeePreference                   AnnualFeePreference   `json:"annualFeePreference"`
	MaxAnnualFeeCents                     int                   `json:"maxAnnualFeeCents"`
	AcceptsAmex                           bool                  `json:"acceptsAmex"`
	CardHistory                           []UserCardHistoryItem `json:"cardHistory"`
}

type UserCardHistoryItem struct {
	Issuer        string     `json:"issuer"`
	CardName      string     `json:"cardName"`
	OpenedAt      *time.Time `json:"openedAt,omitempty"`
	ClosedAt      *time.Time `json:"closedAt,omitempty"`
	CurrentlyHeld bool       `json:"currentlyHeld"`
}

type ValueBreakdown struct {
	SignupBonusValueCents         int `json:"signupBonusValueCents"`
	RequiredSpendPointsValueCents int `json:"requiredSpendPointsValueCents"`
	TravelCreditValueCents        int `json:"travelCreditValueCents"`
	AnnualFeeCents                int `json:"annualFeeCents"`
	NetEstimatedValueCents        int `json:"netEstimatedValueCents"`
}

type SpendDifficulty string

const (
	SpendDifficultyEasy       SpendDifficulty = "easy"
	SpendDifficultyAchievable SpendDifficulty = "achievable"
	SpendDifficultyTight      SpendDifficulty = "tight"
	SpendDifficultyUnlikely   SpendDifficulty = "unlikely"
)

type SpendRequirementResult struct {
	MinimumSpendCents       int             `json:"minimumSpendCents"`
	SpendWindowDays         int             `json:"spendWindowDays"`
	ProjectedUserSpendCents int             `json:"projectedUserSpendCents"`
	Achievable              bool            `json:"achievable"`
	Difficulty              SpendDifficulty `json:"difficulty"`
	Reason                  string          `json:"reason"`
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
	Eligible bool              `json:"eligible"`
	Status   EligibilityStatus `json:"status"`
	Reasons  []string          `json:"reasons"`
	Warnings []string          `json:"warnings"`
}

type ScoreResult struct {
	Score    int      `json:"score"`
	Reasons  []string `json:"reasons"`
	Warnings []string `json:"warnings"`
}

type RecommendationCandidate struct {
	Offer            CardOffer              `json:"offer"`
	Rank             int                    `json:"rank"`
	Score            int                    `json:"score"`
	Eligibility      EligibilityResult      `json:"eligibility"`
	ValueBreakdown   ValueBreakdown         `json:"valueBreakdown"`
	SpendRequirement SpendRequirementResult `json:"spendRequirement"`
	Reasons          []string               `json:"reasons"`
	Warnings         []string               `json:"warnings"`
}

type RecommendationSummary struct {
	EstimatedYearOneValueCents int `json:"estimatedYearOneValueCents"`
	CardsConsidered            int `json:"cardsConsidered"`
	EligibleCards              int `json:"eligibleCards"`
	HighestNetValueCents       int `json:"highestNetValueCents"`
}

type RecommendationResult struct {
	Summary                  RecommendationSummary     `json:"summary"`
	BestRecommendation       *RecommendationCandidate  `json:"bestRecommendation,omitempty"`
	Alternatives             []RecommendationCandidate `json:"alternatives"`
	IneligibleOrCautionCards []RecommendationCandidate `json:"ineligibleOrCautionCards"`
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
	Kind        ActionChecklistItemKind `json:"kind"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	DueAt       *time.Time              `json:"dueAt,omitempty"`
}

type RecommendationRoadmap struct {
	HasRecommendation        bool                      `json:"hasRecommendation"`
	Summary                  RecommendationSummary     `json:"summary"`
	BestRecommendation       *RecommendationCandidate  `json:"bestRecommendation,omitempty"`
	Alternatives             []RecommendationCandidate `json:"alternatives"`
	IneligibleOrCautionCards []RecommendationCandidate `json:"ineligibleOrCautionCards"`
	ActionChecklist          []ActionChecklistItem     `json:"actionChecklist"`
	Reasons                  []string                  `json:"reasons"`
	Warnings                 []string                  `json:"warnings"`
	NoRecommendationReasons  []string                  `json:"noRecommendationReasons"`
}
