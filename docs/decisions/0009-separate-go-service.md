# 0009 — Separate Go ticketing service

- **Status:** accepted
- **Covers:** `ticketing/`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0010](0010-build-ticketing-ourselves.md), [0012](0012-repo-layout.md), `ticketing/cmd/ticketing/main.go`

## Context

Money and inventory need stronger isolation and transactional guarantees than the CMS content flow. Go also gives a small static binary and is a learning goal for the team.

## Decision

Run ticketing as a separate Go service with its own Postgres schema and API. The site calls it server-to-server and through a public ticket API.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

A second service to run, auth bridging, duplicated API types, and more deployment work. This was a conscious choice; TypeScript inside Next.js would have been simpler and cheaper.

## Revisit when

Revisit only if the operational overhead of the second service outweighs the isolation and learning benefits.

## Related

[0010](0010-build-ticketing-ourselves.md), [0012](0012-repo-layout.md), `ticketing/cmd/ticketing/main.go`
