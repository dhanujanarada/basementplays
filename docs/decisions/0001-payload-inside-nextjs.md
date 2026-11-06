# 0001 — Payload CMS inside Next.js (one codebase)

- **Status:** accepted
- **Covers:** `src/payload.config.ts`, `src/app/(payload)/`, `src/collections/`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0002](0002-content-only-cms.md), [0012](0012-repo-layout.md), `src/payload.config.ts`

## Context

The group needs a public site, a gallery-heavy React frontend, and an admin panel that non-developers can use. Operating two applications would add deployment and authentication work before the site is useful.

## Decision

Run Payload CMS inside the same Next.js application. One process serves the public site, the admin panel, and the content API.

## Alternatives considered

- **Django** — two applications to run and a weaker fit for the React gallery frontend.
- **A JSON-file CMS** — no admin UI, so content editing would require code or hand-edited files.

## Consequences and trade-offs accepted

One codebase, one deployment, and one login. Payload upgrades and Next.js upgrades move together. The admin bundle is large, but it is isolated behind `/backstage`.

## Revisit when

Revisit if Payload blocks a required Next.js upgrade, or if the admin and public site need different release schedules.

## Related

[0002](0002-content-only-cms.md), [0012](0012-repo-layout.md), `src/payload.config.ts`
