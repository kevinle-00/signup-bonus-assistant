package recommendations

import "fmt"

const pointValueScale = 100

var pointValueCentsScaled = map[RewardType]int{
	RewardTypeQantasPoints:   100,
	RewardTypeVelocityPoints: 100,
	RewardTypeBankPoints:     40,
	RewardTypeFlexiblePoints: 80,
	RewardTypeCashback:       0,
	RewardTypeTravelPerks:    0,
}

func EstimateBonusValueCents(rewardType RewardType, points int, cashbackCents int) (int, error) {
	pointValue, ok := pointValueCentsScaled[rewardType]
	if !ok {
		return 0, fmt.Errorf("unknown reward type %q", rewardType)
	}

	return (points*pointValue)/pointValueScale + cashbackCents, nil
}

func CalculateValue(offer CardOffer) (ValueBreakdown, error) {
	signupBonusValueCents, err := EstimateBonusValueCents(
		offer.RewardType,
		offer.SignupBonusPoints,
		offer.CashbackCents,
	)
	if err != nil {
		return ValueBreakdown{}, err
	}

	// MVP deliberately excludes points earned from the required spend. Earn rates vary
	// by card, category, caps, exclusions, and government payments, so modelling them
	// without richer spend data would create false precision.
	requiredSpendPointsValueCents := 0

	netEstimatedValueCents := signupBonusValueCents + requiredSpendPointsValueCents + offer.TravelCreditCents - offer.AnnualFeeCents

	return ValueBreakdown{
		SignupBonusValueCents:         signupBonusValueCents,
		RequiredSpendPointsValueCents: requiredSpendPointsValueCents,
		TravelCreditValueCents:        offer.TravelCreditCents,
		AnnualFeeCents:                offer.AnnualFeeCents,
		NetEstimatedValueCents:        netEstimatedValueCents,
	}, nil
}
