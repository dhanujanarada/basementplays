# Operations

- **Status:** partial
- **Covers:** `ticketing/Makefile`, `ticketing/Dockerfile`, `.env.example`, `ticketing/.env.example`, `package.json`
- **Last checked:** 8de3320

This is the on-call checklist. No deployment automation exists yet; the
`deploy/` directory, backup script, and restore test are planned.

## Backups

### Site content

The current local adapter stores Payload content in `payload.db` (SQLite). For a
consistent copy, stop the site process and copy the file. For production, move
Payload to Postgres as decision [0008](../decisions/0008-postgres-everywhere.md)
requires, then use `pg_dump`.

Media files live in `media/` unless media storage is moved to S3 later.

### Ticketing

```sh
pg_dump "$DATABASE_URL" > ticketing-$(date +%F).sql
```

Keep the dump and the media backup together. A database restore without media
leaves broken images.

### Restore

```sh
createdb -T template0 ticketing_restore
psql ticketing_restore < ticketing-YYYY-MM-DD.sql
```

Restore media to the matching path or bucket. Start the service, check
`/readyz`, then open one ticket and one event page.

The planned `deploy/backup.sh` should automate the dump, media copy, retention,
and a restore test.

## Deploy checklist

- [ ] Hosting setup A, B, or C chosen and documented.
- [ ] Domain and HTTPS configured.
- [ ] `PAYLOAD_SECRET`, `INTERNAL_TOKEN`, `QR_SECRET` are long random values.
- [ ] PayHere sandbox credentials are present; live mode is still false.
- [ ] `DATABASE_URL` points at the intended database with the right `sslmode`.
- [ ] Migrations applied.
- [ ] `/healthz` and `/readyz` pass.
- [ ] `/ticket-api` is routed to Go, or the site is explicitly running mock mode.
- [ ] `/backstage` is reachable and login works.
- [ ] A test booking completes and a ticket QR opens.
- [ ] Backups run and a restore has been tested.
- [ ] Privacy Policy page exists before taking real bookings.
- [ ] `robots.txt` still disallows `/backstage`.

## Troubleshooting

### Site returns 404 for every page

Usually a stale `.next` cache after an interrupted dev process.

```sh
npm run devsafe
```

If using a production build, rebuild with `npm run build` and start with
`npm start`.

### Backstage login fails

- Check the account exists.
- Five failed attempts lock the account for ten minutes.
- Confirm `PAYLOAD_SECRET` did not change unexpectedly; sessions depend on it.

### Booking says the service is unavailable

- In mock mode, check the browser session is not blocked from `sessionStorage`.
- In HTTP mode, check `/ticket-api` routing and Go `/readyz`.
- Check `NEXT_PUBLIC_TICKET_API_MODE` was set at build time, then rebuild.

### Go will not start

- Check `INTERNAL_TOKEN` and `QR_SECRET` are at least 16 characters.
- Check `DATABASE_URL` is a Postgres URL or libpq keyword string.
- Check `PAYHERE_SANDBOX=false` is not set without `PAYHERE_ALLOW_LIVE=true`.
- Check migrations with `make migrate`.

### Holds do not expire

- Check the sweeper logs.
- Correctness does not depend on the sweeper; the claim query treats expired
  holds as available.
- Check the database clock and `expires_at`.

### Ticket page 404

- The token is invalid, tampered, or the ticket no longer exists.
- Check `QR_SECRET` has not changed since the ticket was issued.

### Seating changes are rejected

The current Go service returns `show_in_use` or `plan_in_use` when a plan change
would affect existing sales or non-available seats. The revamp target uses
`seating_locked` for tier changes. This is deliberate. Do not bypass it. The site
warning is planned.

## Related

- [Local setup](local-setup.md)
- [Security](../architecture/security.md)
- [Hosting and portability](../architecture/hosting-and-portability.md)
- [Payments](../architecture/payments.md)
