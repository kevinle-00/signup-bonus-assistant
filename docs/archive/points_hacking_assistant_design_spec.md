# Points Hacking Assistant — Detailed Design Specification

## 1. Project Context

This project implements **Category 1A — Points Hacking Assistant** from the New Product Feature take-home assessment.

The product helps users identify the best credit card sign-up bonus to target next based on their spending profile, card history, preferences, eligibility constraints, and available Australian credit card offers.

The assessment should demonstrate:

- Product thinking
- Design craft
- Engineering quality
- Production-minded implementation
- End-to-end ownership
- Clear trade-off decisions

The goal is **not** to build a full credit card comparison site or a live financial data ingestion platform. The goal is to build a polished, realistic, production-oriented product focused on the core user value: helping users decide which credit card sign-up bonus to pursue next and why.

---

## 2. Product Summary

### Product name

**Points Hacking Assistant**

### One-line description

A decision-support tool that recommends the best credit card sign-up bonus for a user to target next, based on their spending, eligibility, card history, and rewards goals.

### Core user problem

Many users keep the same credit card for years and miss out on valuable sign-up bonuses. Even when they know points-churning exists, it is difficult to compare offers, understand eligibility rules, know whether they can meet minimum spend requirements, and plan which card to apply for next.

### Product promise

> Answer a few questions and get a clear, personalised card-switching recommendation with estimated value, eligibility confidence, spend requirements, and a simple action plan.

---

## 3. Scope

### In scope

Build a working web application with:

1. A guided onboarding form that collects user inputs relevant to recommendation quality.
2. A curated dataset of Australian credit card offers stored in PostgreSQL.
3. A Go backend API that evaluates eligibility, calculates expected value, scores offers, and returns recommendations.
4. A React frontend that presents the recommendation clearly and trustworthily.
5. A roadmap/action-plan view that shows what the user should do next.
6. Explanation logic showing why a card was recommended.
7. Basic tests around the recommendation logic.
8. Clear README explaining assumptions, trade-offs, setup, and future improvements.

### Out of scope

Do **not** build:

- Live bank scraping
- Real-time offer ingestion
- User authentication
- Bank account linking
- Transaction syncing
- Complex admin dashboard
- Payment integrations
- Full compliance-grade financial advice workflows
- Multi-user SaaS account management
- Microservices
- Background job workers unless absolutely necessary

### Important product trade-off

Use a **curated dataset** of Australian credit card offers instead of live offer ingestion.

Reasoning:

- Live offer ingestion is a meaningful production concern but not the core product experience.
- Manual curation allows the project to focus on recommendation quality, UX, data modelling, and explainability.
- The backend and database should still be designed so a future ingestion pipeline could update offers without changing the recommendation engine.

This trade-off should be explained in the README and walkthrough.

---

## 4. Recommended Tech Stack

### Frontend

- React
- TypeScript
- Vite
- Tailwind CSS
- Shadcn/ui or equivalent component library
- React Hook Form or controlled components
- Zod for frontend form validation if convenient

### Backend

- Go
- Chi router
- PostgreSQL
- pgx or database/sql
- Goose, golang-migrate, or simple SQL migration files

### Deployment target

Preferred:

- Frontend: Vercel
- Backend: Railway, Render, Fly.io, or similar
- Database: hosted PostgreSQL via Railway/Supabase/Neon/etc.

For local development:

- Docker Compose for PostgreSQL
- Backend runs on localhost
- Frontend runs on Vite dev server

---

## 5. High-Level Architecture

Keep the architecture intentionally simple:

```text
React Frontend
    ↓ HTTP/JSON
Go REST API
    ↓
Domain services
    ↓
PostgreSQL
```

The application complexity should live in the **domain layer**, not infrastructure.

### Backend layers

Use a familiar layered architecture:

```text
Handler → Service → Repository
```

Responsibilities:

- **Handlers** parse HTTP requests, validate input shape, call services, and return JSON.
- **Services** contain business logic such as recommendation scoring, eligibility checks, and value calculation.
- **Repositories** handle database access.

---

## 6. Suggested Repository Structure

Use a monorepo:

```text
points-hacking-assistant/
  README.md
  docker-compose.yml
  .env.example

  frontend/
    package.json
    src/
      main.tsx
      App.tsx
      api/
        client.ts
        recommendations.ts
        cardOffers.ts
      components/
        layout/
        onboarding/
        recommendation/
        ui/
      pages/
        HomePage.tsx
        ResultsPage.tsx
      types/
        api.ts
      utils/
        format.ts

  backend/
    go.mod
    go.sum
    cmd/
      api/
        main.go
    internal/
      config/
        config.go
      db/
        postgres.go
      httpapi/
        router.go
        middleware.go
      cardoffers/
        handler.go
        service.go
        repository.go
        types.go
      recommendations/
        handler.go
        service.go
        scorer.go
        eligibility.go
        value.go
        roadmap.go
        types.go
      shared/
        errors.go
        money.go
        response.go
    migrations/
      001_create_card_offers.sql
      002_create_recommendation_runs.sql
    seed/
      card_offers_seed.sql
    tests/
      recommendation_service_test.go
```

This structure is intentionally simple but production-minded.

---

## 7. Core User Flow

### Flow overview

```text
Landing page
  ↓
Guided onboarding form
  ↓
Submit recommendation request
  ↓
Results dashboard
  ↓
Best card recommendation
  ↓
Explanation + value breakdown
  ↓
Action checklist + future roadmap
```

### User journey

1. User lands on the app and sees the main value proposition.
2. User answers a small set of questions:
   - Monthly card spend
   - Expected large purchases in the next 3 months
   - Optimization goal
   - Current cards
   - Recently held cards
   - Annual fee preference and maximum annual fee
   - Whether they are open to Amex
