package recommendations

import "testing"

func TestEstimateBonusValueCents(t *testing.T) {
	tests := []struct {
		name          string
		rewardType    RewardType
		points        int
		cashbackCents int
		want          int
	}{
		{
			name:       "qantas points use one cent per point",
			rewardType: RewardTypeQantasPoints,
			points:     100000,
			want:       100000,
		},
		{
			name:       "velocity points use one cent per point",
			rewardType: RewardTypeVelocityPoints,
			points:     70000,
			want:       70000,
		},
		{
			name:          "bank points use forty hundredths of a cent and include cashback",
			rewardType:    RewardTypeBankPoints,
			points:        130000,
			cashbackCents: 10000,
			want:          62000,
		},
		{
			name:       "flexible points use eighty hundredths of a cent",
			rewardType: RewardTypeFlexiblePoints,
			points:     200000,
			want:       160000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EstimateBonusValueCents(tt.rewardType, tt.points, tt.cashbackCents)
			if err != nil {
				t.Fatalf("EstimateBonusValueCents() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("EstimateBonusValueCents() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEstimateBonusValueCentsUnknownRewardType(t *testing.T) {
	_, err := EstimateBonusValueCents(RewardType("unknown"), 100000, 0)
	if err == nil {
		t.Fatal("EstimateBonusValueCents() error = nil, want error")
	}
}

func TestCalculateValue(t *testing.T) {
	offer := CardOffer{
		RewardType:        RewardTypeFlexiblePoints,
		SignupBonusPoints: 200000,
		AnnualFeeCents:    145000,
		TravelCreditCents: 45000,
	}

	got, err := CalculateValue(offer)
	if err != nil {
		t.Fatalf("CalculateValue() error = %v", err)
	}

	want := ValueBreakdown{
		SignupBonusValueCents:         160000,
		RequiredSpendPointsValueCents: 0,
		TravelCreditValueCents:        45000,
		AnnualFeeCents:                145000,
		NetEstimatedValueCents:        60000,
	}

	if got != want {
		t.Fatalf("CalculateValue() = %+v, want %+v", got, want)
	}
}
