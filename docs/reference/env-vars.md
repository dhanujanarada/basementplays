# Environment variables

- **Status:** implemented
- **Covers:** `.env.example`, `ticketing/.env.example`, `src/lib/ticketing/http.ts`, `next.config.ts`, `ticketing/internal/config/config.go`
- **Last checked:** 8de3320

Values below are generated from the two example files and the current code. Do
not copy real credentials into documentation.

## Site (`.env.example`)

| Variable | Required | Current use |
| --- | --- | --- |
| `DATABASE_URL` | yes | Payload database URL. The example uses `file:./payload.db` (SQLite). |
| `PAYLOAD_SECRET` | yes | Payload signing secret. Use a long random value. |
| `SEED_ADMIN_EMAIL` | for seeding | First admin email. |
| `SEED_ADMIN_PASSWORD` | for seeding | First admin password. |
| `SEED_MEMBER_EMAIL` | optional | Member account for lockout/role testing. |
| `SEED_MEMBER_PASSWORD` | optional | Member password. |
| `NEXT_PUBLIC_TICKET_API_MODE` | optional | `mock` (default) or `http`. The example comment says HTTP is not wired; the client module now exists, but the `/ticket-api` proxy route does not. |
| `NEXT_PUBLIC_SITE_URL` | optional | Absolute site URL for QR links when request headers do not carry the host. Commented out in the example. |
| `NEXT_PUBLIC_TICKET_API_BASE` | optional, not in the example | Base path for the HTTP client; defaults to `/ticket-api`. |

`NEXT_PUBLIC_*` values are inlined at build time. Rebuild after changing them.

## Ticketing service (`ticketing/.env.example`)

| Variable | Default / example | Current use |
| --- | --- | --- |
| `DATABASE_URL` | `postgres:///ticketing_dev?host=/var/run/postgresql` | Postgres DSN. Required. |
| `PORT` | `8080` | HTTP listen port. |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error`. |
| `INTERNAL_TOKEN` | placeholder | Bearer token for `/v1/internal/*`; minimum 16 characters. |
| `QR_SECRET` | placeholder | HMAC key for ticket QR tokens; minimum 16 characters. |
| `PUBLIC_BASE_URL` | `http://localhost:3000` | Origin used to build `/t/<token>` links. |
| `HOLD_MINUTES` | `10` | Initial hold lifetime. |
| `CHECKOUT_HOLD_MINUTES` | `15` | Hold extension after PayHere checkout starts. |
| `MANUAL_HOLD_HOURS` | `24` | Bank-transfer / manual order hold. |
| `MAX_SEATS_PER_ORDER` | `6` | Maximum seats in one order. |
| `MAX_PLAN_SEATS` | `500` | Current code's plan seat limit. The revamp target calls this `MAX_SHOW_SEATS`; the code still uses the old name. |
| `CORS_ALLOWED_ORIGINS` | empty | Exact origins for `/v1/public/*`; empty disables CORS. |
| `PAYHERE_MERCHANT_ID` | empty | Planned PayHere merchant id. |
| `PAYHERE_MERCHANT_SECRET` | empty | Planned PayHere merchant secret. |
| `PAYHERE_SANDBOX` | `true` | Refuse live mode unless explicitly allowed. |
| `PAYHERE_ALLOW_LIVE` | `false` | Must be true to start with sandbox disabled. |
| `RUN_MIGRATIONS_ON_START` | `true` | Apply migrations during `serve`. |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown budget. |

The Go service fails fast when a required value is missing or malformed.
`PAYHERE_SANDBOX=false` without `PAYHERE_ALLOW_LIVE=true` is a startup error.

## Notes

- The site example still says SQLite, matching the current Payload adapter.
  Decision [0008](../decisions/0008-postgres-everywhere.md) says Postgres
  everywhere; that switch is not complete.
- The ticketing example and code still use `MAX_PLAN_SEATS`. The revamp target
  removes plans; this name will change when the Go model changes.
- No secrets are listed here. Use the placeholder files, then fill a local
  `.env` that is never committed.

## Related

- [Commands](commands.md)
- [Hosting and portability](../architecture/hosting-and-portability.md)
- [Security](../architecture/security.md)