3. Backend evaluates all available card offers.
4. App shows:
   - Estimated value of the recommended 12-month action plan
   - Best next card
   - Why it was chosen
   - Spend requirement achievability
   - Eligibility confidence
   - Alternatives
   - Ineligible or lower-confidence cards
   - Suggested action plan
5. User can inspect details and compare alternatives.

---

## 8. UX Design Principles

### 8.1 Make the value tangible immediately

The results page should lead with a clear dollar-value insight.

Example:

```text
You could earn an estimated $1,240 in net value over the next 12 months.
```

Avoid leading with a raw table of credit cards. Users should first understand the opportunity.

### 8.2 Recommend one clear next action

The product should not feel like a comparison table. It should feel like a decision assistant.

Primary recommendation card should say:

```text
Recommended next card
NAB Rewards Signature
Estimated net value: $720
Spend required: $3,000 in 90 days
Eligibility confidence: High
```

### 8.3 Explain the reasoning

Users need to trust why the system made the recommendation.

Every recommendation should include:

- Why this card scored highly
- Value breakdown
- Spend requirement achievability
- Eligibility notes
- Any warnings

Example explanation:

```text
We recommended this card because it has the highest net estimated value among cards you appear eligible for, the $3,000 spend requirement fits your projected 3-month spend, and it matches your preference for Qantas points.
```

### 8.4 Treat eligibility as confidence, not certainty

Credit card eligibility rules can be messy and issuer-specific. Do not pretend the app can guarantee approval or bonus eligibility.

Use statuses like:

- High confidence
- Medium confidence
- Low confidence
- Ineligible
- Manual review recommended

This is more trustworthy than pretending the recommendation is legally or financially definitive.

### 8.5 Keep onboarding lightweight

The onboarding form should not feel like a bank application.

Suggested copy:

```text
Answer a few questions and we’ll estimate which sign-up bonus is most worth targeting next.
```

Use 4–5 steps max.

---

## 9. Frontend Design Specification

### 9.1 Main pages

#### HomePage

Purpose:

- Explain the product simply.
- Start onboarding.

Sections:

1. Hero section
   - Headline: “Find your next best credit card bonus”
   - Subheadline: “Estimate which sign-up bonus is worth targeting based on your spend, card history, and rewards goals.”
   - CTA: “Find my next card”

2. How it works
   - Step 1: Tell us your spending profile
   - Step 2: Add cards you currently or recently held
   - Step 3: Get a recommendation and action plan

3. Trust/disclaimer block
   - “Estimates are based on curated offer data and simplified assumptions. Always check issuer terms before applying.”

#### Onboarding form

Can live on the HomePage or as a separate route.

Suggested steps:

##### Step 1 — Goal

Fields:

- optimisationGoal
  - max_net_value
  - qantas_points
  - velocity_points
  - cashback
  - low_effort

##### Step 2 — Spending profile

Fields:

- monthlySpendCents
- expectedLargePurchasesNext90DaysCents
- largePurchaseDescription optional
- spendingCategories optional
  - groceries
  - dining
  - travel
  - bills
  - online_shopping
  - fuel
  - other

Important calculation:

```text
projectedSpendInWindow = monthlySpendCents × ceil(spendWindowDays / 30) + expectedLargePurchasesNext90DaysCents
```

Since different offers may have different spend windows, calculate achievability per offer. If an offer has a non-standard spend window, surface that the estimate is approximate.

##### Step 3 — Card history

Fields:

- currentlyHeldCards: list of issuer/card names
- recentlyHeldCards: list with issuer/card name and closed date

For this version, allow simple manual entry rather than autocomplete.

Each history item:

- issuer
- cardName
- openedAt optional
- closedAt optional
- currentlyHeld boolean

##### Step 4 — Preferences

Fields:

- annualFeePreference
  - strict_max
  - prefer_low
  - flexible
- maxAnnualFeeCents
- acceptsAmex boolean
- minEstimatedValueCents optional

##### Step 5 — Review and submit

Show a summary of user inputs before submitting.

### 9.2 Results page

Sections:

#### Summary banner

Show:

- Estimated year-one value
- Number of offers considered
- Number of likely eligible offers
- Best recommendation

Example:

```text
Estimated 12-month action plan value: $1,240
Based on 20 available offers and your current spending profile.
```

#### Best recommendation card

Display:

- Card name
- Issuer
- Reward program
- Estimated net value
- Sign-up bonus points
- Minimum spend requirement
- Spend window
- Annual fee
- Travel credit if present
- Eligibility confidence
- Score

#### Why this card

Render backend-provided reasons.

Example bullets:

- Highest net estimated value among eligible cards
- Spend requirement appears achievable based on your projected spend
- Matches your preference for Qantas points
- Annual fee is within your selected limit

#### Value breakdown

Show a simple breakdown:

```text
Sign-up bonus value:      +$800
Required spend points:    +$45
Travel credit:            +$200
Annual fee:               -$395
--------------------------------
Estimated net value:       $650
```

#### Action checklist

Generate deterministic action items from the selected candidate. Do not use AI-generated prose for the checklist; it should be traceable to card-offer fields and recommendation results.

Example actions:

- Verify the current public offer terms.
- Review eligibility and spend cautions when warnings exist.
- Apply for the recommended card.
- Spend the required amount within the offer window.
- Track bonus posting.
- Review card before the next annual fee.
- Review later bonus conditions separately when present.

#### 12-month roadmap

Show a timeline:

```text
Month 0: Apply for recommended card
Month 1–3: Meet spend requirement
Month 4: Bonus expected after criteria met
Month 11: Review annual fee before renewal
Month 12: Consider next card opportunity
```

#### Alternatives

