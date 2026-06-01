# Remaining Work Plan

## Context

The original brief is:

```text
Help users identify the best credit card to switch to next based on their spending patterns, current card history, and available bonus offers. Show them clearly what they stand to gain and what they need to do to qualify.
```

The primary end-to-end flow now works:

```text
React wizard
→ Go API
→ Postgres card offers
→ recommendation engine
→ roadmap response
→ polished result screen
→ recommendation run snapshot
```

The remaining work should deepen the product around the brief, not add unrelated features.

## Phase 1: Profile And Card History

Status: implemented in the current frontend/backend slice. Keep this section as the acceptance record for what was added and what remains explicitly out of scope.

### Goal

Make the recommendation genuinely depend on the user's current and recent card history, and give the app a small production-minded profile area for reviewing and editing that data.

The backend supports `cardHistory`, and the frontend now collects a lightweight self-reported version through the profile/card-history flow. This closes the biggest product gap relative to the brief without adding auth or account linking.

This phase should also remove free-typed spend amounts from the main mobile flow. The polished wizard should prefer range option cards because the model is approximate and mobile users should not need to type exact dollar figures.

### Scope

Add a small profile section and simple card-history collection.

Profile navigation:

- Add a profile/settings entry point from the app shell.
- Use simple React view state; do not add routing unless it becomes clearly necessary.
- Suggested views:
  - `wizard`
  - `profile`
  - `cardHistory`

Profile screen:

- Show a simple profile header.
- Show navigation cards for:
  - `Card history`
  - `Preferences`
  - `Latest recommendation`
  - `About this estimate`
- Keep these lightweight; only `Card history` needs meaningful functionality in this phase.

Card history screen:

- Show current/recent cards already entered.
- Allow adding a current card.
- Allow adding a recently closed card.
- Allow removing an entry.
- Store in React state and persist to `localStorage` so history survives refresh.
- Use a dedicated mobile issuer picker screen with known issuers rather than arbitrary free text.
- Use option cards for card status instead of a native dropdown.
- Use closed-timing range cards instead of exact date/month entry.
- Keep card name as self-reported free text because a complete card-product catalogue is out of scope.
- Explain that this is self-reported history and not proof of issuer eligibility.

Wizard integration:

Screens:

1. Add an eligibility-check step after Amex preference.
2. Let the user continue without history if unsure.
3. Let the user open the shared card-history screen from either the wizard or profile.
4. For current cards, collect issuer and card name.
5. For recently closed cards, collect issuer, card name, and rough closed timing.
6. Include this data in `POST /api/recommendations` as `cardHistory`.

Spend range screens:

Monthly card spend options:

- `$0-$1,000` → send `$750`
- `$1,000-$2,000` → send `$1,500`
- `$2,000-$4,000` → send `$3,000`
- `$4,000-$6,000` → send `$5,000`
- `$6,000-$10,000` → send `$8,000`
- `$10,000+` → send `$10,000`

Large purchases in the next 90 days:

- `$0` → send `$0`
- `$1-$1,000` → send `$750`
- `$1,000-$3,000` → send `$2,000`
- `$3,000-$5,000` → send `$4,000`
- `$5,000+` → send `$5,000`

Use conservative representative values. Above `$10,000/month`, exact spend does not materially affect the MVP because normal single-card sign-up requirements are already comfortably met.

Keep it simple:

- No full bank/card search catalogue.
- No account linking.
- No exact opened date requirement.
- Limit to a small number of entries if needed.
- No authentication or real user account.

### Acceptance Criteria

- User can add at least one currently held card.
- User can add at least one recently closed card.
- User can review and remove card-history entries from a profile/card-history screen.
- Card-history entries persist across refresh using `localStorage`.
- Issuer entry uses a dedicated mobile picker screen of canonical known issuers, not arbitrary free text.
- Card status and closed timing use large tap targets, not native dropdown/date controls.
- Wizard submits the same card-history state shown in the profile section.
- User selects spend ranges instead of typing exact monthly spend/large purchase amounts.
- The submitted request still sends integer cent values expected by the backend.
- Submitted `cardHistory` changes eligibility confidence/recommendations when matching issuer/card rules apply.
- Backend issuer matching handles common aliases such as `Amex` and `Commonwealth Bank`/`CommBank`, plus the St.George/BankSA/Bank of Melbourne regional issuer group.
- Result UI calls out when saved card history affected eligibility confidence.
- Review screen shows card-history summary.
- Existing wizard remains mobile-first.
- Frontend build/lint pass.
- Backend checks pass if touched.

## Phase 2: Switching Roadmap Copy

### Goal

Make the result explicitly answer: “what card should I switch to next, and what should I do over the next 6-12 months?”

### Scope

Improve frontend copy and result sections using existing backend roadmap data.

Add/adjust result copy:

- `Switch to this card next` framing.
- Timeline-style section:
  - Month 0: verify terms and apply.
  - Months 1-3: meet spend requirement.
  - Month 4: track bonus posting.
  - Month 11: review before annual fee renewal.
  - Month 12: re-run assistant for next opportunity.
- Clear statement that the MVP recommends one immediate card only.

### Acceptance Criteria

- User can understand the switch plan within 5 seconds.
- Checklist and timeline are more prominent than alternatives.
- Disclaimer remains visible.
- No backend model changes unless clearly needed.

## Phase 3: Frontend Quality Pass

### Goal

Make the product feel polished enough for a take-home demo.

### Scope

- Keyboard/focus states for wizard controls.
- Better mobile spacing around 390px width.
- Better empty/no-recommendation state.
- Better API error state.
- Ensure all buttons have clear labels.
- Avoid raw backend field names in user-facing copy.
- Keep desktop as centred mobile shell.

### Acceptance Criteria

- Wizard can be completed by keyboard.
- Form errors are clear and non-technical.
- Result screen remains readable with long card names.
- Frontend build/lint pass.

## Phase 4: CI And Test Hardening

### Goal

Make the project more production-minded and easier to evaluate.

### Scope

- Add frontend CI job:
  - `npm ci`
  - `npm run build`
  - `npm run lint`
- Consider a lightweight repository integration test path for Postgres mapping.
- Keep handler tests fake-repository based.
- Document any integration test setup if added.

### Acceptance Criteria

- CI verifies backend and frontend checks.
- SQL/JSONB mapping has either an automated integration test or a documented smoke-test path.
- No flaky external dependencies.

## Phase 5: README And Demo Polish

### Goal

Make the final project easy to run, review, and discuss.

### Scope

Update README with:

- Project summary.
- Architecture diagram.
- Local setup from fresh clone.
- Database migration/seed instructions.
- Backend run instructions.
- Frontend run instructions.
- Test/check commands.
- Smoke-test inputs and expected output.
- Product assumptions.
- Trade-offs.
- Future work.

### Acceptance Criteria

- A reviewer can run the app locally without asking for missing steps.
- The README clearly explains why auth/live scraping/card images are excluded.
- The README highlights core engineering decisions: domain-first engine, curated data, Postgres JSONB, fake-repo handler tests, snapshot persistence.

## Explicitly Defer

Do not add these before the core take-home is polished:

- Authentication/users.
- Bank account linking.
- Live card-offer scraping.
- Real bank logos or card images.
- Multi-card sequencing.
- Credit-score modelling.
- Complex rewards redemption modelling.
- OpenAI-generated explanations.

These are valid future directions, but they distract from the core brief right now.
