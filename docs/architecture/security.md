# Security

- **Status:** implemented
- **Covers:** `src/payload.config.ts`, `src/collections/Users.ts`, `src/collections/access.ts`, `src/app/robots.ts`, `ticketing/internal/httpapi/guards.go`, `ticketing/internal/httpapi/internal.go`
- **Last checked:** 8de3320

The system has two perimeters: the CMS login and the ticketing service. Hiding a
URL is not one of them.

## Backstage

- The admin route is `/backstage`, configured in `src/payload.config.ts`.
- `robots.txt` disallows `/backstage`. This reduces indexing, not access.
- Users have `admin` or `member` roles. Only admins manage users. Content
  collections are publicly readable; Users is not.
- Login lockout is five attempts with a ten-minute lock.
- There is no public registration. The first admin is seeded from environment
  variables.
- No public page links to Backstage.

The admin is still internet-facing. The login, lockout, and role checks are the
real controls. An identity-aware proxy is possible future hardening.

## Ticketing service

- Internal routes require `Authorization: Bearer <INTERNAL_TOKEN>`, compared
  with `crypto/subtle.ConstantTimeCompare`.
- Order access uses `X-Order-Token`. Only its SHA-256 hash is stored; comparison
  is constant-time in the storage query. The token is returned once and kept in
  `sessionStorage`.
- QR tickets are HMAC-SHA256 signed with `QR_SECRET`. Verification is
  constant-time. The token is `v1.<ticket-uuid>.<signature>`.
- Scanner tokens are per-show SHA-256 hashes. The plaintext is returned once.
- Public rate limits are in-memory and per instance: 1,200 requests/minute per
  IP, plus 300 writes/minute for holds, checkout, and release.
- CORS is allowlisted only for `/v1/public/*`; no wildcard, no credentials.
- JSON bodies are limited to 1 MiB and must contain exactly one object.
- Request ids are accepted only if they match a bounded pattern; otherwise a
  random id is generated and returned in `X-Request-Id`.
- Panics are recovered and logged; responses do not leak stack traces.

## Secrets

Secrets live in environment variables, never in the repository:

- `PAYLOAD_SECRET` for Payload.
- `INTERNAL_TOKEN` for the Next.js-to-Go internal API.
- `QR_SECRET` for ticket signatures.
- `PAYHERE_MERCHANT_SECRET` for payment hashes.

Never commit credentials. `.env.example` contains placeholders only.

## Known gaps

- The `/backstage-tools` staff pages are planned, so staff actions currently use
  the internal API directly rather than a logged-in Next.js page.
- There is no `/ticket-api` proxy route yet.
- PayHere signature verification and notify handling are planned.
- Rate limiting is per process. The documented deployment assumes one Go
  instance; a horizontally scaled service would need a shared limiter.

## Related

- [0004](../decisions/0004-admin-at-backstage.md)
- [0017](../decisions/0017-order-access-token.md)
- [0019](../decisions/0019-qr-tickets.md)
- [API contract](api-contract.md)