Show top 3 alternatives with:

- Card name
- Net value
- Reason it was not first
- Eligibility confidence

#### Ineligible / caution cards

Optional but valuable.

Show cards filtered out and why:

```text
ANZ Rewards Black
Not recommended because you currently hold this card.
```

or:

```text
Westpac Altitude Black
Medium confidence because recent-cardholder exclusion may apply.
```

---

## 10. Backend API Specification

### 10.1 Health check

```http
GET /health
```

Response:

```json
{
  "status": "ok"
}
```

### 10.2 List card offers

```http
GET /api/card-offers
```

Purpose:

- Useful for debugging and optional UI inspection.

Response:

```json
{
  "offers": [
    {
      "id": "uuid",
      "issuer": "NAB",
      "cardName": "NAB Rewards Signature",
      "rewardProgram": "nab_rewards",
      "network": "visa",
      "signupBonusPoints": 120000,
      "estimatedBonusValueCents": 72000,
      "minimumSpendCents": 300000,
      "spendWindowDays": 90,
      "annualFeeCents": 29500,
      "travelCreditCents": 0,
      "offerExpiresAt": "2026-07-31T00:00:00Z",
      "termsSummary": [
        "Bonus points awarded after meeting minimum spend requirement.",
        "Eligibility subject to issuer terms."
      ]
    }
  ]
}
```

### 10.3 Create recommendation

```http
POST /api/recommendations
```

Request:

```json
{
  "optimisationGoal": "qantas_points",
  "monthlySpendCents": 250000,
  "expectedLargePurchasesNext90DaysCents": 100000,
  "spendingCategories": ["groceries", "dining", "bills"],
  "annualFeePreference": "prefer_low",
  "maxAnnualFeeCents": 45000,
  "acceptsAmex": true,
  "cardHistory": [
    {
      "issuer": "ANZ",
      "cardName": "ANZ Rewards Black",
      "openedAt": "2024-01-01",
      "closedAt": "2024-12-01",
      "currentlyHeld": false
    }
  ]
}
```

Response:

```json
{
  "recommendationRunId": "uuid",
  "summary": {
    "estimatedYearOneValueCents": 47000,
    "cardsConsidered": 20,
    "eligibleCards": 8,
    "highestNetValueCents": 47000
  },
  "bestRecommendation": {
    "offer": {
      "id": "uuid",
      "issuer": "NAB",
      "cardName": "NAB Rewards Signature",
      "rewardProgram": "nab_rewards",
      "network": "visa"
    },
    "score": 91,
    "rank": 1,
    "eligibility": {
      "status": "high_confidence",
      "eligible": true,
      "reasons": [
        "You do not appear to currently hold this card.",
        "No recent matching card history was found."
      ],
      "warnings": []
    },
    "valueBreakdown": {
      "signupBonusValueCents": 72000,
      "requiredSpendPointsValueCents": 4500,
      "travelCreditValueCents": 0,
      "annualFeeCents": 29500,
      "netEstimatedValueCents": 47000
    },
    "spendRequirement": {
      "minimumSpendCents": 300000,
      "spendWindowDays": 90,
      "projectedUserSpendCents": 850000,
      "achievable": true,
      "difficulty": "easy"
    },
    "reasons": [
      "This card has the strongest combination of estimated net value and achievable spend requirement.",
      "The minimum spend requirement fits comfortably within your projected spend over the offer window.",
      "The reward program aligns with your Qantas points goal."
    ],
    "actionChecklist": [
      "Review the issuer terms before applying.",
      "Apply before the offer expiry date if still available.",
      "Spend $3,000 within 90 days to qualify for the bonus.",
      "Review the card before the next annual fee is charged."
    ]
  },
  "alternatives": [],
  "ineligibleOrCautionCards": [],
  "roadmap": [
    {
      "month": 0,
      "title": "Apply for recommended card",
      "description": "Start with the highest-scoring eligible offer."
    },
    {
      "month": 3,
      "title": "Complete minimum spend",
      "description": "Aim to meet the spend requirement before the deadline."
    },
    {
      "month": 11,
      "title": "Review annual fee",
      "description": "Decide whether to keep, downgrade, or close before renewal."
    },
    {
      "month": 12,
      "title": "Consider next card",
      "description": "Re-run the assistant to find the next eligible sign-up bonus."
    }
  ]
}
```

### 10.4 Get recommendation run

Optional but recommended if persisting runs.

```http
GET /api/recommendations/{id}
```

Response:

- Return the stored result snapshot.

---

## 11. Database Design

### 11.1 `card_offers`

Stores curated offer data.

```sql
CREATE TABLE card_offers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  issuer TEXT NOT NULL,
  card_name TEXT NOT NULL,
  reward_program TEXT NOT NULL,
  network TEXT NOT NULL,

  signup_bonus_points INTEGER NOT NULL DEFAULT 0,
  estimated_bonus_value_cents INTEGER NOT NULL DEFAULT 0,
  minimum_spend_cents INTEGER NOT NULL DEFAULT 0,
  spend_window_days INTEGER NOT NULL DEFAULT 90,

  annual_fee_cents INTEGER NOT NULL DEFAULT 0,
  travel_credit_cents INTEGER NOT NULL DEFAULT 0,
  purchase_rate_percent NUMERIC(5,2),

  offer_expires_at TIMESTAMPTZ,
  eligibility_rules JSONB NOT NULL DEFAULT '[]'::jsonb,
  earn_rates JSONB NOT NULL DEFAULT '[]'::jsonb,
  terms_summary JSONB NOT NULL DEFAULT '[]'::jsonb,

  source_url TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Use JSONB for `eligibility_rules`, `earn_rates`, and `terms_summary` to avoid overfitting the schema too early.

This is a deliberate trade-off:

- Structured columns are used for fields needed by the recommendation engine.
- JSONB is used for messy issuer-specific details that are important to show but not always used in scoring.

### 11.2 `recommendation_runs`

Stores snapshots for reproducibility and auditability.

```sql
CREATE TABLE recommendation_runs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  input_snapshot JSONB NOT NULL,
  result_snapshot JSONB NOT NULL,
  best_card_offer_id UUID REFERENCES card_offers(id),
  estimated_year_one_value_cents INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Reasoning:

