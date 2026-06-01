import { useState } from 'react'
import type { FormEvent } from 'react'
import './App.css'
import { createRecommendation } from './api'
import { dollarsToCents, formatCents, formatDate, formatRewardType } from './format'
import type {
  AnnualFeePreference,
  OptimisationGoal,
  RecommendationCandidate,
  RecommendationInput,
  RecommendationRoadmap,
} from './types'

type FormState = {
  optimisationGoal: OptimisationGoal
  monthlySpendDollars: string
  largePurchasesDollars: string
  annualFeePreference: AnnualFeePreference
  maxAnnualFeeDollars: string
  acceptsAmex: boolean
}

const initialForm: FormState = {
  optimisationGoal: 'qantas_points',
  monthlySpendDollars: '2500',
  largePurchasesDollars: '1000',
  annualFeePreference: 'flexible',
  maxAnnualFeeDollars: '450',
  acceptsAmex: true,
}

function App() {
  const [form, setForm] = useState<FormState>(initialForm)
  const [roadmap, setRoadmap] = useState<RecommendationRoadmap | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError(null)
    setIsLoading(true)

    const input: RecommendationInput = {
      optimisationGoal: form.optimisationGoal,
      monthlySpendCents: dollarsToCents(form.monthlySpendDollars),
      expectedLargePurchasesNext90DaysCents: dollarsToCents(form.largePurchasesDollars),
      annualFeePreference: form.annualFeePreference,
      maxAnnualFeeCents: dollarsToCents(form.maxAnnualFeeDollars),
      acceptsAmex: form.acceptsAmex,
    }

    try {
      const nextRoadmap = await createRecommendation(input)
      setRoadmap(nextRoadmap)
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Could not create recommendation.')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <span className="brand-mark">SB</span>
        <span className="pill pill-warm">POINTS ENGINE</span>
      </header>

      <section className="hero-panel">
        <p className="mono-line">132 OFFERS REVIEWED TODAY</p>
        <h1>There is a strong card offer for you.</h1>
        <p className="hero-copy">
          Tell us your spend and preferences. We will check the seeded offers through the
          Go recommendation engine and return a roadmap.
        </p>
      </section>

      <form className="wizard-card" onSubmit={handleSubmit}>
        <div className="form-row">
          <label htmlFor="optimisationGoal">What are you optimising for?</label>
          <select
            id="optimisationGoal"
            value={form.optimisationGoal}
            onChange={(event) =>
              setForm((current) => ({
                ...current,
                optimisationGoal: event.target.value as OptimisationGoal,
              }))
            }
          >
            <option value="qantas_points">Qantas Points</option>
            <option value="velocity_points">Velocity Points</option>
            <option value="max_net_value">Maximum net value</option>
            <option value="cashback">Cashback</option>
            <option value="low_effort">Low effort</option>
          </select>
        </div>

        <MoneyInput
          id="monthlySpend"
          label="Monthly card spend"
          value={form.monthlySpendDollars}
          onChange={(value) => setForm((current) => ({ ...current, monthlySpendDollars: value }))}
        />

        <MoneyInput
          id="largePurchases"
          label="Large purchases in the next 90 days"
          value={form.largePurchasesDollars}
          onChange={(value) => setForm((current) => ({ ...current, largePurchasesDollars: value }))}
        />

        <div className="form-row">
          <label htmlFor="annualFeePreference">Annual fee preference</label>
          <select
            id="annualFeePreference"
            value={form.annualFeePreference}
            onChange={(event) =>
              setForm((current) => ({
                ...current,
                annualFeePreference: event.target.value as AnnualFeePreference,
              }))
            }
          >
            <option value="flexible">Flexible if value is strong</option>
            <option value="prefer_low">Prefer lower fees</option>
            <option value="strict_max">Strict maximum</option>
          </select>
        </div>

        {form.annualFeePreference === 'strict_max' ? (
          <MoneyInput
            id="maxAnnualFee"
            label="Maximum annual fee"
            value={form.maxAnnualFeeDollars}
            onChange={(value) => setForm((current) => ({ ...current, maxAnnualFeeDollars: value }))}
          />
        ) : null}

        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={form.acceptsAmex}
            onChange={(event) =>
              setForm((current) => ({ ...current, acceptsAmex: event.target.checked }))
            }
          />
          <span>I am open to American Express cards</span>
        </label>

        <button className="primary-button" type="submit" disabled={isLoading}>
          {isLoading ? 'Searching offers...' : 'Find my best offer'}
        </button>
      </form>

      {error ? <div className="error-card">{error}</div> : null}

      {roadmap ? <RoadmapView roadmap={roadmap} /> : null}
    </main>
  )
}

type MoneyInputProps = {
  id: string
  label: string
  value: string
  onChange: (value: string) => void
}

