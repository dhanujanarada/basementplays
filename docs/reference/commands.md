# Commands

- **Status:** implemented
- **Covers:** `package.json`, `ticketing/Makefile`, `scripts/check-docs.sh`
- **Last checked:** 8de3320

Run site commands from the repository root, and ticketing commands from
`ticketing/`. The two services have separate environment files.

## Site commands

| Command | What it does |
| --- | --- |
| `npm run dev` | Next.js dev server with the mock ticket client by default. |
| `npm run devsafe` | Remove `.next` first, then start dev. Use after an interrupted dev process. |
| `npm run build` | Production build. |
| `npm start` | Start the production build. |
| `npm run lint` | ESLint. |
| `npm run payload` | Payload CLI. |
| `npm run generate:types` | Regenerate `src/payload-types.ts` from the Payload config. |
| `npm run generate:importmap` | Regenerate the Payload admin import map. |
| `npm run seed` | Admin users, Site Settings, and Dear Daughter. |
| `npm run seed:admin` | Admin and optional member accounts from `.env`. |
| `npm run seed:site-settings` | Placeholder Site Settings and bank details. |
| `npm run seed:dear-daughter` | Dear Daughter event, shows, tier, gallery, and media. |
| `npm run seed:demo` | Fictional PlainDesign event for local development. |

Seeding is re-runnable by slug, but the scripts do not delete existing content.
Use a fresh database when the schema or seed shape changes.

## Ticketing commands

| Command | What it does |
| --- | --- |
| `make help` | List targets. |
| `make dev` | Load `ticketing/.env` and run the API. |
| `make migrate` | Apply pending goose migrations. |
| `make seed` | Create the current placeholder sample seat plan. |
| `make test` | Unit tests. |
| `make test-race` | Unit tests with the race detector. |
| `make test-integration` | Postgres integration tests tagged `integration`. |
| `make test-integration-race` | Postgres integration tests under `-race`. |
| `make lint` | `go vet` plus a gofmt check. |
| `make fmt` | Format Go source. |
| `make tidy` | `go mod tidy`. |
| `make db-create` | Create `ticketing_dev` if missing. |
| `make db-drop` | Drop `ticketing_dev`. |
| `make psql` | Open psql on `ticketing_dev`. |
| `make docker-build` | Build the Go container image with podman. |

Integration tests use `TEST_DATABASE_URL`, falling back to `DATABASE_URL`, and
skip when neither is set.

## Documentation command

| Command | What it does |
| --- | --- |
| `./scripts/check-docs.sh` | Check relative Markdown links and every `Covers:` path. |

## Related

- [Environment variables](env-vars.md)
- [Local setup](../guides/local-setup.md)
- [Operations](../guides/operations.md)
