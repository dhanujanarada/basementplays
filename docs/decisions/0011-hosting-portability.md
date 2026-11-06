# 0011 — Hosting kept portable

- **Status:** accepted
- **Covers:** `ticketing/Dockerfile`, `ticketing/Makefile`, `next.config.ts`, `.env.example`
- **Last checked:** 8de3320
- **Implementation:** partial
- **Date:** 2026-11
- **Related:** [0008](0008-postgres-everywhere.md), [0009](0009-separate-go-service.md), `ticketing/Dockerfile`

## Context

The final host is not chosen. Moving between one VPS, separate hosts, or a hybrid should be configuration, not a rewrite.

## Decision

Keep configuration in environment variables, ship containers, use a same-origin ticket API path, use header tokens instead of cookies, keep media storage switchable, and keep the Go service stateless. Choose A (one VPS with Compose and Caddy), B (separate hosts), or C (hybrid) before PayHere live mode.

## Alternatives considered

No alternatives were recorded in the source plan; the A/B/C hosting options are choices still to be made, not rejected alternatives.

## Consequences and trade-offs accepted

More environment plumbing and a container-first workflow. The host decision is deferred, so deployment automation is not complete. Vercel Hobby is non-commercial, free backend tiers tend to sleep, and always-on hosting is roughly a few dollars a month; these findings need checking again.

## Revisit when

Revisit before PayHere live mode, because the final domain and merchant setup depend on the host.

## Related

[0008](0008-postgres-everywhere.md), [0009](0009-separate-go-service.md), `ticketing/Dockerfile`
