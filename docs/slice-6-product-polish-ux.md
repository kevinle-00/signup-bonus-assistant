# Slice 6: Product Polish UX

## Goal

Turn the frontend smoke UI into a guided mobile-first assistant flow that matches the design reference direction.

The current single-page form is only a test harness. Slice 6 should replace it with a wizard-style experience before visual polish goes deeper.

## Design Constraint

Use `docs/design-reference/Form wizard.png` as the primary interaction reference.

Important traits to preserve:

- One main question per screen.
- Black background.
- Large lightweight question text.
- Small warm value/status text near the top.
- Progress indicator near the top-right.
- Large tap targets.
- Bottom primary action when needed.
- Sparse helper text.
- Minimal chrome.

Desktop should centre the same mobile app shell. Do not create a wide dashboard for this slice.

## User Flow

### Screen 1: Intro

Purpose: explain what the assistant will do.

Content:

- Status pill: `POINTS ENGINE`
- Heading: `Find your next best credit card bonus`
- Body: `Answer a few questions and we’ll estimate which sign-up bonus is worth targeting next.`
- Primary CTA: `Start`

Action:

- Tapping `Start` moves to goal selection.

### Screen 2: Reward Goal

Question:

```text
What are you optimising for?
```

Options:

- `Qantas Points`
- `Velocity Points`
- `Maximum net value`
- `Cashback`
- `Low effort`

Interaction:

- Option cards, not a dropdown.
- Tapping an option selects it and advances automatically.

### Screen 3: Monthly Spend

Question:

```text
Roughly how much do you put on cards each month?
```

Helper:

```text
A ballpark is fine. Exclude rent, mortgage repayments, and anything you would not usually put on a card.
```

Input:

- Money input.
- Default from smoke UI can remain `$2,500` for local testing, but final UX should not feel pre-filled as a recommendation.

Action:

- Bottom `Continue` button.

### Screen 4: Large Purchases

Question:

```text
Any large card purchases coming up in the next 90 days?
```

Helper:

```text
Flights, appliances, insurance, or planned bills can make a bonus easier to reach.
```

Input:

- Money input.
- Allow `$0`.

Action:

- Bottom `Continue` button.

### Screen 5: Annual Fee Preference

Question:

```text
How should we treat annual fees?
```

Options:

- `Flexible if the value is strong`
- `Prefer lower fees`
- `Set a strict maximum`

Interaction:

- Option cards.
- If the user chooses strict maximum, continue to Screen 6.
- Otherwise skip Screen 6 and continue to Amex preference.

### Screen 6: Maximum Annual Fee

Condition: only shown if annual fee preference is `strict_max`.

Question:

```text
What is the most you are comfortable paying upfront?
```

Input:

- Money input.

Action:

- Bottom `Continue` button.

### Screen 7: Amex Preference

Question:

```text
Are you open to American Express cards?
```

Options:

- `Yes, include Amex cards`
- `No, avoid Amex cards`

Interaction:

- Option cards.
- Tapping an option selects it and advances to review.

### Screen 8: Review

Purpose: show the selected inputs before submitting.

Content:

- Goal.
- Monthly spend.
- Large purchases.
- Annual fee preference.
- Maximum annual fee if relevant.
- Amex preference.

Actions:

- Primary: `Find my best offer`
- Secondary: `Back`

### Screen 9: Loading

Purpose: make the recommendation feel like an assistant doing work.

Content:

- Status text: `Scanning active offers`
- Supporting text: `Checking value, spend achievability, and eligibility confidence.`
- Optional subtle progress/ring treatment inspired by `Frame 2147261057.png`.

Do not fake a long delay. Show this only while the request is actually pending.

### Screen 10: Result

Purpose: answer the core user question clearly.

Top hierarchy:

1. `Best card found` status pill.
2. Card name and issuer.
3. Big value statement: `Estimated year-one value of $830`.
4. Key facts row:
   - bonus points
   - minimum spend
   - annual fee
   - spend difficulty

Primary action section:

- Heading: `Your next steps`
- Render action checklist immediately after the best card.
- The spend requirement item should be visually prominent.

Explanation section:

- Heading: `Why this card`
- Show backend reasons.
- Do not show raw score as a headline.
- If score is shown, keep it as small metadata only.

Warnings section:

- Only show if warnings exist.
- Use caution styling, not panic styling.
- Copy should say `Review before applying`, not `Error`.

Alternatives section:

- Lower priority than checklist and reasons.
- Show up to three alternative cards.
- Use compact cards.

Caution cards section:

- Collapsed or visually de-emphasised if many cards exist.
- Explain that these are cards not selected because of eligibility, spend, or review cautions.

Disclaimer:

Show near the end of the result:

```text
This tool provides estimates based on curated offer data and simplified assumptions. It is not financial advice. Always check issuer terms and consider your personal circumstances before applying.
```

## Navigation Rules

- Every wizard screen after intro has a back control.
- Progress should show current step percentage or count.
- Do not allow submission until required fields are valid.
- Keep form state in React state for now; do not add a form library yet.
- If API submission fails, keep the user's inputs and show a styled error state with retry.

## Validation Rules

- Monthly spend must be greater than `$0`.
- Large purchases can be `$0` or more.
- Strict maximum annual fee must be greater than `$0` when selected.
- All option screens require a selected value before proceeding, except option-card screens that auto-advance on selection.

## Empty And Error States

No recommendation:

- Show `No safe card found yet`.
- Explain backend-provided no-recommendation reasons.
- Offer a `Review answers` action.

API error:

- Show `We could not check offers right now`.
- Include the backend error message if safe.
- Offer `Try again` and `Review answers`.

## Acceptance Criteria

- The onboarding is wizard-based, not a single long form.
- The UI remains mobile-first and usable around 390px width.
- Desktop centres the mobile shell rather than becoming a wide dashboard.
- The user can understand the best card and next action within 5 seconds of the result screen loading.
- Checklist appears before alternatives.
- Warnings are clear but not alarming.
- Disclaimer is present.
- Backend response shape is not exposed raw to the user.
- Existing frontend build and lint commands pass.
- Existing backend checks continue to pass if touched.
