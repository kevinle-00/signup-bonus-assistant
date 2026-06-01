export type OptimisationGoal =
  | 'max_net_value'
  | 'qantas_points'
  | 'velocity_points'
  | 'cashback'
  | 'low_effort'

export type AnnualFeePreference = 'strict_max' | 'prefer_low' | 'flexible'

export type RecommendationInput = {
  optimisationGoal: OptimisationGoal
  monthlySpendCents: number
  expectedLargePurchasesNext90DaysCents: number
  annualFeePreference: AnnualFeePreference
  maxAnnualFeeCents: number
  acceptsAmex: boolean
}

export type CardOffer = {
  issuer: string
  cardName: string
  rewardProgram: string
  rewardType: string
  network: string
  signupBonusPoints: number
  cashbackCents: number
  minimumSpendCents: number
  spendWindowDays: number
  annualFeeCents: number
  laterBonusCondition?: string
  termsSummary: string[]
}

export type ValueBreakdown = {
  signupBonusValueCents: number
  travelCreditValueCents: number
  annualFeeCents: number
  netEstimatedValueCents: number
}

export type SpendRequirement = {
  minimumSpendCents: number
  spendWindowDays: number
  projectedUserSpendCents: number
  achievable: boolean
  difficulty: string
  reason: string
}

export type RecommendationCandidate = {
  offer: CardOffer
  rank: number
  score: number
  valueBreakdown: ValueBreakdown
  spendRequirement: SpendRequirement
  reasons: string[] | null
  warnings: string[] | null
}

export type ActionChecklistItem = {
  kind: string
  title: string
  description: string
  dueAt?: string
}

export type RecommendationRoadmap = {
  hasRecommendation: boolean
  summary: {
    estimatedYearOneValueCents: number
    cardsConsidered: number
    eligibleCards: number
    highestNetValueCents: number
  }
  bestRecommendation?: RecommendationCandidate
  alternatives: RecommendationCandidate[] | null
  ineligibleOrCautionCards: RecommendationCandidate[] | null
  actionChecklist: ActionChecklistItem[] | null
  reasons: string[] | null
  warnings: string[] | null
  noRecommendationReasons: string[] | null
}
