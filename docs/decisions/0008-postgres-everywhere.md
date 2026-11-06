# 0008 — Postgres everywhere; SQLite dropped from the site

- **Status:** accepted
- **Covers:** `src/payload.config.ts`, `ticketing/internal/db/`, `ticketing/migrations/`
- **Last checked:** 8de3320
- **Implementation:** partial
- **Date:** 2026-11
- **Related:** [0009](0009-separate-go-service.md), [0011](0011-hosting-portability.md), `src/payload.config.ts`

## Context

Running SQLite for the CMS and Postgres for ticketing would mean two database engines and a production CMS that needs a persistent disk.

## Decision

Use Postgres for both the site and the Go service. Develop against native Postgres on Nobara Linux; Docker is not required. Do not use SQLite in production.

## Alternatives considered

- **Supabase free tier** — secondary sources say it pauses after about a week of inactivity and has no automatic backups; current terms need rechecking (unverified).
- **SQLite in production** — needs a persistent disk and adds a second engine.

## Consequences and trade-offs accepted

One database engine and one backup story. Current site code still uses Payload's SQLite adapter, so this decision is only partly implemented; the Go service already uses Postgres.

## Revisit when

Revisit when Payload's adapter is switched to Postgres, or if the hosting choice makes a managed Postgres materially cheaper.

## Related

[0009](0009-separate-go-service.md), [0011](0011-hosting-portability.md), `src/payload.config.ts`
