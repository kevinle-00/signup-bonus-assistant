import { useState } from 'react'
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

type WizardStep =
  | 'intro'
  | 'goal'
  | 'monthlySpend'
  | 'largePurchases'
  | 'annualFee'
  | 'maxAnnualFee'
  | 'amex'
  | 'review'

type FormState = {
  optimisationGoal?: OptimisationGoal
  monthlySpendDollars: string
  largePurchasesDollars: string
  annualFeePreference?: AnnualFeePreference
  maxAnnualFeeDollars: string
  acceptsAmex?: boolean
}

const initialForm: FormState = {
  monthlySpendDollars: '',
  largePurchasesDollars: '0',
  maxAnnualFeeDollars: '',
}

const goalOptions: Array<{ value: OptimisationGoal; label: string; description: string }> = [
  { value: 'qantas_points', label: 'Qantas Points', description: 'Prioritise Qantas sign-up bonuses.' },
  { value: 'velocity_points', label: 'Velocity Points', description: 'Prioritise Velocity sign-up bonuses.' },
  { value: 'max_net_value', label: 'Maximum net value', description: 'Rank by estimated value after fees.' },
  { value: 'cashback', label: 'Cashback', description: 'Prefer offers with direct cash value.' },
  { value: 'low_effort', label: 'Low effort', description: 'Favour easier spend and cleaner eligibility.' },
]

const feeOptions: Array<{ value: AnnualFeePreference; label: string; description: string }> = [
  { value: 'flexible', label: 'Flexible if the value is strong', description: 'Let net value do most of the work.' },
  { value: 'prefer_low', label: 'Prefer lower fees', description: 'Penalise high-fee cards without hiding them.' },
  { value: 'strict_max', label: 'Set a strict maximum', description: 'Exclude cards above your fee limit.' },
]

