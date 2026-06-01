# Slice 1: Domain Recommendation Engine

## Scope

This slice corresponds to `23.1 Collaborative AI Build Slices` in the design spec: Domain-first recommendation prototype.

The goal is to build and understand the recommendation engine in isolation before adding HTTP handlers, repositories, database loading, or frontend UI.

In scope for this slice:

- Core Go domain types.
- Value calculation.
- Spend achievability.
- Eligibility confidence.
- Scoring.
- Candidate orchestration and ranking.
- Focused unit tests around the domain logic.

Out of scope for this slice:

- API endpoints.
- Postgres repositories.
- Frontend form wiring.
- Roadmap/action checklist generation.

## Mental Model

The eventual request flow is:

```text
Onboarding form inputs
→ RecommendationInput
→ evaluate each CardOffer
→ eligibility + value + spend achievability
→ score each candidate
→ rank candidates
→ generate explanations and roadmap
```

The backend does not trust the frontend to decide recommendation quality. The frontend collects inputs; the backend domain layer makes the decision.

## Implemented So Far

Files:

- `backend/internal/recommendations/types.go`
- `backend/internal/recommendations/value.go`
- `backend/internal/recommendations/value_test.go`
- `backend/internal/recommendations/spend.go`
- `backend/internal/recommendations/spend_test.go`
- `backend/internal/recommendations/eligibility.go`
- `backend/internal/recommendations/eligibility_test.go`
- `backend/internal/recommendations/scoring.go`
- `backend/internal/recommendations/scoring_test.go`
- `backend/internal/recommendations/recommend.go`
- `backend/internal/recommendations/recommend_test.go`

Implemented concepts:

- `RecommendationInput`: backend representation of onboarding inputs.
- `CardOffer`: backend representation of one card offer.
- `ValueBreakdown`: estimated value components.
- `SpendRequirementResult`: projected spend and difficulty.
- `EligibilityResult`: eligibility confidence and warnings.
- `ScoreResult`: score, reasons, and warnings for a candidate.
- `RecommendationCandidate`: one fully evaluated card offer.
- `RecommendationResult`: best card, alternatives, caution cards, and summary.

## Value Calculation

Value uses integer cents, not floating-point money.

Point values are represented as scaled integer cents:

```text
qantas_points: 100 = 1.00 cent
velocity_points: 100 = 1.00 cent
bank_points: 40 = 0.40 cents
flexible_points: 80 = 0.80 cents
```

The formula is:

```text
signupBonusValue = signupBonusPoints × pointValue + cashback
netEstimatedValue = signupBonusValue + travelCredit - annualFee
```

`RequiredSpendPointsValueCents` is deliberately `0` for MVP. Earn rates vary by card, category, caps, exclusions, and government payments. Without richer spend data, including those points would create false precision.

Later bonuses are not included in `CalculateValue`. They are kept in card terms for explanation, but excluded from MVP year-one value unless future roadmap logic explicitly models them.

## Spend Achievability

Spend achievability estimates whether the user can meet a card's minimum spend requirement.

Formula:

```text
spendWindowMonths = ceil(spendWindowDays / 30)
proratedLargePurchases = expectedLargePurchasesNext90DaysCents × min(spendWindowDays, 90) / 90
projectedSpend = monthlySpendCents × spendWindowMonths + proratedLargePurchases
```

Large purchases are collected as a 90-day estimate. They are prorated for shorter offer windows so a 90-day purchase plan is not fully counted against a 30-day offer window.

Difficulty thresholds:

- `easy`: projected spend is at least 150% of required spend.
- `achievable`: projected spend is at least 110%.
- `tight`: projected spend is at least 85%.
- `unlikely`: projected spend is below 85%.

The thresholds are intentionally conservative. Most Australian issuers exclude categories such as government payments, BPAY, ATO payments, gift cards, balance transfers, and refunds from minimum-spend calculations. A projection that only just clears the stated minimum can still miss the bonus in practice.

Cards with `unlikely` spend should not become the top recommendation even if their theoretical value is high.

## Eligibility Evaluation

Eligibility returns confidence, not legal certainty.

Hard exclusions:

- User currently holds the same card.
- User rejects Amex and the card is Amex.
- User selected `strict_max` annual fee and the card exceeds `maxAnnualFeeCents`.

Confidence downgrades:

- Recent cardholder exclusions lower confidence and add warnings.
- New Amex member rules check recent Amex history within the curated rule window.
- Manual review rules keep the card eligible but add warnings.

Important design decision:

Eligibility rules are structured in `data/card_offers_curated.yaml` and passed through to SQL JSONB. The application does not infer rule types from legal text. Ambiguous terms should be encoded as `manual_review`.

## Scoring

Scoring ranks practical recommendation quality, not raw value.

Inputs:

- `RecommendationInput`
- `CardOffer`
- `EligibilityResult`
- `ValueBreakdown`
- `SpendRequirementResult`

Weighted components:

- Value.
- Spend achievability.
- Reward goal match.
- Eligibility confidence.
- Annual fee comfort.

Urgency is a small tiebreaker bonus, not a full weighted axis. A mediocre offer should not win just because it expires soon.

`OptimisationGoal` changes what “best” means:

- `max_net_value`: prioritises net estimated value.
- `qantas_points`: boosts Qantas matches.
- `velocity_points`: boosts Velocity matches.
- `cashback`: boosts cashback-like offers.
- `low_effort`: boosts easy spend, high eligibility confidence, and lower fee friction.

Annual fee handling:

- `strict_max` is handled in eligibility as a hard exclusion.
- `prefer_low` is handled in scoring as a penalty.
- `flexible` mostly lets net value dominate.

## Candidate Orchestration

`Recommend` is deliberately thin. It does not add new business rules beyond deciding which evaluated candidates are safe enough to recommend.

For each `CardOffer`, it calls:

- `EvaluateEligibility`
- `CalculateValue`
- `AssessSpendRequirement`
- `ScoreCandidateAt`

Then it builds a `RecommendationCandidate` containing the offer, score, value breakdown, eligibility result, spend result, reasons, and warnings.

Cards are recommendable only when:

- They are eligible.
- Their score is positive.
- Spend difficulty is not `unlikely`.
- Eligibility confidence is `high_confidence` or `medium_confidence`.

Other cards are retained in `IneligibleOrCautionCards` so the UI can explain why a tempting card was not recommended.

Ranking is deterministic:

- Higher score wins.
- Higher net estimated value wins ties.
- Stronger eligibility confidence wins further ties.
- Easier spend wins further ties.
- Lower annual fee wins further ties.
- Issuer and card name provide the final stable order.

`BestRecommendation` is the first ranked recommendable candidate. `Alternatives` contains up to three next-best recommendable cards. `Summary.EstimatedYearOneValueCents` is the best card's net estimated value, matching the MVP contract that the first roadmap recommends one immediate card.

## Current Guarantees

Current checks pass from `backend/`:

```sh
go test ./...
go vet ./...
golangci-lint run ./...
go build ./...
```

The card-offer SQL seed is generated from `data/card_offers_curated.yaml` with:

```sh
.venv/bin/python scripts/generate_card_offer_seed.py
```

## Remaining Work In Slice 1

Still left before Slice 1 is complete:

- Action checklist generation.
- Roadmap generation.
- End-to-end recommendation service tests using in-memory offers once roadmap output exists.
