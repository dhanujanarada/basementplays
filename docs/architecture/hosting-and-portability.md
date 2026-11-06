# Hosting and portability

- **Status:** planned
- **Covers:** `ticketing/Dockerfile`, `ticketing/Makefile`, `next.config.ts`, `.env.example`, `ticketing/.env.example`
- **Last checked:** 8de3320

The final host is not chosen. The repository is built so that the choice is a
configuration and deployment change, not an application rewrite. There is no
`deploy/` directory yet.

## Setups

| Setup | Shape | What changes |
| --- | --- | --- |
| A — one VPS | Next.js and Go on one machine; Caddy in front; Postgres local | one origin, proxy `/ticket-api` and `/backstage`; local media; local Postgres |
| B — separate hosts | Next.js host and Go host; managed Postgres | `TICKETING_INTERNAL_URL`, public ticket API URL, CORS, network access, separate TLS |
| C — hybrid | Next.js on a platform; Go on a VPS; managed Postgres | same as B plus object storage for media and a stable public site URL |

The decision is deferred until before PayHere live mode. Vercel Hobby is
non-commercial; free backend tiers tend to sleep; always-on hosting is roughly a
few dollars a month. These findings need checking again.

## Portability rules

1. All configuration is environment variables. No host-specific code or SDKs.
2. Each service can ship as a container. The Go Dockerfile is present; the Next
   standalone deployment is not configured yet.
3. The browser uses one same-origin ticket path, `/ticket-api`, routed by a
   reverse proxy or a runtime proxy route.
4. The Next.js server calls Go server-to-server with `INTERNAL_TOKEN`. Internal
   routes are never public.
5. CORS is only needed if the browser calls Go directly; the preferred setup is
   same-origin, so CORS can stay empty.
6. Postgres is standard. No provider-specific SQL.
7. Media is local today. The intended S3 switch is not implemented yet.
8. Token-based order access means no cookies and no sticky sessions. The service
   is stateless except for the in-memory rate limiter.
9. Backups are `pg_dump` plus the deployment's media store. The backup script is
   planned.

## What still has to be built

- `deploy/README.md` with A/B/C instructions.
- `deploy/backup.sh` and a tested restore.
- The `/ticket-api` runtime proxy or a documented Caddy rule.
- The Next.js standalone output setting if the chosen host uses containers.
- S3-compatible media storage if the host has no persistent disk.

## Related

- [0011](../decisions/0011-hosting-portability.md)
- [0012](../decisions/0012-repo-layout.md)
- [0013](../decisions/0013-api-routing.md)
- [Environment variables](../reference/env-vars.md)
