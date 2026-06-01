package recommendations

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
	EligibilityRules             []EligibilityRule
	TermsSummary                 []string
}

type EligibilityRule struct {
	Type        string
	Description string
	WindowDays  *int
}

type ValueBreakdown struct {
	SignupBonusValueCents         int
	RequiredSpendPointsValueCents int
	TravelCreditValueCents        int
	AnnualFeeCents                int
	NetEstimatedValueCents        int
}