- Offers may change over time.
- Storing snapshots allows old recommendations to remain explainable even if seed data changes later.
- This is a production-minded touch without adding much complexity.

---

## 12. Domain Types

Use these as conceptual Go types. Adjust names as needed.

```go
type RewardProgram string

const (
    RewardQantas         RewardProgram = "qantas_points"
    RewardVelocity       RewardProgram = "velocity_points"
    RewardCashback       RewardProgram = "cashback"
    RewardFlexiblePoints RewardProgram = "flexible_points"
    RewardTravelPerks    RewardProgram = "travel_perks"
    RewardNoPreference   RewardProgram = "no_preference"
)
```

```go
type OptimizationGoal string

const (
    GoalMaxNetValue OptimizationGoal = "max_net_value"
    GoalQantas      OptimizationGoal = "qantas_points"
    GoalVelocity    OptimizationGoal = "velocity_points"
    GoalCashback    OptimizationGoal = "cashback"
    GoalLowEffort   OptimizationGoal = "low_effort"
)
```

```go
type AnnualFeePreference string

const (
    AnnualFeeStrictMax AnnualFeePreference = "strict_max"
    AnnualFeePreferLow AnnualFeePreference = "prefer_low"
    AnnualFeeFlexible  AnnualFeePreference = "flexible"
)
```

```go
type CardNetwork string

const (
    NetworkVisa       CardNetwork = "visa"
    NetworkMastercard CardNetwork = "mastercard"
    NetworkAmex       CardNetwork = "amex"
)
```

```go
type CardOffer struct {
    ID                       string
    Issuer                   string
    CardName                 string
    RewardProgram            RewardProgram
    Network                  CardNetwork
    SignupBonusPoints         int
    EstimatedBonusValueCents  int
    MinimumSpendCents         int
    SpendWindowDays           int
    AnnualFeeCents            int
    TravelCreditCents         int
    OfferExpiresAt            *time.Time
    EligibilityRules          []EligibilityRule
    EarnRates                 []EarnRate
    TermsSummary              []string
    SourceURL                 *string
    IsActive                  bool
}
```

```go
type RecommendationInput struct {
    OptimizationGoal                         OptimizationGoal
    MonthlySpendCents                        int
    ExpectedLargePurchasesNext90DaysCents    int
    SpendingCategories                       []string
    AnnualFeePreference                      AnnualFeePreference
    MaxAnnualFeeCents                        int
    AcceptsAmex                              bool
    CardHistory                              []UserCardHistoryItem
}
```

```go
type UserCardHistoryItem struct {
    Issuer        string
    CardName      string
    OpenedAt      *time.Time
    ClosedAt      *time.Time
    CurrentlyHeld bool
}
```

```go
type RecommendationCandidate struct {
    Offer              CardOfferSummary
    Rank               int
    Score              int
    Eligibility        EligibilityResult
    ValueBreakdown     ValueBreakdown
    SpendRequirement   SpendRequirementResult
    Reasons            []string
    Warnings           []string
    ActionChecklist    []string
}
```

---

## 13. Recommendation Engine Design

The recommendation engine should be deterministic and explainable.

Pipeline:

```text
Load active offers
  ↓
Evaluate eligibility
  ↓
Calculate value
  ↓
Assess spend achievability
  ↓
Calculate score
  ↓
Sort and rank
  ↓
Generate explanations
  ↓
Create roadmap
  ↓
Persist recommendation run snapshot
```

### 13.1 Eligibility evaluation

Eligibility should return a status, not just true/false.

```go
type EligibilityStatus string

const (
    EligibilityHighConfidence EligibilityStatus = "high_confidence"
    EligibilityMediumConfidence EligibilityStatus = "medium_confidence"
    EligibilityLowConfidence EligibilityStatus = "low_confidence"
    EligibilityIneligible EligibilityStatus = "ineligible"
    EligibilityManualReview EligibilityStatus = "manual_review"
)
```

```go
type EligibilityResult struct {
    Eligible bool
    Status   EligibilityStatus
    Reasons  []string
    Warnings []string
}
```

Rules to implement:

1. If user currently holds same card, mark ineligible.
2. If user does not accept Amex and card network is Amex, mark ineligible.
3. If `annualFeePreference` is `strict_max` and annual fee exceeds `maxAnnualFeeCents`, mark ineligible.
4. If user recently held the same card or issuer and an exclusion rule applies, mark medium/low confidence or ineligible depending on rule.
5. If eligibility rules contain `manual_review`, include a warning.

Example eligibility rules in JSONB:

```json
[
  {
    "type": "not_held_recently",
    "windowDays": 730,
    "description": "Offer may not be available if you held this issuer's rewards cards in the last 24 months."
  },
  {
    "type": "new_amex_card_members_only",
    "windowDays": 540,
    "description": "Offer applies to new American Express Card Members only."
  },
  {
    "type": "manual_review",
    "description": "Eligibility depends on issuer-specific terms."
  }
]
```

Eligibility rules should be curated as structured data in `data/card_offers_curated.yaml`, then passed through to SQL JSONB. Do not infer rule types from legal text inside the application path; if a rule is ambiguous, encode it as `manual_review` and surface a warning.

### 13.2 Spend achievability

