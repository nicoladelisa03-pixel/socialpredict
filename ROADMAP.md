# Roadmap

Build the startup MVP on the current SocialPredict app. Do one step at a time. Leave a box unchecked until that step's acceptance criterion has been tested. Do not mark a step done from a plan or a code reading.

Product scope is in `PROJECT.md`. Working rules are in `AGENTS.md`.

1. [ ] Verify the existing SocialPredict installation and test original trading.
   - Acceptance: the API `/health` and `/readyz` succeed, a user can sign in, buy and sell a YES/NO position with virtual credits, and an admin can resolve that market. Record the commands and results.

2. [ ] Inspect and document the application architecture.
   - Acceptance: a short note lists the real frontend, backend, database, auth, market, trading, and admin entry points, and states that no startup entity exists yet. `PROJECT.md` already holds the product summary; do not paste the architecture into it.

3. [ ] Design and implement startup data records.
   - Acceptance: an admin can create and edit a startup record stored in Postgres. Trading, balances, and market resolution still pass their existing tests. Field list is TBD until this step starts.

4. [ ] Build the startup directory.
   - Acceptance: a signed-in or public visitor (visibility TBD) can open one page that lists the saved startups and reach a startup from that list.

5. [ ] Build individual startup profile pages.
   - Acceptance: each startup has its own page showing the manually entered identity, funding history, reported valuations, and evidence saved in step 3.

6. [ ] Link multiple prediction markets to each startup.
   - Acceptance: one startup profile shows two or more existing YES/NO markets, and those markets still trade with the current buy/sell flow.

7. [ ] Verify end-to-end virtual-money trading and settlement.
   - Acceptance: on a startup-linked market, a user buys, sells, and an admin resolves the market so credits settle. No real-money path is added.

8. [ ] Implement a simple admin workflow for startup data and market management.
   - Acceptance: an admin can update a startup record and manage that startup's markets from the existing admin area, without a direct database edit.

9. [ ] Add 5–10 real startup profiles and customize branding.
   - Acceptance: 5–10 real startups are visible with sourced profile facts, and the agreed branding (TBD) replaces the leftover BrierFoxForecast seed content on the home page.

10. [ ] Test the application and prepare a secure invitation-only deployment.
    - Acceptance: the MVP checklist above is retested, secrets are not committed, and an invitation-only deploy target is written down. Publishing still waits on the host decision in `PROJECT.md`.
