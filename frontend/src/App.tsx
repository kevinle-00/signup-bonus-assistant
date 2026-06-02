import { useEffect, useState } from 'react'
import type { Dispatch, ReactNode, SetStateAction } from 'react'
import './App.css'
import { createRecommendation } from './api'
import { dollarsToCents, formatCents, formatDate, formatRewardType } from './format'
import type {
  AnnualFeePreference,
  OptimisationGoal,
  RecommendationCandidate,
  RecommendationInput,
  RecommendationRoadmap,
  UserCardHistoryItem,
} from './types'

type AppView = 'wizard' | 'profile' | 'cardHistory' | 'issuerPicker'
type CardHistoryReturnView = 'wizard' | 'profile'
type ResultDetailView = 'overview' | 'timeline' | 'checklist' | 'historyImpact' | 'warnings' | 'alternatives' | 'caution'

type WizardStep =
  | 'intro'
  | 'goal'
  | 'monthlySpend'
  | 'largePurchases'
  | 'annualFee'
  | 'maxAnnualFee'
  | 'amex'
  | 'cardHistory'
  | 'review'

type FormState = {
  optimisationGoal?: OptimisationGoal
  monthlySpendCents?: number
  monthlySpendLabel?: string
  largePurchasesCents?: number
  largePurchasesLabel?: string
  annualFeePreference?: AnnualFeePreference
  maxAnnualFeeDollars: string
  acceptsAmex?: boolean
}

type RangeOption = {
  value: number
  label: string
  description: string
}

type CardHistoryDraft = {
  issuer: string
  cardName: string
  status: CardHistoryStatus
  closedTiming?: ClosedTiming
}

type CardHistoryStatus = 'current' | 'closed'

type ClosedTiming = 'last_6' | 'six_to_twelve' | 'twelve_to_eighteen' | 'eighteen_to_twenty_four' | 'more_than_twenty_four'

const initialForm: FormState = {
  maxAnnualFeeDollars: '',
}

const initialCardDraft: CardHistoryDraft = {
  issuer: '',
  cardName: '',
  status: 'current',
}

const cardHistoryStorageKey = 'signupBonusAssistant.cardHistory'

const goalOptions: Array<{ value: OptimisationGoal; label: string; description: string }> = [
  { value: 'qantas_points', label: 'Qantas Points', description: 'Prioritise Qantas sign-up bonuses.' },
  { value: 'velocity_points', label: 'Velocity Points', description: 'Prioritise Velocity sign-up bonuses.' },
  { value: 'max_net_value', label: 'Maximum net value', description: 'Rank by estimated value after fees.' },
  { value: 'cashback', label: 'Cashback', description: 'Prefer offers with direct cash value.' },
  { value: 'low_effort', label: 'Low effort', description: 'Favour easier spend and cleaner eligibility.' },
]

const monthlySpendOptions: RangeOption[] = [
  { value: 75_000, label: '$0-$1,000', description: 'Use $750/month in the estimate.' },
  { value: 150_000, label: '$1,000-$2,000', description: 'Use $1,500/month in the estimate.' },
  { value: 300_000, label: '$2,000-$4,000', description: 'Use $3,000/month in the estimate.' },
  { value: 500_000, label: '$4,000-$6,000', description: 'Use $5,000/month in the estimate.' },
  { value: 800_000, label: '$6,000-$10,000', description: 'Use $8,000/month in the estimate.' },
  { value: 1_000_000, label: '$10,000+', description: 'Use $10,000/month in the estimate.' },
]

const largePurchaseOptions: RangeOption[] = [
  { value: 0, label: '$0', description: 'No planned large purchases.' },
  { value: 75_000, label: '$1-$1,000', description: 'Use $750 in the estimate.' },
  { value: 200_000, label: '$1,000-$3,000', description: 'Use $2,000 in the estimate.' },
  { value: 400_000, label: '$3,000-$5,000', description: 'Use $4,000 in the estimate.' },
  { value: 500_000, label: '$5,000+', description: 'Use $5,000 in the estimate.' },
]

const feeOptions: Array<{ value: AnnualFeePreference; label: string; description: string }> = [
  { value: 'flexible', label: 'Flexible if the value is strong', description: 'Let net value do most of the work.' },
  { value: 'prefer_low', label: 'Prefer lower fees', description: 'Penalise high-fee cards without hiding them.' },
  { value: 'strict_max', label: 'Set a strict maximum', description: 'Exclude cards above your fee limit.' },
]

