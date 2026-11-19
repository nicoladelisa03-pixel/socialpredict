# Agent rules

Modify this SocialPredict fork. Do not rebuild it. Product scope is `PROJECT.md`. Do one unchecked item in `ROADMAP.md`, then stop.

## How to work

- Read the code you will change before editing it.
- Reuse current accounts, YES/NO markets, virtual-credit buy/sell, resolution, and admin screens.
- Keep each change small. One roadmap step per task.
- Do not rewrite the trading engine unless the user explicitly says so. Trading code is under `backend/handlers/bets/` and `backend/internal/domain/math/`.
- Do not add real-money trading, payments, or scrapers unless the user explicitly asks.
- Do not run destructive commands or overwrite database data unless the user approves that exact command.
- Preserve current behavior. After a change, run the tests that cover it.
- Say a test passed only if you executed it, and include the result.
- Check a roadmap box only after that step's acceptance criterion has been verified.
- Explain changes so a beginner can follow them.
- Do not add files, dependencies, or long docs you do not need.
- Do not start the next roadmap step on your own.
- Use the repo's Git workflow. Do not commit or push unless the user asks.

## Where to look

- API routes: `backend/server/server.go`. Human API notes: `backend/docs/README.md`.
- Models: `backend/models/`. Migrations: `backend/migration/`.
- UI routes: `frontend/src/helpers/AppRoutes.jsx`.
- Local run: `README/LOCAL_SETUP.md`, then `./SocialPredict up` (Docker). Economics: `backend/setup/setup.yaml`.
- Market groups are multiple-choice bundles, not startups.

## End of task

Report only:

a. What changed
b. Files modified
c. Tests performed and their results
d. Remaining problems
e. Suggested next action
