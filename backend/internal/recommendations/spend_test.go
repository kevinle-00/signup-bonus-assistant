package recommendations

import "testing"

func TestAssessSpendRequirement(t *testing.T) {
	tests := []struct {
		name                             string
		monthlySpendCents                int
		expectedLargePurchases90DayCents int
		minimumSpendCents                int
		spendWindowDays                  int
		wantProjectedSpendCents          int
		wantDifficulty                   SpendDifficulty
		wantAchievable                   bool
	}{
		{
			name:                    "easy when projected spend is at least 150 percent of minimum spend",
			monthlySpendCents:       250000,
			minimumSpendCents:       500000,
			spendWindowDays:         90,
			wantProjectedSpendCents: 750000,
			wantDifficulty:          SpendDifficultyEasy,
			wantAchievable:          true,
		},
		{
			name:                    "achievable when projected spend has a small buffer above minimum spend",
			monthlySpendCents:       110000,
			minimumSpendCents:       300000,
			spendWindowDays:         90,
			wantProjectedSpendCents: 330000,
			wantDifficulty:          SpendDifficultyAchievable,
			wantAchievable:          true,
		},
		{
			name:                    "tight when projected spend is at least 85 percent of minimum spend",
			monthlySpendCents:       85000,
			minimumSpendCents:       300000,
			spendWindowDays:         90,
			wantProjectedSpendCents: 255000,
			wantDifficulty:          SpendDifficultyTight,
			wantAchievable:          true,
		},
		{
			name:                    "unlikely when projected spend is below 85 percent of minimum spend",
			monthlySpendCents:       70000,
			minimumSpendCents:       300000,
			spendWindowDays:         90,
			wantProjectedSpendCents: 210000,
			wantDifficulty:          SpendDifficultyUnlikely,
			wantAchievable:          false,
		},
		{
			name:                             "large purchases increase projected spend",
			monthlySpendCents:                200000,
			expectedLargePurchases90DayCents: 100000,
			minimumSpendCents:                500000,
			spendWindowDays:                  90,
			wantProjectedSpendCents:          700000,
			wantDifficulty:                   SpendDifficultyAchievable,
			wantAchievable:                   true,
		},
		{
			// 90-day large purchases should not be credited in full to a
			// 30-day spend window. $9,000 over 90 days = $3,000 prorated to
			// 30 days, on top of $3,000 monthly spend = $6,000 projected
			// against a $5,000 minimum (achievable, not easy).
			name:                             "large purchases are prorated to a shorter spend window",
			monthlySpendCents:                300000,
			expectedLargePurchases90DayCents: 900000,
			minimumSpendCents:                500000,
			spendWindowDays:                  30,
			wantProjectedSpendCents:          600000,
			wantDifficulty:                   SpendDifficultyAchievable,
			wantAchievable:                   true,
		},
		{
			name:                    "non-standard spend windows round up to full months",
			monthlySpendCents:       100000,
			minimumSpendCents:       350000,
			spendWindowDays:         91,
			wantProjectedSpendCents: 400000,
			wantDifficulty:          SpendDifficultyAchievable,
			wantAchievable:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AssessSpendRequirement(
				RecommendationInput{
					MonthlySpendCents:                     tt.monthlySpendCents,
					ExpectedLargePurchasesNext90DaysCents: tt.expectedLargePurchases90DayCents,
				},
				CardOffer{
					MinimumSpendCents: tt.minimumSpendCents,
					SpendWindowDays:   tt.spendWindowDays,
				},
			)

			if got.ProjectedUserSpendCents != tt.wantProjectedSpendCents {
				t.Fatalf("ProjectedUserSpendCents = %d, want %d", got.ProjectedUserSpendCents, tt.wantProjectedSpendCents)
			}
			if got.Difficulty != tt.wantDifficulty {
				t.Fatalf("Difficulty = %q, want %q", got.Difficulty, tt.wantDifficulty)
			}
			if got.Achievable != tt.wantAchievable {
				t.Fatalf("Achievable = %t, want %t", got.Achievable, tt.wantAchievable)
			}
			if got.Reason == "" {
				t.Fatal("Reason is empty")
			}
		})
	}
}

func TestAssessSpendRequirementNoMinimumSpend(t *testing.T) {
	got := AssessSpendRequirement(
		RecommendationInput{MonthlySpendCents: 0},
		CardOffer{MinimumSpendCents: 0, SpendWindowDays: 90},
	)

	if got.Difficulty != SpendDifficultyEasy {
		t.Fatalf("Difficulty = %q, want %q", got.Difficulty, SpendDifficultyEasy)
	}
	if !got.Achievable {
		t.Fatal("Achievable = false, want true")
	}
}

func TestCeilDiv(t *testing.T) {
	tests := []struct {
		value   int
		divisor int
		want    int
	}{
		{value: 90, divisor: 30, want: 3},
		{value: 91, divisor: 30, want: 4},
		{value: 1, divisor: 30, want: 1},
		{value: 0, divisor: 30, want: 0},
		{value: 90, divisor: 0, want: 0},
	}

	for _, tt := range tests {
		got := ceilDiv(tt.value, tt.divisor)
		if got != tt.want {
			t.Fatalf("ceilDiv(%d, %d) = %d, want %d", tt.value, tt.divisor, got, tt.want)
		}
	}
}
