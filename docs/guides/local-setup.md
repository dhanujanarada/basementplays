# Local setup

- **Status:** implemented
- **Covers:** `.env.example`, `ticketing/.env.example`, `package.json`, `ticketing/Makefile`, `ticketing/README.md`, `scripts/check-docs.sh`
- **Last checked:** 8de3320

There are two local paths: the **site demo** with the mock ticket client, and the
**full Go service** with Postgres. Most UI work needs only the first.

## Requirements

- Node.js `^22.23.4` or `>=24.22.1`.
- npm.
- For the Go service only: Go 1.26 or newer and PostgreSQL 16 or newer.
- `make`, `psql`, `createdb`, `dropdb` for the Go targets.

## Site demo (mock ticketing)

From the repository root:

```sh
npm install
cp .env.example .env
# edit PAYLOAD_SECRET and the seed credentials
npm run dev
```

Open `http://localhost:3000`. The first start creates the Payload schema in
`payload.db`.

If this is a fresh database, seed it in another terminal:

```sh
npm run seed
npm run seed:demo
```

`npm run seed` creates the admin/member users, Site Settings, and Dear Daughter.
`npm run seed:demo` adds the fictional PlainDesign event.

The default ticketing mode is `mock`. Holds, orders, and buyer details live in
`sessionStorage`, so book and confirm in the same tab.

If the site returns 404 after an interrupted dev process:

```sh
npm run devsafe
```

That removes `.next` and starts clean.

## Reset the site database

Stop the dev server, then remove `payload.db` and start again. This deletes all
local content. Re-run the seeds afterward. Do not do this against a database you
care about.

## Go ticketing service

### Install Postgres on Nobara / Fedora / RHEL

```sh
sudo dnf install -y postgresql-server postgresql
sudo postgresql-setup --initdb
sudo systemctl enable --now postgresql
sudo -u postgres createuser -s "$USER"
```

### Configure and run

```sh
cd ticketing
cp .env.example .env
# edit INTERNAL_TOKEN, QR_SECRET, and PUBLIC_BASE_URL
make db-create
make dev
```

`make db-create` creates `ticketing_dev`. `make dev` applies migrations unless
`RUN_MIGRATIONS_ON_START=false`, then serves on port 8080.

Probes:

```sh
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
```

Seed the current placeholder seat plan:

```sh
make seed
```

### Tests

```sh
make test
make test-race
make test-integration
make test-integration-race
```

Integration tests need `TEST_DATABASE_URL` or `DATABASE_URL`.

## Full stack caveat

The site has an HTTP ticketing client, but the browser-facing `/ticket-api`
proxy route is not implemented. A full browser-to-Go setup needs a reverse proxy
or that route. Until then, use the mock client for UI work. See
[API routing](../decisions/0013-api-routing.md).

## Documentation check

```sh
./scripts/check-docs.sh
```

## Related

- [Commands](../reference/commands.md)
- [Environment variables](../reference/env-vars.md)
- [Operations](operations.md)
- [Working with opencode](working-with-opencode.md)
