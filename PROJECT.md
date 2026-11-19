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

There is no startup record, directory, or market-to-startup link yet.

## Open decisions

- Product name and branding. TBD.
- Which 5–10 startups to list. TBD.
- Who may create a startup-linked market: admin only, or the current moderator flow. TBD.
- Invitation-only host, domain, and who receives access. TBD. A public deployment needs a paid Docker host.
- Whether virtual-credit amounts in `backend/setup/setup.yaml` stay as they are. TBD.
