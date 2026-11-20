# Project

Startup prediction-market MVP built on the existing SocialPredict fork. This is a working online product, not a visual demo. Product decisions that are not settled are marked TBD.

## Vision

A prediction market only about startups, from pre-seed through late stage. People browse a startup, read its manually maintained funding history, reported valuations, and evidence, then trade YES/NO contracts tied to that company.

Example markets:

- Will Startup X reach a $100M valuation by 2028?
- Will Startup Y raise a Series A before 2028?

Each startup can have more than one market. Trades use the existing virtual-credit accounts. No real money.

## MVP

- 5–10 curated startups. The specific companies are TBD.
- Public startup directory and one profile page per startup.
- Profile content, entered by an admin: identity, funding history, reported valuations, and sourced evidence. Money is stored as integer minor units. An unknown amount is null, never zero.
- Multiple YES/NO markets linked to each startup.
- Existing user accounts, virtual-credit balances, buy/sell, and resolution.
- Admin management of startup records and market outcomes.
- Startup facts are updated by hand. External data APIs are later, not part of this MVP.

Information on profiles must be traceable to a source the admin records. A funding entry shows that source as a link.

## Out of scope

- Real-money trading and payment integrations.
- Automated scraping or live data feeds.
- Changing the trading math unless a later task explicitly asks for it.

## What already exists

SocialPredict already provides the pieces below. Reuse them. See `AGENTS.md` for where the code lives.

- Go API (`backend/`), React/Vite UI (`frontend/`), Postgres.
- YES/NO markets, virtual-credit buy and sell, resolution, and admin review.
- Accounts, JWT login, and balances.
- Market groups, which bundle multiple-choice answers. They are not startup profiles.

Startup identity records now exist in the API. The public directory is `/startups`, and each startup has a profile at `/startups/{slug}`. A market can optionally point at one startup.

## Open decisions

- Product name and branding. Temporary public name: Startup Predictions. A permanent name is still TBD.
- Which 5–10 startups to list. The curated roster is eight private companies in `backend/seed/startups.go`: Stripe, OpenAI, Databricks, Canva, Notion, Discord, Mercor, and Granola. A fresh database inserts any missing company and Mercor's sourced Series D when that announcement is not already stored. It does not change or delete rows that already exist, and it does not insert Anthropic, SpaceX, `step3-verify`, or `step8-admin`. Those test profiles stay only in this local database.
- Who may create a startup-linked market: admin only, or the current moderator flow. TBD.
- Invitation-only host, domain, and who receives access. TBD. A public deployment needs a paid Docker host.
- Whether virtual-credit amounts in `backend/setup/setup.yaml` stay as they are. TBD.
- Public startup reads. Provisional: `GET /v0/startups` and `GET /v0/startups/{slug}` are public for this MVP.

## Architecture snapshot

Checked against the current code. Step 3 added a startup record and API. The public directory is `/startups`, and each profile is `/startups/{slug}`. A market's optional `startup_id` is set by an admin and does not change trading math.

`backend/main.go` opens Postgres through `backend/internal/app/runtime/db.go`, applies migrations, and serves `backend/server/server.go`. New tables are added as files in `backend/migration/migrations/` that call `migration.Register`. GORM models live in `backend/models/`. `Market` in `backend/models/market.go` is one YES/NO question with a creator and a steward. `MarketGroup` is a multiple-choice bundle, not a company. `MarketTag` and `MarketTagAssignment` are the closest pattern: an admin-managed record that can point at markets without changing bet math.

`POST /v0/login` issues a JWT. `RequireAdmin` in `backend/internal/service/auth/auth_service.go` allows only `usertype` ADMIN and still enforces a required password change. Buys and sells are `POST /v0/bet` and `POST /v0/sell` (`backend/handlers/bets/`, `backend/internal/domain/bets/`). Resolution is `POST /v0/markets/{id}/resolve`. Admin user, review, and tag routes are registered under `/v0/admin/` in `backend/server/server.go`, with handlers in `backend/handlers/admin/`.

