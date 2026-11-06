# 0004 — Admin hidden at /backstage

- **Status:** accepted
- **Covers:** `src/payload.config.ts`, `src/app/(payload)/`, `src/collections/Users.ts`, `src/app/robots.ts`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0017](0017-order-access-token.md), [0030](0030-access-control.md), `src/app/robots.ts`

## Context

The default Payload admin path is well known and constantly probed. The group still needs a usable login for members who work from changing networks.

## Decision

Move the admin route to `/backstage`, add no public links, disallow it in `robots.txt`, enable login lockout, and use admin/member roles. Hiding is not security; the login and access checks are.

## Alternatives considered

- **Default `/admin`** — constantly probed.
- **A secret random path** — false security and members forget it.
- **An IP allowlist** — members' IPs change.
- **Cloudflare Access** — possible later hardening, but extra setup now.

## Consequences and trade-offs accepted

The admin is less noisy in logs, but the path is not a security boundary. Accounts remain the real control.

## Revisit when

Revisit if public exposure or login abuse justifies an identity-aware proxy.

## Related

[0017](0017-order-access-token.md), [0030](0030-access-control.md), `src/app/robots.ts`