function App() {
  const [form, setForm] = useState<FormState>(initialForm)
  const [step, setStep] = useState<WizardStep>('intro')
  const [roadmap, setRoadmap] = useState<RecommendationRoadmap | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [fieldError, setFieldError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  function moveTo(nextStep: WizardStep) {
    setFieldError(null)
    setStep(nextStep)
  }

  function goBack() {
    setFieldError(null)
    setStep(previousStep(step, form))
  }

  function startOver() {
    setForm(initialForm)
    setRoadmap(null)
    setError(null)
    setFieldError(null)
    setIsLoading(false)
    setStep('intro')
  }

  function continueFromMoneyStep(nextStep: WizardStep, field: 'monthlySpendDollars' | 'largePurchasesDollars' | 'maxAnnualFeeDollars') {
    const cents = dollarsToCents(form[field])
    if (field === 'monthlySpendDollars' && cents <= 0) {
      setFieldError('Enter a monthly card spend greater than $0.')
      return
    }
    if (field === 'maxAnnualFeeDollars' && cents <= 0) {
      setFieldError('Enter a maximum annual fee greater than $0.')
      return
    }
    if (Number(form[field] || '0') < 0) {
      setFieldError('Enter $0 or more.')
      return
    }
    moveTo(nextStep)
  }

  async function submitRecommendation() {
    if (!form.optimisationGoal || !form.annualFeePreference || form.acceptsAmex === undefined) {
      setFieldError('Review your answers before continuing.')
      return
    }

    setError(null)
    setFieldError(null)
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
        {step !== 'intro' && !roadmap && !isLoading ? (
          <button className="icon-button" type="button" onClick={goBack} aria-label="Go back">
            ‹
          </button>
        ) : (
          <span className="brand-mark">SB</span>
        )}
        <Progress step={step} form={form} roadmap={roadmap} isLoading={isLoading} />
      </header>

      {isLoading ? <LoadingScreen /> : null}
      {!isLoading && error ? <ErrorScreen message={error} onRetry={submitRecommendation} onReview={() => moveTo('review')} /> : null}
      {!isLoading && !error && roadmap ? <RoadmapView roadmap={roadmap} onStartOver={startOver} /> : null}
      {!isLoading && !error && !roadmap ? (
        <WizardScreen
          form={form}
          step={step}
          fieldError={fieldError}
          setForm={setForm}
          moveTo={moveTo}
          continueFromMoneyStep={continueFromMoneyStep}
          submitRecommendation={submitRecommendation}
        />
      ) : null}
    </main>
  )
}

function WizardScreen({
  form,
  step,
  fieldError,
  setForm,
  moveTo,
  continueFromMoneyStep,
  submitRecommendation,
}: {
  form: FormState
  step: WizardStep
  fieldError: string | null
  setForm: React.Dispatch<React.SetStateAction<FormState>>
  moveTo: (step: WizardStep) => void
  continueFromMoneyStep: (nextStep: WizardStep, field: 'monthlySpendDollars' | 'largePurchasesDollars' | 'maxAnnualFeeDollars') => void
  submitRecommendation: () => void
}) {
  if (step === 'intro') {
    return (
      <section className="screen intro-screen">
        <span className="pill pill-warm">POINTS ENGINE</span>
        <div>
          <h1>Find your next best credit card bonus</h1>
          <p className="hero-copy">
            Answer a few questions and we’ll estimate which sign-up bonus is worth targeting next.
          </p>
        </div>
        <button className="primary-button bottom-action" type="button" onClick={() => moveTo('goal')}>
          Start
        </button>
      </section>
    )
  }

  if (step === 'goal') {
    return (
      <QuestionScreen eyebrow="Estimated value depends on your goal" title="What are you optimising for?">
        <OptionList
          options={goalOptions}
          selected={form.optimisationGoal}
          onSelect={(value) => {
            setForm((current) => ({ ...current, optimisationGoal: value }))
            moveTo('monthlySpend')
          }}
        />
      </QuestionScreen>
    )
  }

  if (step === 'monthlySpend') {
    return (
      <QuestionScreen
        eyebrow="Spend profile"
        title="Roughly how much do you put on cards each month?"
        helper="A ballpark is fine. Exclude rent, mortgage repayments, and anything you would not usually put on a card."
      >
        <MoneyInput
          id="monthlySpend"
          value={form.monthlySpendDollars}
          placeholder="2500"
          onChange={(value) => setForm((current) => ({ ...current, monthlySpendDollars: value }))}
        />
        <FieldError message={fieldError} />
        <button
          className="primary-button bottom-action"
          type="button"
          onClick={() => continueFromMoneyStep('largePurchases', 'monthlySpendDollars')}
        >
          Continue
        </button>
      </QuestionScreen>
    )
  }

  if (step === 'largePurchases') {
    return (
      <QuestionScreen
        eyebrow="Next 90 days"
        title="Any large card purchases coming up?"
        helper="Flights, appliances, insurance, or planned bills can make a bonus easier to reach."
      >
        <MoneyInput
          id="largePurchases"
          value={form.largePurchasesDollars}
          placeholder="0"
          onChange={(value) => setForm((current) => ({ ...current, largePurchasesDollars: value }))}
        />
        <FieldError message={fieldError} />
        <button
          className="primary-button bottom-action"
          type="button"
          onClick={() => continueFromMoneyStep('annualFee', 'largePurchasesDollars')}
        >
          Continue
        </button>
      </QuestionScreen>
    )
  }

  if (step === 'annualFee') {
    return (
      <QuestionScreen eyebrow="Fee comfort" title="How should we treat annual fees?">
        <OptionList
          options={feeOptions}
          selected={form.annualFeePreference}
          onSelect={(value) => {
            setForm((current) => ({ ...current, annualFeePreference: value }))
            moveTo(value === 'strict_max' ? 'maxAnnualFee' : 'amex')
          }}
        />
      </QuestionScreen>
    )
  }

  if (step === 'maxAnnualFee') {
    return (
      <QuestionScreen eyebrow="Strict maximum" title="What is the most you are comfortable paying upfront?">
        <MoneyInput
          id="maxAnnualFee"
          value={form.maxAnnualFeeDollars}
          placeholder="200"
          onChange={(value) => setForm((current) => ({ ...current, maxAnnualFeeDollars: value }))}
        />
        <FieldError message={fieldError} />
        <button
          className="primary-button bottom-action"
          type="button"
          onClick={() => continueFromMoneyStep('amex', 'maxAnnualFeeDollars')}
        >
          Continue
        </button>
      </QuestionScreen>
    )
  }

  if (step === 'amex') {
    return (
      <QuestionScreen eyebrow="Network preference" title="Are you open to American Express cards?">
        <OptionList
          options={[
            { value: true, label: 'Yes, include Amex cards', description: 'Include high-value Amex offers when they fit.' },
            { value: false, label: 'No, avoid Amex cards', description: 'Filter Amex out of the recommendation.' },
          ]}
          selected={form.acceptsAmex}
          onSelect={(value) => {
            setForm((current) => ({ ...current, acceptsAmex: value }))
            moveTo('review')
          }}
        />
      </QuestionScreen>
    )
  }

  return (
    <QuestionScreen eyebrow="Review" title="Ready to check the active offers?">
      <div className="review-card">
        <ReviewRow label="Goal" value={labelForGoal(form.optimisationGoal)} />
        <ReviewRow label="Monthly spend" value={formatDollars(form.monthlySpendDollars)} />
        <ReviewRow label="Large purchases" value={formatDollars(form.largePurchasesDollars)} />
        <ReviewRow label="Annual fee" value={labelForFee(form.annualFeePreference)} />
        {form.annualFeePreference === 'strict_max' ? (
          <ReviewRow label="Fee limit" value={formatDollars(form.maxAnnualFeeDollars)} />
        ) : null}
        <ReviewRow label="Amex" value={form.acceptsAmex ? 'Included' : 'Excluded'} />
      </div>
      <FieldError message={fieldError} />
      <button className="primary-button bottom-action" type="button" onClick={submitRecommendation}>
        Find my best offer
      </button>
    </QuestionScreen>
  )
}

function QuestionScreen({
  eyebrow,
  title,
  helper,
  children,
}: {
  eyebrow: string
  title: string
  helper?: string
  children: React.ReactNode
}) {
  return (
    <section className="screen question-screen">
      <p className="eyebrow">{eyebrow}</p>
      <div>
        <h1>{title}</h1>
        {helper ? <p className="helper-copy">{helper}</p> : null}
      </div>
      {children}
    </section>
  )
}

function OptionList<T extends string | boolean>({
  options,
  selected,
  onSelect,
}: {
  options: Array<{ value: T; label: string; description: string }>
  selected?: T
  onSelect: (value: T) => void
}) {
  return (
    <div className="option-list">
      {options.map((option) => (
        <button
          className={`option-card ${selected === option.value ? 'option-card-selected' : ''}`}
          key={String(option.value)}
          type="button"
          onClick={() => onSelect(option.value)}
        >
          <span>{option.label}</span>
          <small>{option.description}</small>
          <b>→</b>
        </button>
      ))}
    </div>
  )
}

function MoneyInput({
  id,
  value,
  placeholder,
  onChange,
}: {
  id: string
  value: string
  placeholder: string
  onChange: (value: string) => void
}) {
  return (
    <label className="money-input" htmlFor={id}>
      <span>$</span>
      <input
        id={id}
        type="number"
        min="0"
        inputMode="decimal"
        placeholder={placeholder}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
    </label>
  )
}

function FieldError({ message }: { message: string | null }) {
  if (!message) {
    return null
  }
  return <p className="field-error">{message}</p>
}

function LoadingScreen() {
  return (
    <section className="screen loading-screen">
      <div className="scan-ring">
        <span>SCANNING ACTIVE OFFERS</span>
      </div>
      <p className="helper-copy centered-copy">Checking value, spend achievability, and eligibility confidence.</p>
    </section>
  )
}

function ErrorScreen({ message, onRetry, onReview }: { message: string; onRetry: () => void; onReview: () => void }) {
  return (
    <section className="screen question-screen">
      <p className="eyebrow error-eyebrow">Offer check failed</p>
      <h1>We could not check offers right now.</h1>
      <div className="error-card">{message}</div>
      <button className="primary-button" type="button" onClick={onRetry}>
        Try again
      </button>
      <button className="secondary-button" type="button" onClick={onReview}>
        Review answers
      </button>
    </section>
  )
}

function RoadmapView({ roadmap, onStartOver }: { roadmap: RecommendationRoadmap; onStartOver: () => void }) {
  if (!roadmap.hasRecommendation || !roadmap.bestRecommendation) {
    return (
      <section className="screen results-stack">
        <p className="eyebrow">No safe card found yet</p>
        <h1>No card is safe enough to recommend.</h1>
        <ListItems items={roadmap.noRecommendationReasons} />
        <button className="primary-button" type="button" onClick={onStartOver}>
          Review answers
        </button>
      </section>
    )
  }

  const best = roadmap.bestRecommendation

  return (
    <section className="screen results-stack">
      <BestRecommendationCard candidate={best} cardsConsidered={roadmap.summary.cardsConsidered} />

      <section className="section-block next-steps-block">
        <div className="section-heading">
          <h2>Your next steps</h2>
          <span className="pill">ACTION PLAN</span>
        </div>
        <div className="checklist">
          {(roadmap.actionChecklist ?? []).map((item) => (
            <article className={`checklist-item ${item.kind === 'meet_spend' ? 'checklist-item-primary' : ''}`} key={`${item.kind}-${item.title}`}>
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

      <section className="section-block">
        <div className="section-heading">
          <h2>Why this card</h2>
          <span className="quiet-meta">Rank #{best.rank || 1}</span>
        </div>
        <ListItems items={roadmap.reasons} />
      </section>

      {(roadmap.warnings ?? []).length > 0 ? (
        <section className="section-block warning-block">
          <h2>Review before applying</h2>
          <ListItems items={roadmap.warnings} tone="warning" />
        </section>
      ) : null}

      <CandidateList title="Alternatives" candidates={roadmap.alternatives ?? []} />
      <CandidateList
        title="Caution cards"
        intro="Not selected because of eligibility, spend, or manual-review cautions."
        candidates={(roadmap.ineligibleOrCautionCards ?? []).slice(0, 3)}
        caution
      />

      <p className="disclaimer">
        This tool provides estimates based on curated offer data and simplified assumptions. It is not financial advice.
        Always check issuer terms and consider your personal circumstances before applying.
      </p>

      <button className="secondary-button" type="button" onClick={onStartOver}>
        Start again
      </button>
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
        <CardVisual issuer={offer.issuer} rewardType={offer.rewardType} size="large" />
        <div>
          <h2>{offer.cardName}</h2>
          <p className="issuer-name">{offer.issuer}</p>
        </div>
        <p className="value-statement">
          Estimated year-one value of {formatCents(valueBreakdown.netEstimatedValueCents)}
        </p>
        <div className="metadata-grid">
          <span>{offer.signupBonusPoints.toLocaleString('en-AU')} points</span>
          <span>{formatCents(spendRequirement.minimumSpendCents)} spend</span>
          <span>{formatCents(offer.annualFeeCents)} fee</span>
          <span>{spendRequirement.difficulty}</span>
        </div>
      </article>
    </section>
  )
}

function CandidateList({
  title,
  intro,
  candidates,
  caution = false,
}: {
  title: string
  intro?: string
  candidates: RecommendationCandidate[]
  caution?: boolean
}) {
  if (candidates.length === 0) {
    return null
  }

  return (
    <section className="section-block">
      <h2>{title}</h2>
      {intro ? <p className="helper-copy small-copy">{intro}</p> : null}
      <div className="compact-list">
        {candidates.map((candidate) => (
          <article className="offer-card compact-card" key={`${candidate.offer.issuer}-${candidate.offer.cardName}`}>
            <span className={`pill ${caution ? 'pill-hot' : ''}`}>
              {caution ? 'CAUTION' : `RANK ${candidate.rank}`}
            </span>
            <div className="compact-card-header">
              <CardVisual issuer={candidate.offer.issuer} rewardType={candidate.offer.rewardType} size="small" />
              <div>
                <h3>{candidate.offer.cardName}</h3>
                <p>{candidate.offer.issuer}</p>
              </div>
            </div>
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

function CardVisual({ issuer, rewardType, size }: { issuer: string; rewardType: string; size: 'large' | 'small' }) {
  const initials = issuer
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0]?.toUpperCase())
    .join('')

  return (
    <div className={`card-visual card-visual-${size} ${cardVisualClass(issuer, rewardType)}`} aria-hidden="true">
      <div className="card-chip" />
      <span>{initials || 'CC'}</span>
      <div className="card-orbit" />
    </div>
  )
}

function cardVisualClass(issuer: string, rewardType: string): string {
  if (issuer.toLowerCase().includes('american express')) {
    return 'card-visual-amex'
  }
  switch (rewardType) {
    case 'qantas_points':
      return 'card-visual-qantas'
    case 'velocity_points':
      return 'card-visual-velocity'
    case 'bank_points':
      return 'card-visual-bank'
    case 'flexible_points':
      return 'card-visual-flexible'
    default:
      return 'card-visual-default'
  }
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

function Progress({
  step,
  form,
  roadmap,
  isLoading,
}: {
  step: WizardStep
  form: FormState
  roadmap: RecommendationRoadmap | null
  isLoading: boolean
}) {
  if (roadmap) {
    return <span className="pill pill-warm">RESULT</span>
  }
  if (isLoading) {
    return <span className="progress-ring">90%</span>
  }
  if (step === 'intro') {
    return <span className="pill pill-warm">POINTS ENGINE</span>
  }
  return <span className="progress-ring">{progressPercent(step, form)}%</span>
}

function progressPercent(step: WizardStep, form: FormState): number {
  const steps = activeSteps(form)
  const index = steps.indexOf(step)
  if (index < 0) {
    return 0
  }
  return Math.round((index / (steps.length - 1)) * 100)
}

function activeSteps(form: FormState): WizardStep[] {
  const steps: WizardStep[] = ['goal', 'monthlySpend', 'largePurchases', 'annualFee']
  if (form.annualFeePreference === 'strict_max') {
    steps.push('maxAnnualFee')
  }
  steps.push('amex', 'review')
  return steps
}

function previousStep(step: WizardStep, form: FormState): WizardStep {
  if (step === 'intro') {
    return 'intro'
  }
  const steps = activeSteps(form)
  const index = steps.indexOf(step)
  if (index <= 0) {
    return 'intro'
  }
  return steps[index - 1]
}

function ReviewRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="review-row">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

function labelForGoal(goal?: OptimisationGoal): string {
  return goalOptions.find((option) => option.value === goal)?.label ?? 'Not selected'
}

function labelForFee(preference?: AnnualFeePreference): string {
  return feeOptions.find((option) => option.value === preference)?.label ?? 'Not selected'
}

function formatDollars(value: string): string {
  return formatCents(dollarsToCents(value))
}

export default App