The UI routes are in `frontend/src/helpers/AppRoutes.jsx`. `frontend/src/api/httpClient.js` calls the API with `API_URL` from `frontend/src/config.js`. `frontend/src/pages/startups/Startups.jsx` is the public startup directory, and `frontend/src/pages/startups/StartupProfile.jsx` is one startup. `frontend/src/pages/markets/Markets.jsx` is the public market list. `frontend/src/pages/user/User.jsx` is a person, not a company. Admin UI is `frontend/src/pages/admin/AdminDashboard.jsx`. Tag create and update, the closest form, is `frontend/src/api/marketTagsApi.js` and `frontend/src/components/layouts/admin/ModeratorMarketReview.jsx`.

Step 3 stores `models.Startup` in `backend/models/startup.go` through migration `20261119120000`. Fields are id, name, unique slug, optional description, optional website URL, and timestamps. Admins create and update with `POST /v0/admin/startups` and `PATCH /v0/admin/startups/{slug}`. The slug stays fixed, and there is no delete. Public list and get-by-slug are provisional. Writes use `RequireAdmin`. Step 6 adds a nullable `startup_id` on `Market` and does not change bet math. Logos remain later work. The visible product name is the temporary label Startup Predictions. On startup, when the process is the database writer, `seed.SeedCuratedStartups` runs after the homepage seed. It inserts Stripe, OpenAI, Databricks, Canva, Notion, Discord, Mercor, and Granola only when that slug is absent. Descriptions come from each company's public website and do not include funding or valuation figures. The same seed then inserts Mercor's 10 November 2026 Series D only when that startup is active and no stored round already has the announcement URL `https://www.mercor.com/blog/series-d/` or the same Series D label and date. Amounts are integer cents: 75000000000 raised and 2000000000000 reported valuation, basis unspecified. A corrected round is not overwritten. A new database therefore has the eight companies and that one funding round, not the local test startups.

A funding round is `models.FundingRound`, added by migration `20261119140000`. It belongs to one startup and stores a round label, an `YYYY-MM-DD` announcement date, an optional amount raised, an optional reported valuation, a valuation basis (`pre-money`, `post-money`, or `unspecified`), a three-letter currency, a required http(s) evidence URL, a short source label, and an optional note. Amounts are integer minor units, such as cents, not floating-point values. Null means the figure was not disclosed, and zero is rejected. Public read is `GET /v0/startups/{slug}/funding`. An admin creates a round with `POST /v0/admin/startups/{slug}/funding` and corrects it with `PATCH /v0/admin/startups/{slug}/funding/{entryId}`. There is no delete. Rounds are listed from oldest to newest. The latest reported valuation is the newest round that includes a valuation, with its date, basis, currency, and source. That figure is a sourced fundraising valuation, not a prediction-market price or probability.

## Step 10A readiness

Checked on 20 November 2026. This is not a deployment, and Step 10 stays open.

Do not invite users until three things are true. The Step 9–9D work is saved in git; the current branch tip `1f0e776` does not contain the startup seed or funding. Beta uses a new database and images built from this fork, not this machine's database or `.env`, and not the upstream `openpredictionmarkets` images or its deploy workflow. The founder picks the host, domain, and whether anonymous visitors may read markets and balances.

What already holds: there is no public signup. An admin creates each account, and a new account must change its password before trading. Buys, sells, and startup admin writes reject a missing token. Proposed markets are not public. A fresh database seed is the eight private companies plus Mercor's sourced Series D. The packaged Docker layout can terminate HTTPS at Traefik, keeps Postgres on the internal network, and exposes `/health` and `/readyz`. `GET /ops/status` requires an authenticated admin and is not a public probe. `./SocialPredict backup` can save a database dump, but nothing schedules that backup.

Public trading privacy, decided for this MVP: a published or resolved market shows its YES/NO probability, probability history, total volume, participant count, and aggregated YES/NO volume. It does not show a named trader's bets, positions, share counts, spending, or profit. The same rule covers market-group activity and the global participant count. Market creator and steward names stay visible. Each person can still open their own account, and admin account tools stay on the admin routes. Step 10 stays open.
