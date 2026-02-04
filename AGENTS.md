# AGENTS.md

## Purpose
This repository is a two-part app for remotely controlling OBS Studio:
- `server/`: Go + Wails desktop server that talks to OBS over WebSocket.
- `client-web/`: Vite web client wrapped with Electron (desktop) and Capacitor (mobile).

When making changes, keep diffs focused, avoid broad rewrites, and preserve current behavior unless the task explicitly asks for a behavioral change.

## Repo Map
- `server/main.go`, `server/app.go`: app entrypoints.
- `server/internal/api/`: HTTP API routes/handlers.
- `server/internal/manager/`: OBS, config, session, and button orchestration.
- `server/internal/models/`: shared data models.
- `server/internal/storage/`: persistence.
- `client-web/src/js/app.js`: main UI logic/state.
- `client-web/src/js/api.js`: API client calls to server.
- `client-web/electron-main.cjs`: Electron main process.
- Root `Makefile`: common dev/build/release workflows.

## Default Workflow
1. Read the target files and nearby modules before editing.
2. Make the smallest viable change that solves the request.
3. Run relevant checks (at minimum build/smoke checks for changed area).
4. Report what changed, why, and any follow-up risks.

## Dev Commands
From repo root:
- Server dev: `make dev-server`
- Client dev: `make dev-client`
- Server tests: `make test-server` (or `cd server && go test ./...`)
- Server lint: `cd server && golangci-lint run`
- Full build: `make all`

Client notes:
- `client-web/package.json` does not currently define a reliable `test` script.
- For client changes, prefer `cd client-web && npm run build` as the baseline verification unless a task adds tests.

## Editing Rules
- Prefer editing source files, not generated artifacts.
- Do **not** manually edit build outputs unless the user explicitly asks:
  - `client-web/dist/`
  - `client-web/electron-dist/`
  - `server/frontend/dist/`
  - `server/build/`
  - `releases/`
- Keep API contracts backward-compatible when possible (client/server are tightly coupled across versions).
- Reuse existing patterns in managers/models before introducing new abstractions.

## Server Conventions (Go)
- Keep manager responsibilities separated (OBS logic in `obs_manager`, config in `config_manager`, etc.).
- Return actionable errors with context; avoid silent failures.
- Preserve thread-safety and session integrity when touching shared state.
- Prefer explicit structs/types over untyped maps.

## Client Conventions (Web/Electron/Capacitor)
- Keep UI changes responsive for phone/tablet/desktop layouts.
- Preserve existing localStorage keys and session behavior unless migration is intentional.
- Avoid adding new framework dependencies unless requested.
- Keep server URL and connection flows compatible with current onboarding docs.

## Validation Checklist Before Hand-off
- Changed files are scoped to the task.
- Relevant build/test commands were run (or clearly noted if not run).
- No accidental edits to generated artifacts.
- Any API or config format change is documented in the response.

## Release/Safety Notes
- Do not perform destructive git operations (`reset --hard`, force checkout, deleting user changes).
- Treat signing/versioning files carefully:
  - `server/wails.json`
  - `client-web/package.json`
  - mobile/electron signing settings
- If a task impacts packaging/release, reference `RELEASE.md` and root `Makefile` targets.
