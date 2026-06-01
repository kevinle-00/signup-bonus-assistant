# Frontend Design Guide

## Reference Assets

The current design references live in:

```text
docs/design-reference/
```

They appear to come from the same startup's mortgage and utility products, so the credit-card assistant should reuse the same visual language rather than inventing a new one.

These files are design references, not runtime app assets. Keeping them under `docs/` makes that clear. If we later export real icons, logos, or images used by the frontend, those should live inside the frontend app, for example `frontend/src/assets/` or `frontend/public/`.

## Overall Direction

Build the frontend as a mobile-first, dark, high-contrast assistant experience.

The product should feel like:

- A financial concierge, not a spreadsheet.
- Direct and action-oriented.
- Sparse, premium, and slightly editorial.
- More like a guided app flow than a generic dashboard.

Avoid generic SaaS cards, blue gradients, dense tables, and default form layouts.

## Visual Language

Core traits from the reference screens:

- Black background as the dominant surface.
- White primary text.
- Muted grey secondary text.
- Warm amber/orange accent for savings, alerts, progress, and highlights.
- White filled primary buttons.
- Thin white/grey outlined secondary buttons and cards.
- Rounded rectangles, usually medium radius.
- Large whitespace and strong vertical rhythm.
- Minimal borders rather than shadows.

Suggested colour tokens:

```text
background: #000000
surface: #0A0A0A
surfaceRaised: #111111
textPrimary: #FFFFFF
textSecondary: #9A9A9A
border: #3A3A3A
borderStrong: #FFFFFF
accent: #F6AA45
accentHot: #FF5A1F
error: #FF2F7D
buttonPrimaryBg: #F7F7F7
buttonPrimaryText: #000000
```

These do not need to be exact yet, but the first frontend pass should use tokens so we can tune them later.

## Typography

The references use a clean, modern sans-serif for headings and body text, with a monospaced/all-caps microcopy style for system/status text.

Use:

- Large lightweight headings.
- Regular-weight body copy.
- Small uppercase monospaced labels for status lines, pills, metadata, and scanning-style copy.
- Avoid heavy bold headings except for emphasis inside cards.

Approximate hierarchy:

```text
pageTitle: 28-34px, light/regular
sectionTitle: 18-22px, regular
body: 15-17px, regular
caption: 13-14px, muted
monoLabel: 10-12px, uppercase, letter-spaced
```

## Layout

Start mobile-first around a 390px wide viewport.

Use:

- Full black app shell.
- 16px page gutters.
- 24-40px vertical spacing between major sections.
- Cards with 1px borders and 12-16px radius.
- Large bottom primary actions on form screens.
- Progress or status near the top-right for wizard steps.

Desktop can initially centre the same mobile-width app shell. Do not stretch the UI into a wide dashboard until we have a specific desktop design.

## Components

### Buttons

Primary buttons:

- White fill.
- Black text.
- Large height, around 56-64px.
- Rounded corners.

Secondary buttons:

- Transparent black background.
- White border.
- White text.

Pills:

- Small rounded capsules.
- Use white fill for selected filters.
- Use outline for unselected filters.
- Use amber/orange for alert/status pills.

### Cards

Use cards for recommendations and account-like entities.

Recommendation cards should include:

- Small status pill, for example `BEST OFFER` or `CAUTION`.
- Issuer/card name.
- Big value statement.
- Metadata line for points, fee, or spend requirement.
- Strong action row/button.

Avoid dense tables in the smoke UI. Lists of alternatives can be stacked cards.

### Forms

Forms should feel like wizard screens, not admin forms.

Use:

- One major question per section when practical.
- Large tap targets.
- White input backgrounds only when the design calls for high contrast selection cards.
- Dark inputs with white borders for numeric entry.
- Error text in hot pink/red.
- Helper text in muted grey.

For the first smoke UI, it is acceptable to put all fields on one page, but style each field as a guided question block.

## Content Tone

Use concise assistant-like copy.

Good examples:

- `There is a strong Qantas offer for you.`
- `Spend $5,000 in 90 days to unlock the bonus.`
- `Review these eligibility cautions before applying.`

Avoid:

- `Submit form`
- `An error occurred`
- `Recommendation result object`
- raw backend field names in the UI.

## Credit-Card Adaptation

Map the mortgage design language into credit-card concepts:

- `Potential savings` becomes `Estimated year-one value`.
- `Offer found` becomes `Best card found`.
- `Check approval odds` becomes `Review eligibility`.
- `Meet spend requirement` becomes the primary action checklist item.
- `Completed tasks` becomes `Action checklist`.

The frontend should make the domain result feel like a guided plan, not just an API response viewer.

## First Smoke UI Target

The initial frontend should prove the end-to-end flow:

```text
Form input
→ POST /api/recommendations
→ render roadmap
```

Minimum UI sections:

- Intro/status hero.
- Onboarding form.
- Best recommendation card.
- Reasons and warnings.
- Action checklist.
- Alternatives.
- Caution cards.

Keep it simple, but make it visually consistent with the references from the first pass.