const issuerOptions = [
  'American Express',
  'ANZ',
  'Bank of Melbourne',
  'BankSA',
  'Bankwest',
  'Bendigo Bank',
  'Citi',
  'Commonwealth Bank',
  'HSBC',
  'Macquarie Bank',
  'NAB',
  'Qantas Money',
  'St.George',
  'Suncorp Bank',
  'Virgin Money',
  'Westpac',
]

const cardStatusOptions: Array<{ value: CardHistoryStatus; label: string; description: string }> = [
  { value: 'current', label: 'Currently held', description: 'You still have this card open.' },
  { value: 'closed', label: 'Recently closed', description: 'You closed this card in the past few years.' },
]

const closedTimingOptions: Array<{ value: ClosedTiming; label: string; description: string; monthsAgo: number }> = [
  { value: 'last_6', label: 'Last 6 months', description: 'Most issuer exclusion windows will still apply.', monthsAgo: 3 },
  { value: 'six_to_twelve', label: '6-12 months ago', description: 'Useful for 12, 18, and 24 month rules.', monthsAgo: 9 },
  { value: 'twelve_to_eighteen', label: '12-18 months ago', description: 'Useful for 18 and 24 month rules.', monthsAgo: 15 },
  { value: 'eighteen_to_twenty_four', label: '18-24 months ago', description: 'Useful for 24 month rules.', monthsAgo: 21 },
  { value: 'more_than_twenty_four', label: 'More than 24 months', description: 'Usually outside common bonus exclusion windows.', monthsAgo: 30 },
]