Calculate whether the user can realistically meet the sign-up bonus spend requirement.

For each offer:

```text
spendWindowMonths = ceil(spendWindowDays / 30)
proratedLargePurchases = expectedLargePurchasesNext90DaysCents × min(spendWindowDays, 90) / 90
projectedSpend = monthlySpendCents × spendWindowMonths + proratedLargePurchases
```

`expectedLargePurchasesNext90DaysCents` represents purchases the user expects to make within the next three months. Prorate it to shorter offer windows so a 90-day large-purchase estimate is not fully credited against a 14- or 30-day spend window. Cap the proration at 90 days because the user did not provide a large-purchase forecast beyond that period.

Then compare:

```text
projectedSpend / minimumSpend
```

Suggested difficulty thresholds:

- >= 1.50: easy
- >= 1.10: achievable
- >= 0.85: tight
- < 0.85: unlikely

These thresholds are intentionally conservative. Most Australian issuers exclude categories such as government payments, BPAY, ATO payments, gift cards, balance transfers, and refunds from minimum-spend calculations. A projection that only just clears the stated minimum can still miss the bonus in practice.

Type:

```go
type SpendDifficulty string

const (
    SpendEasy       SpendDifficulty = "easy"
    SpendAchievable SpendDifficulty = "achievable"
    SpendTight      SpendDifficulty = "tight"
    SpendUnlikely   SpendDifficulty = "unlikely"
)
```

```go
type SpendRequirementResult struct {
    MinimumSpendCents       int
    SpendWindowDays         int
    ProjectedUserSpendCents int
    Achievable              bool
    Difficulty              SpendDifficulty
    Reason                  string
}
```

Important UX point:

Do not recommend a card as the top choice if the user is unlikely to meet the minimum spend, even if the card has high theoretical value.

### 13.3 Value calculation

Use a simple and explainable formula:

```text
netEstimatedValue = signupBonusValue + requiredSpendPointsValue + travelCreditValue - annualFee
```

For this version, `requiredSpendPointsValue` can be estimated simply:

```text
requiredSpendPointsValue = minimumSpend × approximateEarnRateValue
```

If this is too time-consuming, set `requiredSpendPointsValue` to 0 or a simple estimate and explain the limitation.

Recommended type:

```go
type ValueBreakdown struct {
    SignupBonusValueCents        int
    RequiredSpendPointsValueCents int
    TravelCreditValueCents       int
    AnnualFeeCents               int
    NetEstimatedValueCents       int
}
```

### 13.3.1 Estimated year-one value

`estimatedYearOneValueCents` represents the total projected net value of the recommended 12-month action plan.

For this version, the roadmap recommends one immediate card and then asks the user to re-run the assistant later. In that case:

```text
estimatedYearOneValueCents = bestRecommendation.valueBreakdown.netEstimatedValueCents
```

The definition supports future multi-card roadmap recommendations. If a later version recommends multiple cards within the next 12 months, `estimatedYearOneValueCents` should equal the sum of the net estimated values of those roadmap cards. It should not represent the total value of every eligible offer, because the user cannot realistically capture every offer at once.

### 13.4 Scoring

Do not rank only by net estimated value. Ranking should balance value with practicality.

Suggested scoring out of 100:

```text
weightedScore = valueScore + spendScore + goalScore + eligibilityScore + annualFeeComfortScore
score = clamp(weightedScore + urgencyTiebreakerBonus, 0, 100)
```

Weights for each optimisation goal should sum to 100 so the weighted portion of the score reads as a clear percentage. Default weights:

- Value score: 45 points
- Spend achievability: 20 points
- Reward goal match: 15 points
- Eligibility confidence: 15 points
- Annual fee comfort: 5 points

Offer urgency is intentionally **not** a weighted axis. A soon-to-expire mediocre offer should not outrank a stronger, longer-window offer. Urgency is instead applied as a small post-hoc bonus capped at roughly 3 points so it only changes the ordering of cards whose underlying recommendation quality is already similar.

`optimisationGoal` should adjust these weights while keeping the scoring logic hardcoded and easy to inspect:

- `max_net_value`: prioritise net estimated value.
- `qantas_points`: boost Qantas reward-program matches.
- `velocity_points`: boost Velocity reward-program matches.
- `cashback`: boost cashback or gift-card-like reward matches.
- `low_effort`: prioritise easy spend achievability, high eligibility confidence, lower fee friction, and fewer warnings.

Annual fees should usually affect score rather than eligibility:

- `strict_max`: exclude cards above `maxAnnualFeeCents`.
- `prefer_low`: keep higher-fee cards but apply a scoring penalty.
- `flexible`: focus mostly on net value and apply little or no annual fee penalty.

Implementation detail:

Keep scoring simple and readable. The exact weights are less important than being explainable.

Example:

```go
func ScoreCandidate(input RecommendationInput, offer CardOffer, eligibility EligibilityResult, value ValueBreakdown, spend SpendRequirementResult) ScoreResult
```

```go
type ScoreResult struct {
    Score   int
    Reasons []string
    Warnings []string
}
```

### 13.5 Recommendation explanations

Every candidate should include explanation text generated by deterministic logic.

Do not require an LLM API key for the core app.

Example reason generation:

- If highest net value: “This card has one of the strongest net value estimates among eligible offers.”
- If spend easy: “Your projected spend appears comfortably above the minimum spend requirement.”
- If reward match: “The reward program matches your selected goal.”
- If medium confidence: “Some eligibility terms may require manual review before applying.”

#### Score transparency

The raw 0–100 score is not meaningful to a user on its own. The frontend should prefer rank + reasons over a bare score, and if a numeric score is shown it should be accompanied by the same component breakdown the scoring function uses (for example: `Value 38/45 · Spend 16/20 · Goal 15/15 · Eligibility 15/15 · Fee comfort 4/5`). The explanation sentences must stay aligned with the components that actually contributed to the score, so a user can always trace why one card outranked another.

