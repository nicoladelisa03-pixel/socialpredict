# Roadmap

Build the startup MVP on the current SocialPredict app. Do one step at a time. Leave a box unchecked until that step's acceptance criterion has been tested. Do not mark a step done from a plan or a code reading.

Product scope is in `PROJECT.md`. Working rules are in `AGENTS.md`.

1. [x] Verify the existing SocialPredict installation and test original trading.
   - Acceptance: the API `/health` and `/readyz` succeed, a user can sign in, buy and sell a YES/NO position with virtual credits, and an admin can resolve that market. Record the commands and results.

2. [x] Inspect and document the application architecture.
   - Acceptance: a short note lists the real frontend, backend, database, auth, market, trading, and admin entry points, and states that no startup entity exists yet. `PROJECT.md` already holds the product summary; do not paste the architecture into it.

3. [x] Design and implement startup data records.
   - Acceptance: an admin can create and edit a startup record stored in Postgres. The record has a name, a unique slug, an optional short description, an optional website, and timestamps. Trading, balances, and market resolution still pass their existing tests.

4. [x] Build the startup directory.
   - Acceptance: a public visitor can open `/startups`, see the startups returned by the API, and open a startup website when one is saved. Profile pages come in the next step.

5. [x] Build individual startup profile pages.
   - Acceptance: each startup has its own page showing the name, description, and website saved in step 3. Funding history, reported valuations, and evidence are still to be added.

6. [x] Link multiple prediction markets to each startup.
   - Acceptance: one startup profile shows two or more existing YES/NO markets, and those markets still trade with the current buy/sell flow.

7. [x] Verify end-to-end virtual-money trading and settlement.
   - Acceptance: on a startup-linked market, a user buys, sells, and an admin resolves the market so credits settle. No real-money path is added.
   - Public detail, summary, bet, position, leaderboard, and projection reads return 404 for a proposed market. Admin review stays on the admin market queue.

8. [x] Implement a simple admin workflow for startup data and market management.
   - Acceptance: an admin can update a startup record and manage that startup's markets from the existing admin area, without a direct database edit.

9. [x] Add 5–10 real startup profiles and customize branding.
   - Acceptance: 5–10 real startups are visible with sourced profile facts, and the agreed branding (TBD) replaces the leftover BrierFoxForecast seed content on the home page.
   - The temporary public name is Startup Predictions. Eight sourced company profiles were added. Existing test startups were left in place.

9B. [x] Keep the curated roster limited to private companies, and make that roster reproducible.
   - Acceptance: Anthropic and SpaceX are not in the public directory, Mercor and Granola are, and a fresh database can insert the eight curated companies without changing rows that already exist.
   - The two public-company rows had no markets. They were hidden with `deleted_at` and were not hard-deleted. `step3-verify` still has its four market links. Test profiles stay visible and are labeled test records.

9C. [x] Record sourced funding rounds and reported valuations.
   - Acceptance: an admin can add and correct a startup's funding history from Manage Startups, and the public profile shows that history, the latest reported valuation, and its source. Unknown amounts stay empty. No price feed is added.
   - Mercor's 10 November 2026 Series D is stored from Mercor's own announcement. Granola's site did not state a funding round, so that profile has no funding entry.

9D. [x] Make the private-beta roster and Mercor's sourced round reproducible.
   - Acceptance: a fresh database seeds only Stripe, OpenAI, Databricks, Canva, Notion, Discord, Mercor, and Granola, plus Mercor's 10 November 2026 Series D from the company's announcement. Repeating the seed does not duplicate that round or overwrite a correction. `step3-verify` and `step8-admin` stay in this local database and are not part of the seed.
   - Checked with a separate sqlite database and a rolled-back pass over the local database. Local users, balances, markets, test startups, and the existing Mercor round were unchanged.

10. [ ] Test the application and prepare a secure invitation-only deployment.
    - Acceptance: the MVP checklist above is retested, secrets are not committed, and an invitation-only deploy target is written down. Publishing still waits on the host decision in `PROJECT.md`.
