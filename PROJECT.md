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
- Profile content, entered by an admin: identity, funding history, reported valuations, and sourced evidence. The exact fields are TBD.
- Multiple YES/NO markets linked to each startup.
- Existing user accounts, virtual-credit balances, buy/sell, and resolution.
- Admin management of startup records and market outcomes.
- Startup facts are updated by hand. External data APIs are later, not part of this MVP.

Information on profiles must be traceable to a source the admin records. How sources are displayed is TBD.

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

- Product name and branding. TBD.
- Which 5–10 startups to list. TBD.
- Who may create a startup-linked market: admin only, or the current moderator flow. TBD.
- Invitation-only host, domain, and who receives access. TBD. A public deployment needs a paid Docker host.
- Whether virtual-credit amounts in `backend/setup/setup.yaml` stay as they are. TBD.
- Public startup reads. Provisional: `GET /v0/startups` and `GET /v0/startups/{slug}` are public for this MVP.

## Architecture snapshot

Checked against the current code. Step 3 added a startup record and API. The public directory is `/startups`, and each profile is `/startups/{slug}`. A market's optional `startup_id` is set by an admin and does not change trading math.

`backend/main.go` opens Postgres through `backend/internal/app/runtime/db.go`, applies migrations, and serves `backend/server/server.go`. New tables are added as files in `backend/migration/migrations/` that call `migration.Register`. GORM models live in `backend/models/`. `Market` in `backend/models/market.go` is one YES/NO question with a creator and a steward. `MarketGroup` is a multiple-choice bundle, not a company. `MarketTag` and `MarketTagAssignment` are the closest pattern: an admin-managed record that can point at markets without changing bet math.

`POST /v0/login` issues a JWT. `RequireAdmin` in `backend/internal/service/auth/auth_service.go` allows only `usertype` ADMIN and still enforces a required password change. Buys and sells are `POST /v0/bet` and `POST /v0/sell` (`backend/handlers/bets/`, `backend/internal/domain/bets/`). Resolution is `POST /v0/markets/{id}/resolve`. Admin user, review, and tag routes are registered under `/v0/admin/` in `backend/server/server.go`, with handlers in `backend/handlers/admin/`.

The UI routes are in `frontend/src/helpers/AppRoutes.jsx`. `frontend/src/api/httpClient.js` calls the API with `API_URL` from `frontend/src/config.js`. `frontend/src/pages/startups/Startups.jsx` is the public startup directory, and `frontend/src/pages/startups/StartupProfile.jsx` is one startup. `frontend/src/pages/markets/Markets.jsx` is the public market list. `frontend/src/pages/user/User.jsx` is a person, not a company. Admin UI is `frontend/src/pages/admin/AdminDashboard.jsx`. Tag create and update, the closest form, is `frontend/src/api/marketTagsApi.js` and `frontend/src/components/layouts/admin/ModeratorMarketReview.jsx`.

Step 3 stores `models.Startup` in `backend/models/startup.go` through migration `20261119120000`. Fields are id, name, unique slug, optional description, optional website URL, and timestamps. Admins create and update with `POST /v0/admin/startups` and `PATCH /v0/admin/startups/{slug}`. The slug stays fixed, and there is no delete. Public list and get-by-slug are provisional. Writes use `RequireAdmin`. Step 6 adds a nullable `startup_id` on `Market` and does not change bet math. Funding, valuation, and logos remain later work.