Optional enhancement:

Add an `OPENAI_API_KEY`-guarded endpoint or feature for a natural language explanation. However, the app must work fully without it.

---

## 14. Roadmap Generation

The roadmap makes the product feel like an assistant rather than a comparison table.

For this version, the roadmap recommends one immediate card. It wraps the ranked recommendation result with reasons, warnings, alternatives, caution cards, and an action checklist.

The checklist maps roughly to this timeline:

```text
Month 0: Apply for recommended card
Month 1–3: Meet spend requirement
Month 4: Bonus expected after spend requirement is met
Month 11: Review before annual fee renewal
Month 12: Re-run assistant for next opportunity
```

Domain types:

```go
type ActionChecklistItem struct {
    Kind        ActionChecklistItemKind
    Title       string
    Description string
    DueAt       *time.Time
}

type RecommendationRoadmap struct {
    HasRecommendation          bool
    Summary                    RecommendationSummary
    BestRecommendation         *RecommendationCandidate
    Alternatives               []RecommendationCandidate
    IneligibleOrCautionCards   []RecommendationCandidate
    ActionChecklist            []ActionChecklistItem
    Reasons                    []string
    Warnings                   []string
    NoRecommendationReasons    []string
}
```

If offer has an expiry date, add:

```text
Apply before [expiry date] if the offer is still available.
```

Dates are reminders, not legal deadlines. For example, the minimum-spend due date assumes approval today and should be adjusted once the card is approved.

If no card is safe enough to recommend, return `HasRecommendation = false`, explain why, and retain caution cards so the UI can show why tempting offers were not selected.

---

## 15. Seed Dataset

Create around 12–20 representative Australian card offers.

Important:

- The dataset does not need to be perfect.
- It should be plausible and varied enough to test recommendations.
- Include different issuers, reward programs, annual fees, spend requirements, and networks.

Include examples across:

- Qantas points
- Velocity points
- Flexible bank rewards
- Cashback/gift-card-like rewards
- Amex and non-Amex cards
- Low-fee and high-fee cards
- Easy and difficult spend requirements

Each seed record should include:

- issuer
- card_name
- reward_program
- network
- signup_bonus_points
- estimated_bonus_value_cents
- minimum_spend_cents
- spend_window_days
- annual_fee_cents
- travel_credit_cents
- offer_expires_at optional
- eligibility_rules
- terms_summary
- source_url optional

If using real offer data, include source URLs in the database and README. If using placeholder/mock data, clearly label it as sample data.

---

## 16. Validation Rules

### Frontend validation

- Monthly spend must be greater than 0.
- Expected large purchases in the next 90 days can be 0.
- Max annual fee must be >= 0.
- Optimization goal required.
- Annual fee preference required.
- If card history item has closedAt, it must not be in the future.

### Backend validation

Validate all recommendation request fields server-side as well.