function App() {
  const [view, setView] = useState<AppView>('wizard')
  const [cardHistoryReturnView, setCardHistoryReturnView] = useState<CardHistoryReturnView>('profile')
  const [form, setForm] = useState<FormState>(initialForm)
  const [cardHistory, setCardHistory] = useState<UserCardHistoryItem[]>(readStoredCardHistory)
  const [cardDraft, setCardDraft] = useState<CardHistoryDraft>(initialCardDraft)
  const [cardDraftError, setCardDraftError] = useState<string | null>(null)
  const [step, setStep] = useState<WizardStep>('intro')
  const [roadmap, setRoadmap] = useState<RecommendationRoadmap | null>(null)
  const [resultDetail, setResultDetail] = useState<ResultDetailView>('overview')
  const [error, setError] = useState<string | null>(null)
  const [fieldError, setFieldError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    writeStoredCardHistory(cardHistory)
  }, [cardHistory])

  function moveTo(nextStep: WizardStep) {
    setFieldError(null)
    setStep(nextStep)
  }

  function goBack() {
    setFieldError(null)
    setStep(previousStep(step, form))
  }

  function handleTopbarBack() {
    if (view === 'wizard' && roadmap && resultDetail !== 'overview') {
      setResultDetail('overview')
      return
    }
    if (view === 'issuerPicker') {
      setView('cardHistory')
      return
    }
    if (view === 'cardHistory') {
      setView(cardHistoryReturnView)
      return
    }
    if (view === 'profile') {
      setView('wizard')
      return
    }
    goBack()
  }

  function startOver() {
    setForm(initialForm)
    setRoadmap(null)
    setResultDetail('overview')
    setError(null)
    setFieldError(null)
    setIsLoading(false)
    setStep('intro')
    setView('wizard')
    setCardHistoryReturnView('profile')
    setCardDraft(initialCardDraft)
    setCardDraftError(null)
  }

  function openCardHistory(returnView: CardHistoryReturnView) {
    setCardHistoryReturnView(returnView)
    setView('cardHistory')
  }

  function selectIssuer(issuer: string) {
    setCardDraft((current) => ({ ...current, issuer }))
    setCardDraftError(null)
    setView('cardHistory')
  }

  function continueFromMaxFee() {
    const cents = dollarsToCents(form.maxAnnualFeeDollars)
    if (cents <= 0) {
      setFieldError('Enter a maximum annual fee greater than $0.')
      return
    }
    moveTo('amex')
  }

  async function submitRecommendation() {
    if (!form.optimisationGoal || !form.annualFeePreference || form.acceptsAmex === undefined || form.monthlySpendCents === undefined || form.largePurchasesCents === undefined) {
      setFieldError('Review your answers before continuing.')
      return
    }

    setError(null)
    setFieldError(null)
    setIsLoading(true)

    const input: RecommendationInput = {
      optimisationGoal: form.optimisationGoal,
      monthlySpendCents: form.monthlySpendCents,
      expectedLargePurchasesNext90DaysCents: form.largePurchasesCents,
      annualFeePreference: form.annualFeePreference,
      maxAnnualFeeCents: dollarsToCents(form.maxAnnualFeeDollars),
      acceptsAmex: form.acceptsAmex,
      cardHistory,
    }

    try {
      const nextRoadmap = await createRecommendation(input)
      setRoadmap(nextRoadmap)
      setResultDetail('overview')
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Could not create recommendation.')
    } finally {
      setIsLoading(false)
    }
  }

  const showBack = view !== 'wizard' || resultDetail !== 'overview' || (step !== 'intro' && !roadmap && !isLoading)

  return (
    <main className="app-shell">
      <header className="topbar">
        {showBack ? (
          <button className="icon-button" type="button" onClick={handleTopbarBack} aria-label="Go back">
            ‹
          </button>
        ) : (
          <button className="brand-mark" type="button" onClick={() => setView('profile')} aria-label="Open profile">
            ME
          </button>
        )}
        <Progress step={step} form={form} roadmap={roadmap} isLoading={isLoading} view={view} />
      </header>

      {view === 'profile' ? (
        <ProfileScreen cardHistoryCount={cardHistory.length} roadmap={roadmap} onOpenCardHistory={() => openCardHistory('profile')} />
      ) : null}

      {view === 'cardHistory' ? (
        <CardHistoryScreen
          cardHistory={cardHistory}
          draft={cardDraft}
          draftError={cardDraftError}
          setCardHistory={setCardHistory}
          setDraft={setCardDraft}
          setDraftError={setCardDraftError}
          openIssuerPicker={() => setView('issuerPicker')}
        />
      ) : null}

      {view === 'issuerPicker' ? (
        <IssuerPickerScreen selectedIssuer={cardDraft.issuer} onSelect={selectIssuer} />
      ) : null}

      {view === 'wizard' && isLoading ? <LoadingScreen /> : null}
      {view === 'wizard' && !isLoading && error ? <ErrorScreen message={error} onRetry={submitRecommendation} onReview={() => moveTo('review')} /> : null}
      {view === 'wizard' && !isLoading && !error && roadmap ? (
        <RoadmapView
          roadmap={roadmap}
          resultDetail={resultDetail}
          setResultDetail={setResultDetail}
          onStartOver={startOver}
          onOpenProfile={() => setView('profile')}
        />
      ) : null}
      {view === 'wizard' && !isLoading && !error && !roadmap ? (
        <WizardScreen
          form={form}
          step={step}
          fieldError={fieldError}
          cardHistory={cardHistory}
          setForm={setForm}
          moveTo={moveTo}
          continueFromMaxFee={continueFromMaxFee}
          submitRecommendation={submitRecommendation}
          openCardHistory={() => openCardHistory('wizard')}
        />
      ) : null}
    </main>
  )
}