function MoneyInput({ id, label, value, onChange }: MoneyInputProps) {
  return (
    <div className="form-row">
      <label htmlFor={id}>{label}</label>
      <div className="money-input">
        <span>$</span>
        <input
          id={id}
          type="number"
          min="0"
          inputMode="decimal"
          value={value}
          onChange={(event) => onChange(event.target.value)}
        />
      </div>
    </div>
  )
}

function RoadmapView({ roadmap }: { roadmap: RecommendationRoadmap }) {
  if (!roadmap.hasRecommendation || !roadmap.bestRecommendation) {
    return (
      <section className="section-block">
        <p className="mono-line">NO SAFE OFFER FOUND</p>
        <h2>No card is safe enough to recommend yet.</h2>
        <ListItems items={roadmap.noRecommendationReasons} />
      </section>
    )
  }

  const best = roadmap.bestRecommendation

  return (
    <section className="results-stack">
      <BestRecommendationCard candidate={best} cardsConsidered={roadmap.summary.cardsConsidered} />

      <section className="section-block">
        <div className="section-heading">
          <h2>Why this card</h2>
          <span className="pill">SCORE {best.score}</span>
        </div>
        <ListItems items={roadmap.reasons} />
        <ListItems items={roadmap.warnings} tone="warning" />
      </section>

      <section className="section-block">
        <div className="section-heading">
          <h2>Action checklist</h2>
          <span className="pill">NEXT STEPS</span>
        </div>
        <div className="checklist">
          {(roadmap.actionChecklist ?? []).map((item) => (
            <article className="checklist-item" key={`${item.kind}-${item.title}`}>
              <div className="check-icon">✓</div>
              <div>
                <h3>{item.title}</h3>
                <p>{item.description}</p>
                {item.dueAt ? <span className="due-date">Due around {formatDate(item.dueAt)}</span> : null}
              </div>
            </article>
          ))}
        </div>
      </section>

      <CandidateList title="Alternatives" candidates={roadmap.alternatives ?? []} />
      <CandidateList
        title="Caution cards"
        candidates={(roadmap.ineligibleOrCautionCards ?? []).slice(0, 4)}
        caution
      />
    </section>
  )
}

function BestRecommendationCard({
  candidate,
  cardsConsidered,
}: {
  candidate: RecommendationCandidate
  cardsConsidered: number
}) {
  const { offer, valueBreakdown, spendRequirement } = candidate

  return (
    <section className="best-card">
      <p className="mono-line">WE SEARCHED {cardsConsidered} PRODUCTS THIS MONTH</p>
      <article className="offer-card offer-card-featured">
        <span className="pill pill-hot">BEST CARD FOUND</span>
        <div className="issuer-row">
          <div className="diamond" aria-hidden="true" />
          <div>
            <h2>{offer.cardName}</h2>
            <p>{offer.issuer}</p>
          </div>
        </div>
        <p className="value-statement">
          Estimated year-one value of {formatCents(valueBreakdown.netEstimatedValueCents)}
        </p>
        <div className="metadata-grid">
          <span>{offer.signupBonusPoints.toLocaleString('en-AU')} points</span>
          <span>{formatCents(offer.annualFeeCents)} fee</span>
          <span>{formatCents(spendRequirement.minimumSpendCents)} spend</span>
          <span>{spendRequirement.difficulty}</span>
        </div>
      </article>
    </section>
  )
}

function CandidateList({
  title,
  candidates,
  caution = false,
}: {
  title: string
  candidates: RecommendationCandidate[]
  caution?: boolean
}) {
  if (candidates.length === 0) {
    return null
  }

  return (
    <section className="section-block">
      <h2>{title}</h2>
      <div className="compact-list">
        {candidates.map((candidate) => (
          <article className="offer-card compact-card" key={`${candidate.offer.issuer}-${candidate.offer.cardName}`}>
            <span className={`pill ${caution ? 'pill-hot' : ''}`}>
              {caution ? 'CAUTION' : `RANK ${candidate.rank}`}
            </span>
            <h3>{candidate.offer.cardName}</h3>
            <p>{candidate.offer.issuer}</p>
            <div className="metadata-grid">
              <span>{formatCents(candidate.valueBreakdown.netEstimatedValueCents)} value</span>
              <span>{formatRewardType(candidate.offer.rewardType)}</span>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}

function ListItems({ items, tone }: { items: string[] | null; tone?: 'warning' }) {
  const values = items?.filter(Boolean) ?? []
  if (values.length === 0) {
    return null
  }

  return (
    <ul className={`reason-list ${tone === 'warning' ? 'warning-list' : ''}`}>
      {values.map((item) => (
        <li key={item}>{item}</li>
      ))}
    </ul>
  )
}

export default App