Return errors in a consistent shape:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Average monthly spend must be greater than zero.",
    "fields": {
      "monthlySpendCents": "must be greater than zero"
    }
  }
}
```

---

## 17. Error Handling

Backend should use consistent JSON error responses.

Examples:

- Invalid request body
- Validation failure
- Database unavailable
- No active card offers found
- Recommendation could not be generated

Do not expose raw SQL errors to the frontend.

Use structured logging in the backend where practical.

---

## 18. Security and Privacy Considerations

For this version:

- Do not collect sensitive personal identity data.
- Do not ask for income, credit score, address, or banking credentials.
- Do not store more user data than needed.
- Store reduced recommendation snapshots but avoid personally identifying information.
- Include clear disclaimer that this is an estimate and users should check issuer terms.

Recommendation run snapshots should include only the fields needed to reproduce the recommendation:

- `monthlySpendCents`
- `expectedLargePurchasesNext90DaysCents`
- `optimisationGoal`
- `annualFeePreference`
- `maxAnnualFeeCents`
- `acceptsAmex`
- `spendingCategories`
- `cardHistorySummary`
- recommendation result

Do not collect or store names, emails, phone numbers, raw bank statements, exact transaction data, or banking credentials.

No authentication is required for this take-home project.

Reasoning:

- Auth does not materially improve the core recommendation experience.
- Skipping auth keeps focus on product, recommendation quality, and UX.
- In production, saved profiles and reminders would require accounts.

---

## 19. Production-Minded Design Decisions

Include signs of production thinking:

1. Use PostgreSQL instead of a static JSON file for offers.
2. Keep recommendation logic in backend domain services, not frontend.
3. Store recommendation run snapshots for auditability.
4. Use environment variables for configuration.
5. Provide database migrations and seed data.
6. Validate input on frontend and backend.
7. Add basic backend tests for recommendation logic.
8. Use clear error responses.
9. Add README with assumptions and future work.
10. Keep core functionality deterministic and testable.

---

## 20. Testing Plan

### Backend unit tests

Focus tests on recommendation logic.

Test cases:

1. Highest net value card is recommended when all else is equal.
2. Card is excluded if currently held.
3. Amex card is excluded if user does not accept Amex.
4. High-value card is penalized if spend requirement is unrealistic.
5. Reward goal match improves score.
6. Annual fee above max limit is excluded for `strict_max` and penalized for `prefer_low`.
7. Recommendation response includes reasons and action checklist.
8. Empty offer list returns a useful error.
9. `low_effort` scoring prefers easy spend achievability and high eligibility confidence.

### Frontend tests

If time allows:

- Form validation smoke test
- Results page renders recommendation data

If time is limited, prioritise backend tests.

---

## 21. Walkthrough Narrative

The final written or video walkthrough should cover:

### Problem focus

“I focused on the difficulty users face when deciding which sign-up bonus to target next. The problem is not just comparing cards; it is understanding eligibility, spend requirements, timing, and expected value.”

### Key product decisions

1. Guided onboarding instead of a generic comparison table.
2. One clear top recommendation instead of overwhelming the user.
3. Eligibility confidence instead of pretending certainty.
4. Curated data instead of live ingestion to focus on the core product experience.
5. Roadmap/action plan to make the recommendation actionable.

### Key engineering decisions

1. React frontend and Go REST API.
2. PostgreSQL-backed offer data.
3. Handler → Service → Repository architecture.
4. Recommendation engine split into eligibility, value calculation, scoring, and explanation.
5. Recommendation run snapshots for reproducibility.

### What would be improved with more time

1. Live offer ingestion pipeline.
2. Admin interface for managing offers.
3. User accounts and saved card history.
4. Email/calendar reminders for annual fee review and bonus deadlines.
5. Transaction-based spend analysis.
6. More robust eligibility rules.
7. LLM-assisted explanation layer with deterministic fallback.
8. Offer freshness monitoring.

---

## 22. Acceptance Criteria

A coding agent should consider the feature complete when:

1. The app runs locally with one command for frontend and backend, plus PostgreSQL via Docker Compose.
2. The database has migrations and seed data for at least 12 card offers.
3. The frontend has a polished onboarding form.
4. Submitting the form calls the Go backend.
5. The backend returns a recommendation with:
   - best card
   - alternatives
   - value breakdown
   - eligibility result
   - spend achievability
   - reasons
   - warnings
   - action checklist
   - roadmap
6. The frontend renders the recommendation clearly.
7. The backend has unit tests for core recommendation logic.
8. README explains setup, assumptions, trade-offs, and future improvements.
9. The app includes a disclaimer that estimates are not financial advice and issuer terms should be checked.
10. The implementation avoids unnecessary complexity such as auth, scraping, and microservices.

---

## 23. Implementation Order for Codex

Implement in this order:

1. Create repo structure.
2. Create Docker Compose for PostgreSQL.
3. Create backend config and DB connection.
4. Create migrations for `card_offers` and `recommendation_runs`.
5. Add seed data.
6. Implement backend domain types.
7. Implement card offer repository and list endpoint.
8. Implement recommendation service:
   - eligibility
   - spend achievability
   - value calculation
   - scoring
   - explanation
   - roadmap
9. Implement recommendation endpoint.
10. Add backend tests.
11. Create frontend app shell.
12. Build onboarding form.
13. Build API client.
14. Build results page.
15. Polish UI copy and formatting.
16. Add README.
17. Add `.env.example`.
18. Final smoke test.

### 23.1 Collaborative AI Build Slices

The project scope is intentionally ambitious enough to demonstrate effective AI-assisted delivery. To avoid comprehension debt, implementation should happen in vertical slices with explicit checkpoints rather than one large generated code drop.

Each slice should produce a working, inspectable increment. After each slice, the developer should understand the new concepts introduced before moving on.

Current slices:

1. Domain recommendation engine and roadmap
   - Implement core Go recommendation types against in-memory offers.
   - Add value calculation, spend achievability, eligibility confidence, scoring, ranking, action checklist, and roadmap generation.
   - Add focused tests for the domain logic and an in-memory end-to-end recommendation flow.

2. Data and database foundation
   - Add PostgreSQL, migrations, Docker Compose, and curated card-offer seed data.
   - Keep curated YAML as the human-readable source of truth and generate idempotent SQL seed data.
   - Store structured scoring fields as columns and issuer-specific rules/terms as JSONB.

3. Backend recommendation API
   - Implement the card offer repository using plain `pgx`.
   - Add `POST /api/recommendations` with backend validation.
   - Load active offers from Postgres, run the domain engine, and return the full recommendation roadmap shape.
   - Add handler tests with a fake repository and a local smoke test against seeded Postgres.

4. Backend completeness
   - Add `GET /api/card-offers` for inspecting active seeded offers.
   - Persist reduced recommendation run snapshots in `recommendation_runs` after successful recommendation responses.
   - Tighten structured API errors and add focused tests for the new behaviour.
   - Add a repository-level smoke or integration test path if it stays lightweight.

5. Frontend integration skeleton
   - Create the React app shell, API client, and a simple onboarding form.
   - Submit real user input to the backend through the Vite dev proxy.
   - Render the recommendation result clearly before heavy UI polish.

6. Product polish
   - Turn the smoke UI into a more guided decision-assistant experience.
   - Improve copy, formatting, empty states, disclaimers, alternatives, and caution cards.
   - Keep the UX focused on one clear next action.

7. Final hardening
   - Finish README/setup instructions, assumptions, trade-offs, and future work.
   - Run backend and frontend checks plus a final local smoke test.
   - Confirm the project demonstrates product thinking, engineering quality, and AI-assisted execution without unnecessary complexity.

### 23.2 Resolved Product Contract

These decisions guide implementation. The goal is to build a simple, explainable, production-minded product that focuses on the recommendation engine, user experience, and decision-making workflow.

Do not over-engineer live data ingestion, scraping, authentication, complex financial modelling, or multi-card sequencing. Use a curated dataset of Australian credit card offers and focus on the core product experience.

#### Product contract

The app will help a user answer:

```text
Which Australian credit card sign-up bonus should I target next, and why?
```

The app must provide:

1. A guided input flow for goals, spending profile, card history, and preferences.
2. One primary recommended card.
3. Estimated net value for the recommendation.
4. Eligibility confidence, not approval certainty.
5. Spend requirement achievability.
6. Explanation bullets generated by deterministic backend logic.
7. A short action checklist and 12-month roadmap.
8. A small set of alternatives and caution/ineligible cards.
9. A clear disclaimer that the result is not financial advice.

For this version, the roadmap recommends one immediate card. Future versions may recommend multiple cards within a 12-month plan, but multi-card sequencing is not required for this take-home.

#### Resolved value contract

`estimatedYearOneValueCents` represents the total projected net value of the recommended 12-month action plan.

For this version:

- If only one immediate card is recommended, this equals the best card's net estimated value.
- It should not represent the total value of every eligible offer, because the user cannot realistically capture every offer at once.

Future multi-card roadmap support:

- If a later version recommends multiple cards within the next 12 months, this equals the sum of the net estimated values of those roadmap cards.

Example:

```text
Month 0: Card A, estimated net value $650
Month 6: Card B, estimated net value $500
estimatedYearOneValueCents = 115000
```

#### Backend contract

The backend must expose:

1. `GET /health`
2. `GET /api/card-offers`
3. `POST /api/recommendations`

The recommendation engine must be deterministic and testable. It must evaluate:

1. Current-card exclusion.
2. Amex acceptance.
3. Annual fee preference.
4. Recent-cardholder eligibility rules where present in curated data.
5. Minimum spend achievability.
6. Estimated net value.
7. Reward goal match.

Supported optimisation goals:

- `max_net_value`
- `qantas_points`
- `velocity_points`
- `cashback`
- `low_effort`

Hardcoded weights are acceptable. The backend should return clear explanation reasons for why a card ranked highly.

Annual fee handling:

- `annualFeePreference = "strict_max"`: exclude cards above `maxAnnualFeeCents`.
- `annualFeePreference = "prefer_low"`: keep cards above the preferred fee range but apply a scoring penalty.
- `annualFeePreference = "flexible"`: focus mostly on net value and apply little or no annual fee penalty.

A high-fee card may still be a strong recommendation if the sign-up bonus and credits create strong net value. The system should avoid hiding valuable cards unless the user explicitly chooses a strict maximum annual fee.

#### Frontend contract

The frontend must include:

1. A landing page with clear value proposition and disclaimer.
2. A guided onboarding form with no more than 5 steps.
3. A results view that leads with estimated value and the best recommendation.
4. A readable value breakdown.
5. Reasons, warnings, alternatives, and action plan.

Spending categories may be collected during onboarding, but they should not be central to scoring. For scoring, spend achievability should mainly use:

- `monthlySpendCents`
- `expectedLargePurchasesNext90DaysCents`

Spending categories can be used for UX context or future personalisation, but not detailed earn-rate calculations.

User-facing label for large purchases:

```text
Any large purchases expected in the next 3 months?
```

#### Data contract

The app will use curated PostgreSQL seed data with 12-20 representative Australian card offers.

Each offer must include enough structured data to support the recommendation engine:

1. Issuer and card name.
2. Reward program and card network.
3. Sign-up bonus points and estimated value.
4. Minimum spend and spend window.
5. Annual fee and travel credit.
6. Eligibility rules.
7. Terms summary.
8. Optional source URL.

#### Snapshot contract

Store a reduced, non-identifying snapshot of each recommendation run. The snapshot should include only the fields needed to reproduce the recommendation:

- `monthlySpendCents`
- `expectedLargePurchasesNext90DaysCents`
- `optimisationGoal`
- `annualFeePreference`
- `maxAnnualFeeCents`
- `acceptsAmex`
- `spendingCategories`
- `cardHistorySummary`
- recommendation result

`cardHistorySummary` should be structured data, not free text. For this version, it should contain issuer, card name, optional opened date, optional closed date, and whether the card is currently held.

Do not collect or store personal identifying information such as name, email, phone number, raw bank statements, exact transaction data, or banking credentials.

#### Final implementation principle

Keep the system architecture simple:

```text
React frontend
→ Go REST API
→ Recommendation service
→ PostgreSQL
```

The product complexity should live in the domain layer:

```text
Card offers
→ Eligibility evaluation
→ Value calculation
→ Scoring
→ Explainable recommendation
→ 12-month roadmap
```

The strongest version of this product is not the most technically complex version. It is the version that gives the user a clear, trustworthy answer to:

```text
What card should I target next, why, and how much value could I get?
```

---

## 24. Coding Style Guidelines

### Backend

- Keep handlers thin.
- Keep business logic out of handlers.
- Prefer explicit types over loose maps.
- Use context-aware DB methods.
- Return structured errors.
- Keep scoring functions small and testable.
- Do not hide recommendation logic in SQL.

### Frontend

- Keep form state clear and simple.
- Use small reusable components.
- Format money consistently.
- Avoid excessive animations.
- Make the recommendation easy to scan.
- Use plain language over finance jargon.

---

## 25. UI Copy Examples

### Hero

```text
Find your next best credit card bonus
Estimate which sign-up bonus is worth targeting based on your spending, card history, and rewards goals.
```

### CTA

```text
Find my next card
```

### Results summary

```text
You could earn an estimated $1,240 in net value over the next 12 months.
```

### Recommendation explanation

```text
This card is recommended because it offers the strongest combination of net value, achievable minimum spend, and eligibility confidence for your profile.
```

### Disclaimer

```text
This tool provides estimates based on curated offer data and simplified assumptions. It is not financial advice. Always check issuer terms and consider your personal circumstances before applying.
```

---

## 26. Final Notes

The project should feel like a real startup product: simple infrastructure, strong domain logic, polished UX, and clear trade-offs.

The most important thing is not to build the most complex system. The most important thing is to demonstrate that the user problem has been understood and translated into a product that gives clear, actionable, trustworthy guidance.