function WizardScreen({
  form,
  step,
  fieldError,
  cardHistory,
  setForm,
  moveTo,
  continueFromMaxFee,
  submitRecommendation,
  openCardHistory,
}: {
  form: FormState
  step: WizardStep
  fieldError: string | null
  cardHistory: UserCardHistoryItem[]
  setForm: Dispatch<SetStateAction<FormState>>
  moveTo: (step: WizardStep) => void
  continueFromMaxFee: () => void
  submitRecommendation: () => void
  openCardHistory: () => void
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
        helper="A range is enough. We use a conservative representative value in the estimate."
      >
        <OptionList
          options={monthlySpendOptions}
          selected={form.monthlySpendCents}
          onSelect={(value, option) => {
            setForm((current) => ({ ...current, monthlySpendCents: value, monthlySpendLabel: option.label }))
            moveTo('largePurchases')
          }}
        />
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
        <OptionList
          options={largePurchaseOptions}
          selected={form.largePurchasesCents}
          onSelect={(value, option) => {
            setForm((current) => ({ ...current, largePurchasesCents: value, largePurchasesLabel: option.label }))
            moveTo('annualFee')
          }}
        />
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
        <button className="primary-button bottom-action" type="button" onClick={continueFromMaxFee}>
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
            moveTo('cardHistory')
          }}
        />
      </QuestionScreen>
    )
  }

  if (step === 'cardHistory') {
    return (
      <QuestionScreen
        eyebrow="Eligibility check"
        title="Card history helps avoid offers you may not qualify for."
        helper="Add cards you currently hold or recently closed. You can skip this if you are unsure."
      >
        <div className="review-card">
          <ReviewRow label="Cards saved" value={cardHistory.length === 0 ? 'None yet' : `${cardHistory.length} card${cardHistory.length === 1 ? '' : 's'}`} />
        </div>
        <button className="secondary-button" type="button" onClick={openCardHistory}>
          Manage card history
        </button>
        <button className="primary-button bottom-action" type="button" onClick={() => moveTo('review')}>
          Continue
        </button>
      </QuestionScreen>
    )
  }

  return (
    <QuestionScreen eyebrow="Review" title="Ready to check the active offers?">
      <div className="review-card">
        <ReviewRow label="Goal" value={labelForGoal(form.optimisationGoal)} />
        <ReviewRow label="Monthly spend" value={form.monthlySpendLabel ?? 'Not selected'} />
        <ReviewRow label="Large purchases" value={form.largePurchasesLabel ?? 'Not selected'} />
        <ReviewRow label="Annual fee" value={labelForFee(form.annualFeePreference)} />
        {form.annualFeePreference === 'strict_max' ? (
          <ReviewRow label="Fee limit" value={formatDollars(form.maxAnnualFeeDollars)} />
        ) : null}
        <ReviewRow label="Amex" value={form.acceptsAmex ? 'Included' : 'Excluded'} />
        <ReviewRow label="Card history" value={cardHistory.length === 0 ? 'None added' : `${cardHistory.length} saved`} />
      </div>
      <FieldError message={fieldError} />
      <button className="primary-button bottom-action" type="button" onClick={submitRecommendation}>
        Find my best offer
      </button>
    </QuestionScreen>
  )
}

function ProfileScreen({
  cardHistoryCount,
  roadmap,
  onOpenCardHistory,
}: {
  cardHistoryCount: number
  roadmap: RecommendationRoadmap | null
  onOpenCardHistory: () => void
}) {
  return (
    <section className="screen profile-screen">
      <div>
        <p className="eyebrow">Profile</p>
        <h1>Your assistant profile</h1>
        <p className="hero-copy">Review the details used to personalise your card-switching recommendation.</p>
      </div>
      <div className="profile-grid">
        <button className="profile-tile" type="button" onClick={onOpenCardHistory}>
          <span>Card history</span>
          <small>{cardHistoryCount === 0 ? 'No cards added' : `${cardHistoryCount} saved`}</small>
          <b>→</b>
        </button>
        <div className="profile-tile profile-tile-muted">
          <span>Preferences</span>
          <small>Set during the wizard</small>
          <b>·</b>
        </div>
        <div className="profile-tile profile-tile-muted">
          <span>Latest recommendation</span>
          <small>{roadmap?.bestRecommendation?.offer.cardName ?? 'No result yet'}</small>
          <b>·</b>
        </div>
        <div className="profile-tile profile-tile-muted">
          <span>About this estimate</span>
          <small>Curated offers, simplified assumptions, not financial advice</small>
          <b>·</b>
        </div>
      </div>
    </section>
  )
}

