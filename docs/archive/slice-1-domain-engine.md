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
- Roadmap/action checklist generation.
- End-to-end in-memory domain tests.
- Focused unit tests around the domain logic.

Out of scope for this slice:

- API endpoints.
- Postgres repositories.
- Frontend form wiring.
- Multi-card sequencing.
- Credit-score modelling.
- Points redemption strategy.

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
- `backend/internal/recommendations/checklist.go`
- `backend/internal/recommendations/checklist_test.go`
- `backend/internal/recommendations/roadmap.go`
- `backend/internal/recommendations/roadmap_test.go`

Implemented concepts:

- `RecommendationInput`: backend representation of onboarding inputs.
- `CardOffer`: backend representation of one card offer.
- `ValueBreakdown`: estimated value components.
- `SpendRequirementResult`: projected spend and difficulty.
- `EligibilityResult`: eligibility confidence and warnings.
- `ScoreResult`: score, reasons, and warnings for a candidate.
- `RecommendationCandidate`: one fully evaluated card offer.
- `RecommendationResult`: best card, alternatives, caution cards, and summary.
- `ActionChecklistItem`: one deterministic action for the user.
- `RecommendationRoadmap`: final single-card plan built from a ranked result.

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

`RequiredSpendPointsValueCents` is deliberately `0` for this version. Earn rates vary by card, category, caps, exclusions, and government payments. Without richer spend data, including those points would create false precision.

Later bonuses are not included in `CalculateValue`. They are kept in card terms for explanation, but excluded from year-one value unless future roadmap logic explicitly models them.

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

Card history is self-reported. The engine treats it as an eligibility signal, not proof. To reduce avoidable mismatches, issuer comparison normalises common aliases such as `Amex`/`American Express`, `CommBank`/`Commonwealth Bank`, and the St.George/BankSA/Bank of Melbourne regional issuer group. Card names remain exact text matches for same-card hard exclusions.

When self-reported history triggers a rule, the warning starts with `Your card history may affect eligibility` so the frontend can surface a dedicated card-history impact section.

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

`BestRecommendation` is the first ranked recommendable candidate. `Alternatives` contains up to three next-best recommendable cards. `Summary.EstimatedYearOneValueCents` is the best card's net estimated value, matching the product contract that the first roadmap recommends one immediate card.

## Action Checklist

`BuildActionChecklist` turns the best candidate into concrete next steps. It is deterministic and derived from existing domain data, not AI-generated prose.

Checklist items currently include:

- Verify the current public offer terms.
- Review eligibility and spend cautions when warnings exist.
- Apply for the recommended card.
- Spend the required amount within the offer window.
- Track bonus posting.
- Review the card before the next annual fee.
- Review later bonus conditions separately when present.

Dates are conservative reminders, not legal deadlines. For example, the minimum-spend due date assumes approval today and should be adjusted once the card is actually approved.

Later bonuses are surfaced as review items only. They remain excluded from year-one value.

## Roadmap Generation

`BuildRoadmap` wraps a `RecommendationResult` into the final domain-level plan:

- Best recommendation.
- Estimated year-one value summary.
- Action checklist.
- Reasons.
- Warnings.
- Alternatives.
- Ineligible or caution cards.

If there is no safe best recommendation, the roadmap returns `HasRecommendation = false` and explains that no card is safe enough to recommend from the current inputs. Caution cards are still retained so the UI can explain why tempting cards were not recommended.

The roadmap intentionally recommends one immediate card only. It does not yet sequence multiple cards, model credit-score timing, or produce a long-term points strategy.

## In-Memory Service Test

`roadmap_test.go` includes a service-style test that exercises the full domain flow without HTTP, Postgres, or frontend code:

```text
RecommendationInput + []CardOffer
→ Recommend
→ BuildRoadmap
→ best card + alternatives + caution cards + checklist
```

This is the end-to-end boundary for Slice 1. Later slices can add repository/API tests around the same domain functions without changing the engine's decision rules.

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

Slice 1 domain work is complete enough to move to the next vertical slice.

Natural next slices:

- Load active card offers from Postgres into `CardOffer` structs.
- Add an HTTP endpoint that accepts onboarding input and returns the roadmap.
- Wire a small frontend form to the endpoint.
