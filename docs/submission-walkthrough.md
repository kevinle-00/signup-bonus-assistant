# Points Hacking Assistant - Decisions Walkthrough

I built a working Points Hacking Assistant that recommends the next Australian credit card sign-up bonus a user should consider. The frontend is deployed on Vercel, with a Go API and PostgreSQL database running on Railway.

My main product decision was to focus on one clear next action instead of building a broad credit card comparison table. For this type of product, users mainly need to know which card to target next, whether they can meet the spend requirement, what value they could get, and what eligibility risks to check. The app is built around that flow and returns a best recommendation, alternatives, caution cards, a value breakdown, and an action checklist.

The recommendation logic lives on the backend, while the frontend collects the user's spending assumptions, reward preferences, and card history. This keeps the decision model consistent and easier to test or change later. The app uses curated Australian credit card offer data stored in PostgreSQL. I chose curated data over live scraping because the key assessment focus was recommendation quality and user experience, not offer ingestion. The source data is maintained in a human-readable YAML file, and generated SQL seeds keep local and deployed databases reproducible.

The recommendation model balances estimated first-year value, spend achievability, eligibility confidence, reward goal match, and annual-fee comfort. I chose not to include points earned from required spend because category earn rates, caps, exclusions, and government-payment rules vary significantly between products. Without richer transaction data, those calculations would make the result appear more precise than it really is.

Eligibility is treated as confidence rather than certainty. The app uses self-reported card history and structured offer rules to flag likely issues, while still encouraging users to verify issuer terms. This is why the result separates recommended cards from caution cards instead of hiding every card with a possible eligibility issue.

The frontend is mobile-first because this flow felt more natural as a guided assistant than as a dense dashboard. The user answers one question at a time, reviews their inputs, then sees the recommendation with "Why this card?" directly under the main result. Positive fit signals are visually distinct from warnings so the user can quickly understand both the upside and the risks.

I also decided not to build a multi-card sequencing engine. Planning several applications ahead can look useful, but it depends on future approvals, changing offers, bonus posting dates, and updated card history. I thought it was more trustworthy to recommend one next card clearly and let the user rerun the assistant once their situation changes.

The main limitations are intentional: no authentication, bank linking, live offer scraping, compliance-grade advice workflow, or long-term card-churning planner. These would be natural future extensions, but this version focuses on proving the core recommendation experience: "What card should I target next, why, and what value could I reasonably expect?"
