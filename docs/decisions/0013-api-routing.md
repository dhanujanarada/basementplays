# 0013 — API routing through same-origin /ticket-api

- **Status:** accepted
- **Covers:** `src/lib/ticketing/http.ts`, `src/lib/ticketing/client.ts`, `ticketing/internal/httpapi/`
- **Last checked:** 8de3320
- **Implementation:** partial
- **Date:** 2026-11
- **Related:** [0011](0011-hosting-portability.md), [0017](0017-order-access-token.md), `src/lib/ticketing/http.ts`

## Context

The browser needs to reach the Go service without exposing internal endpoints or relying on cookies. Different hosting setups need the same URL.

## Decision

The browser calls `/ticket-api` on the same origin. A reverse proxy or runtime proxy forwards it to Go. Internal endpoints stay private, the Next.js server uses a service token, CORS is optional and allowlisted, and the ticketing API uses no cookies. `NEXT_PUBLIC_*` values are inlined at build time.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

One public URL shape across hosting setups. The HTTP client and base path exist, but the `/ticket-api` proxy route is not implemented yet, so the current mode is still mock.

## Revisit when

Revisit when the Go service is deployed and the runtime proxy or reverse-proxy route is added.

## Related

[0011](0011-hosting-portability.md), [0017](0017-order-access-token.md), `src/lib/ticketing/http.ts`