function CardHistoryScreen({
  cardHistory,
  draft,
  draftError,
  setCardHistory,
  setDraft,
  setDraftError,
  openIssuerPicker,
}: {
  cardHistory: UserCardHistoryItem[]
  draft: CardHistoryDraft
  draftError: string | null
  setCardHistory: Dispatch<SetStateAction<UserCardHistoryItem[]>>
  setDraft: Dispatch<SetStateAction<CardHistoryDraft>>
  setDraftError: Dispatch<SetStateAction<string | null>>
  openIssuerPicker: () => void
}) {
  function addCard() {
    if (!draft.issuer.trim() || !draft.cardName.trim()) {
      setDraftError('Add both issuer and card name.')
      return
    }
    if (draft.status === 'closed' && !draft.closedTiming) {
      setDraftError('Choose roughly when you closed this card.')
      return
    }
    const closedAt = draft.status === 'closed' && draft.closedTiming ? closedAtFromTiming(draft.closedTiming) : undefined
    setCardHistory((current) => [
      ...current,
      {
        issuer: draft.issuer.trim(),
        cardName: draft.cardName.trim(),
        currentlyHeld: draft.status === 'current',
        closedAt,
      },
    ])
    setDraft(initialCardDraft)
    setDraftError(null)
  }

  return (
    <section className="screen card-history-screen">
      <div>
        <p className="eyebrow">Card history</p>
        <h1>Cards you hold or recently closed</h1>
        <p className="hero-copy">This self-reported history helps the assistant spot bonus exclusions and lower-confidence offers.</p>
      </div>

      <div className="history-list">
        {cardHistory.length === 0 ? <p className="helper-copy small-copy">No cards added yet.</p> : null}
        {cardHistory.map((item, index) => (
          <article className="history-card" key={`${item.issuer}-${item.cardName}-${index}`}>
            <div>
              <h3>{item.cardName}</h3>
              <p>{item.issuer}</p>
              <span>{item.currentlyHeld ? 'Currently held' : `Closed ${item.closedAt ? formatDate(item.closedAt) : 'recently'}`}</span>
            </div>
            <button
              className="text-button"
              type="button"
              onClick={() => setCardHistory((current) => current.filter((_, itemIndex) => itemIndex !== index))}
            >
              Remove
            </button>
          </article>
        ))}
      </div>

      <div className="history-form">
        <p className="small-copy">Choose from known issuers so saved history can match the current bonus rules more reliably.</p>
        <button
          className={`picker-button ${draft.issuer ? 'picker-button-selected' : ''}`}
          type="button"
          onClick={openIssuerPicker}
        >
          <span>Issuer</span>
          <strong>{draft.issuer || 'Choose issuer'}</strong>
          <b>→</b>
        </button>
        <input
          className="text-input"
          aria-label="Card name"
          placeholder="Card name, e.g. Rewards Platinum"
          value={draft.cardName}
          onChange={(event) => setDraft((current) => ({ ...current, cardName: event.target.value }))}
        />
        <InlineChoiceGroup
          label="Card status"
          options={cardStatusOptions}
          selected={draft.status}
          onSelect={(status) => setDraft((current) => ({ ...current, status, closedTiming: status === 'current' ? undefined : current.closedTiming }))}
        />
        {draft.status === 'closed' ? (
          <InlineChoiceGroup
            label="When did you close it?"
            helper="Rough timing is enough. We only use this to flag possible bonus exclusions."
            options={closedTimingOptions}
            selected={draft.closedTiming}
            onSelect={(closedTiming) => setDraft((current) => ({ ...current, closedTiming }))}
          />
        ) : null}
        <FieldError message={draftError} />
        <button className="primary-button" type="button" onClick={addCard}>
          Add card
        </button>
      </div>
    </section>
  )
}

function InlineChoiceGroup<T extends string>({
  label,
  helper,
  options,
  selected,
  onSelect,
}: {
  label: string
  helper?: string
  options: Array<{ value: T; label: string; description: string }>
  selected?: T
  onSelect: (value: T) => void
}) {
  return (
    <div className="choice-group">
      <div>
        <p className="field-label-text">{label}</p>
        {helper ? <p className="small-copy choice-helper">{helper}</p> : null}
      </div>
      <div className="choice-card-list">
        {options.map((option) => (
          <button
            className={`choice-card ${selected === option.value ? 'choice-card-selected' : ''}`}
            type="button"
            key={option.value}
            onClick={() => onSelect(option.value)}
          >
            <span>{option.label}</span>
            <small>{option.description}</small>
          </button>
        ))}
      </div>
    </div>
  )
}

