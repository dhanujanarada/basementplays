# Documentation

- **Status:** implemented
- **Covers:** `docs/`
- **Last checked:** 8de3320

This directory explains the Basement Plays site, the ticketing service, and the
decisions behind them. It is written for three readers: group members who run the
admin panel, developers who maintain the code, and a future maintainer returning after
a break.

## Reading order by role

**Group member / box office**

1. [Backstage handbook](guides/backstage-handbook.md) — create events and manage shows.
2. [Operations](guides/operations.md) — backups, deployment checks, common failures.

**Developer**

1. [Architecture overview](architecture/overview.md) — services and request flow.
2. [Data model](architecture/data-model.md) — Payload collections and Postgres schema.
3. [Site walkthrough](code/site-walkthrough.md) — the Next.js app tour.
4. [Ticketing walkthrough](code/ticketing-walkthrough.md) — the Go service tour.
5. [Commands](reference/commands.md) and [environment variables](reference/env-vars.md).
6. [Local setup](guides/local-setup.md) — run the stack locally.
7. [UI parity checklist](ui-parity.md) — the no-regression baseline for the revamp.

**Future me**

1. [Decision records](decisions/README.md) — why the system looks like this.
2. [Hosting and portability](architecture/hosting-and-portability.md) — deployment options.
3. [Code/spec discrepancies](DISCREPANCIES.md) — temporary list of places where the old specs and the code disagreed.
4. [Glossary](reference/glossary.md) — shared vocabulary.

## Conventions

Every document starts with this metadata block:

```markdown
- **Status:** implemented | partial | planned
- **Covers:** `src/`, `ticketing/`
- **Last checked:** <commit hash>
```

- Keep files short and specific. Split a document before it grows past roughly 250
  lines.
- Use plain English. Define a term at first use or link to
  `docs/reference/glossary.md`.
- Explain **why**, not only what. A list of fields is less useful than the reason the
  fields exist.
- Link instead of duplicating. If two pages need the same explanation, link to the
  source page and keep one copy authoritative.
- Code excerpts are at most about 20 lines. Give the file path and function or
  component name; do not use line numbers, which drift.
- Generate reference material from source files (`.env.example`, `Makefile`,
  `package.json`, `src/payload-types.ts`). Do not invent values.
- Never include secrets, API keys, passwords, or real credentials. Placeholders are
  fine.
- Decision records use a slightly different metadata block: their **Status** is
  `accepted`, `superseded by NNNN`, or `deferred`, and they also state
  **Implementation**, **Date**, **Covers**, **Last checked**, and **Related**.

## Keeping docs current

When finishing a build step, update the matching docs in the same commit. When a
decision changes, add a new decision record that supersedes the old one; do not edit
history in place. The decision index records which record wins.

## Structure

- `docs/decisions/` — why choices were made.
- `docs/architecture/` — services, data, APIs, security, hosting.
- `docs/code/` — code and language walkthroughs.
- `docs/guides/` — task-focused instructions.
- [Reference index](reference/README.md) — generated or lookup material.
- [Guides index](guides/README.md) — task-focused instructions.
- [docs/ui-parity.md](ui-parity.md) — visual and behavioural baseline from before the
  revamp.

Run `scripts/check-docs.sh` before committing documentation changes. It checks that
relative links resolve and that every path in a **Covers** line exists.
