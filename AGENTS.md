# AGENTS.md

## Current state

- This repo is a placeholder: it contains only `README.md`; there is no source, manifest, tests, lint/format config, or CI. Inspect the tree before assuming commands; do not invent them.
- The `origin` remote is `https://github.com/dhanujanarada/basementplays.git`; local `main` tracks `origin/main`.

## Related codebase

- `../basement_booking_platform` is the existing Basement Plays booking processor and the only related code in this workspace. Its Google Cloud project ID is `basementplays`.
- That project uses Python 3.11 and `requirements.txt` (`gspread`, `oauth2client`, `qrcode`, `Pillow`). Run it from `scripts/` (`cd scripts && python main.py`) so its local imports resolve; `scripts/main.py` is the entrypoint.
- It requires the env vars `SHEET_ID`, `SENDER_EMAIL`, and `SENDER_APP_PASSWORD`, plus `service-account.json` at the repo root. Never commit that file or other credentials.
- Its CI (`.github/workflows/process_bookings.yml`) runs every 10 minutes, writes the service-account secret to `service-account.json`, then runs `python main.py` from `scripts/`.

<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->