function IssuerPickerScreen({
  selectedIssuer,
  onSelect,
}: {
  selectedIssuer: string
  onSelect: (issuer: string) => void
}) {
  return (
    <section className="screen issuer-picker-screen">
      <div>
        <p className="eyebrow">Known issuers</p>
        <h1>Choose issuer</h1>
        <p className="helper-copy">Pick the issuer shown on your card. This helps match bonus exclusion rules more reliably.</p>
      </div>
      <div className="picker-list">
        {issuerOptions.map((issuer) => (
          <button
            className={`picker-row ${selectedIssuer === issuer ? 'picker-row-selected' : ''}`}
            type="button"
            key={issuer}
            onClick={() => onSelect(issuer)}
          >
            <span>{issuer}</span>
            <b>{selectedIssuer === issuer ? 'Selected' : 'Choose'}</b>
          </button>
        ))}
      </div>
    </section>
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
  children: ReactNode
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

function OptionList<T extends string | boolean | number>({
  options,
  selected,
  onSelect,
}: {
  options: Array<{ value: T; label: string; description: string }>
  selected?: T
  onSelect: (value: T, option: { value: T; label: string; description: string }) => void
}) {
  return (
    <div className="option-list">
      {options.map((option) => (
        <button
          className={`option-card ${selected === option.value ? 'option-card-selected' : ''}`}
          key={String(option.value)}
          type="button"
          onClick={() => onSelect(option.value, option)}
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

function RoadmapView({
  roadmap,
  resultDetail,
  setResultDetail,
  onStartOver,
  onOpenProfile,
}: {
  roadmap: RecommendationRoadmap
  resultDetail: ResultDetailView
  setResultDetail: (view: ResultDetailView) => void
  onStartOver: () => void
  onOpenProfile: () => void
}) {
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
  const historyImpacts = cardHistoryImpacts(roadmap)
  const warnings = roadmap.warnings ?? []
  const alternatives = roadmap.alternatives ?? []
  const cautionCards = (roadmap.ineligibleOrCautionCards ?? []).slice(0, 3)

  if (resultDetail !== 'overview') {
    return (
      <ResultDetailScreen
        view={resultDetail}
        roadmap={roadmap}
        best={best}
        historyImpacts={historyImpacts}
        warnings={warnings}
        alternatives={alternatives}
        cautionCards={cautionCards}
      />
    )
  }

  return (
    <section className="screen results-stack">
      <BestRecommendationCard candidate={best} cardsConsidered={roadmap.summary.cardsConsidered} />

      <section className="section-block why-summary-block">
        <div className="section-heading">
          <h2>Why this card</h2>
          <span className="quiet-meta">Rank #{best.rank || 1}</span>
        </div>
        <ListItems items={firstItems(roadmap.reasons, 3)} />
      </section>

      <section className="result-nav-list">
        <ResultNavCard
          title="12-month switch plan"
          description="Apply, meet spend, track the bonus, review renewal, and rerun the assistant."
          meta="Roadmap"
          onClick={() => setResultDetail('timeline')}
        />
        <ResultNavCard
          title="Switching checklist"
          description="The concrete actions to take before and after applying."
          meta={`${(roadmap.actionChecklist ?? []).length} actions`}
          onClick={() => setResultDetail('checklist')}
        />
        {historyImpacts.length > 0 ? (
          <ResultNavCard
            title="Card history impact"
            description="See which saved history entries affected eligibility confidence."
            meta={`${historyImpacts.length} note${historyImpacts.length === 1 ? '' : 's'}`}
            onClick={() => setResultDetail('historyImpact')}
          />
        ) : null}
        {warnings.length > 0 ? (
          <ResultNavCard
            title="Review before applying"
            description="Read eligibility, spend, or offer cautions before you apply."
            meta={`${warnings.length} warning${warnings.length === 1 ? '' : 's'}`}
            onClick={() => setResultDetail('warnings')}
          />
        ) : null}
        {alternatives.length > 0 ? (
          <ResultNavCard
            title="Alternative offers"
            description="Backup options if the top card no longer fits after checking terms."
            meta={`${alternatives.length} cards`}
            onClick={() => setResultDetail('alternatives')}
          />
        ) : null}
        {cautionCards.length > 0 ? (
          <ResultNavCard
            title="Caution cards"
            description="Cards held back because of eligibility, spend, or manual-review cautions."
            meta={`${cautionCards.length} cards`}
            onClick={() => setResultDetail('caution')}
          />
        ) : null}
      </section>

      <p className="disclaimer">
        This tool provides estimates based on curated offer data and simplified assumptions. It is not financial advice.
        Always check issuer terms and consider your personal circumstances before applying.
      </p>

      <button className="secondary-button" type="button" onClick={onOpenProfile}>
        View profile
      </button>
      <button className="secondary-button" type="button" onClick={onStartOver}>
        Start again
      </button>
    </section>
  )
}

function ResultDetailScreen({
  view,
  roadmap,
  best,
  historyImpacts,
  warnings,
  alternatives,
  cautionCards,
}: {
  view: ResultDetailView
  roadmap: RecommendationRoadmap
  best: RecommendationCandidate
  historyImpacts: CardHistoryImpactItem[]
  warnings: string[]
  alternatives: RecommendationCandidate[]
  cautionCards: RecommendationCandidate[]
}) {
  if (view === 'timeline') {
    return (
      <section className="screen results-stack">
        <SwitchingTimeline candidate={best} />
      </section>
    )
  }

  if (view === 'checklist') {
    return (
      <section className="screen results-stack">
        <SwitchingChecklist items={roadmap.actionChecklist ?? []} />
      </section>
    )
  }

  if (view === 'historyImpact') {
    return (
      <section className="screen results-stack">
        <CardHistoryImpact impacts={historyImpacts} />
      </section>
    )
  }

  if (view === 'warnings') {
    return (
      <section className="screen results-stack">
        <section className="section-block warning-block">
          <h1>Review before applying</h1>
          <ListItems items={warnings} tone="warning" />
        </section>
      </section>
    )
  }

  if (view === 'alternatives') {
    return (
      <section className="screen results-stack">
        <CandidateList title="Alternative offers" intro="Keep these lower priority unless the top card no longer fits after checking issuer terms." candidates={alternatives} />
      </section>
    )
  }

  return (
    <section className="screen results-stack">
      <CandidateList
        title="Caution cards"
        intro="Not selected because of eligibility, spend, or manual-review cautions."
        candidates={cautionCards}
        caution
      />
    </section>
  )
}

function ResultNavCard({ title, description, meta, onClick }: { title: string; description: string; meta: string; onClick: () => void }) {
  return (
    <button className="result-nav-card" type="button" onClick={onClick}>
      <div>
        <span>{meta}</span>
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
      <b>→</b>
    </button>
  )
}

function SwitchingChecklist({ items }: { items: NonNullable<RecommendationRoadmap['actionChecklist']> }) {
  return (
    <section className="section-block next-steps-block">
      <div className="section-heading">
        <h1>Switching checklist</h1>
        <span className="pill">NEXT ACTIONS</span>
      </div>
      <p className="small-copy">Use this as a one-card plan. The MVP is not sequencing multiple churn moves yet.</p>
      <div className="checklist">
        {items.map((item) => (
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
  )
}

function SwitchingTimeline({ candidate }: { candidate: RecommendationCandidate }) {
  const { offer } = candidate
  const spendWindowMonths = Math.max(1, Math.ceil(offer.spendWindowDays / 30))
  const spendWindowLabel = spendWindowMonths === 1 ? 'Month 1' : `Months 1-${spendWindowMonths}`

  const items = [
    {
      marker: 'Today',
      title: 'Verify terms and apply',
      description: `Check the live ${offer.issuer} offer, eligibility rules, annual fee, and exclusions before applying.`,
    },
    {
      marker: spendWindowLabel,
      title: `Meet the ${formatCents(offer.minimumSpendCents)} spend requirement`,
      description: `Use eligible purchases only. The model projects this as ${candidate.spendRequirement.difficulty}.`,
    },
    {
      marker: `Month ${spendWindowMonths + 1}`,
      title: 'Confirm the bonus posted',
      description: 'Keep spend evidence and check that points or cashback land after the issuer processing period.',
    },
    {
      marker: 'Month 11',
      title: 'Review before renewal',
      description: `Decide whether the card is still worth keeping before another ${formatCents(offer.annualFeeCents)} annual fee.`,
    },
    {
      marker: 'Month 12',
      title: 'Run the assistant again',
      description: 'Look for the next switching opportunity once your card history and active offers have changed.',
    },
  ]

  return (
    <section className="section-block timeline-block">
      <div className="section-heading">
        <h2>Your 12-month switch plan</h2>
        <span className="pill">ROADMAP</span>
      </div>
      <div className="timeline-list">
        {items.map((item, index) => (
          <article className="timeline-item" key={item.marker}>
            <span className="timeline-index">{String(index + 1).padStart(2, '0')}</span>
            <div>
              <span className="timeline-marker">{item.marker}</span>
              <h3>{item.title}</h3>
              <p>{item.description}</p>
            </div>
          </article>
        ))}
      </div>
    </section>
  )
}

function CardHistoryImpact({ impacts }: { impacts: CardHistoryImpactItem[] }) {
  if (impacts.length === 0) {
    return null
  }

  return (
    <section className="section-block history-impact-block">
      <div className="section-heading">
        <h2>Card history impact</h2>
        <span className="pill">SELF-REPORTED</span>
      </div>
      <div className="history-impact-list">
        {impacts.map((impact) => (
          <article className="history-impact-card" key={`${impact.issuer}-${impact.cardName}-${impact.warning}`}>
            <h3>{impact.cardName}</h3>
            <p>{impact.issuer}</p>
            <span>{impact.warning}</span>
          </article>
        ))}
      </div>
    </section>
  )
}

type CardHistoryImpactItem = {
  issuer: string
  cardName: string
  warning: string
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
        <span className="pill pill-hot">SWITCH TO THIS CARD NEXT</span>
        <CardVisual issuer={offer.issuer} rewardType={offer.rewardType} size="large" />
        <div>
          <h2>{offer.cardName}</h2>
          <p className="issuer-name">{offer.issuer}</p>
        </div>
        <p className="value-statement">
          Estimated year-one value of {formatCents(valueBreakdown.netEstimatedValueCents)}
        </p>
        <p className="switch-summary">This is the strongest immediate switch based on your spend, card history, and active bonus offers.</p>
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

function firstItems(items: string[] | null, limit: number): string[] {
  return (items?.filter(Boolean) ?? []).slice(0, limit)
}

function cardHistoryImpacts(roadmap: RecommendationRoadmap): CardHistoryImpactItem[] {
  const impacts: CardHistoryImpactItem[] = []
  const candidates = [roadmap.bestRecommendation, ...(roadmap.alternatives ?? []), ...(roadmap.ineligibleOrCautionCards ?? [])]

  for (const candidate of candidates) {
    if (!candidate) {
      continue
    }
    for (const warning of candidate.warnings ?? []) {
      if (!isCardHistoryWarning(warning)) {
        continue
      }
      impacts.push({ issuer: candidate.offer.issuer, cardName: candidate.offer.cardName, warning })
    }
  }

  return impacts.slice(0, 3)
}

function isCardHistoryWarning(value: string): boolean {
  return value.toLowerCase().includes('card history')
}

function Progress({
  step,
  form,
  roadmap,
  isLoading,
  view,
}: {
  step: WizardStep
  form: FormState
  roadmap: RecommendationRoadmap | null
  isLoading: boolean
  view: AppView
}) {
  if (view === 'profile') {
    return <span className="pill pill-warm">PROFILE</span>
  }
  if (view === 'cardHistory') {
    return <span className="pill pill-warm">HISTORY</span>
  }
  if (view === 'issuerPicker') {
    return <span className="pill pill-warm">ISSUER</span>
  }
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
  steps.push('amex', 'cardHistory', 'review')
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

function closedAtFromTiming(timing: ClosedTiming): string {
  const option = closedTimingOptions.find((item) => item.value === timing)
  const monthsAgo = option?.monthsAgo ?? 30
  const now = new Date()
  return new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - monthsAgo, 1)).toISOString()
}

function readStoredCardHistory(): UserCardHistoryItem[] {
  if (typeof window === 'undefined') {
    return []
  }
  try {
    const raw = window.localStorage.getItem(cardHistoryStorageKey)
    if (!raw) {
      return []
    }
    const parsed = JSON.parse(raw) as unknown
    if (!Array.isArray(parsed)) {
      return []
    }
    return parsed.filter(isStoredCardHistoryItem)
  } catch {
    return []
  }
}

function writeStoredCardHistory(history: UserCardHistoryItem[]) {
  if (typeof window === 'undefined') {
    return
  }
  try {
    window.localStorage.setItem(cardHistoryStorageKey, JSON.stringify(history))
  } catch {
    // Keep the in-memory profile usable if browser storage is unavailable.
  }
}

function isStoredCardHistoryItem(value: unknown): value is UserCardHistoryItem {
  if (!value || typeof value !== 'object') {
    return false
  }
  const item = value as Partial<UserCardHistoryItem>
  return (
    typeof item.issuer === 'string' &&
    typeof item.cardName === 'string' &&
    typeof item.currentlyHeld === 'boolean' &&
    (item.openedAt === undefined || typeof item.openedAt === 'string') &&
    (item.closedAt === undefined || typeof item.closedAt === 'string')
  )
}

export default App
