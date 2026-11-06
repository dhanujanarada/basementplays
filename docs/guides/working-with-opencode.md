# Working with opencode

- **Status:** implemented
- **Covers:** `docs/`, `AGENTS.md`, `TICKETING.md`, `DESIGN.md`
- **Last checked:** 8de3320

This project is developed with opencode using a written spec, a plan step, a
build step, and explicit stop points.

## Plan mode first

Use plan mode when the work changes architecture, data, or more than one route.
Ask the agent to read the existing specs and the actual code first. The plan
should name files, trade-offs, and verification steps.

Use build mode for a step that is already planned. Keep one stage in one
conversation where possible so the agent retains context.

## Step gating

Large work is split into stages. After each stage:

1. the agent stops;
2. it says exactly what to review and which URLs or commands to run;
3. it waits for **`next`**.

Do not skip the review. If a stage is wrong, fix it before the next one. This
keeps a bad data-model or UI decision from spreading.

## Tags before refactors

Before a large refactor, commit the current state and tag it. The ticketing
revamp used:

```sh
git tag pre-revamp
```

The tag is the baseline for the UI parity checklist and for comparing bundle
sizes.

## Specs and decisions

- `TICKETING.md` is the long-form ticketing specification.
- `DESIGN.md` is the site and CMS design.
- `docs/decisions/` records why choices were made.

When a decision changes, add a new record that supersedes the old one. Do not
rewrite history. Update the index.

## Documentation rule

When a build step finishes, update the matching documentation in the same
commit. When code and an older spec disagree, document the code and record the
disagreement in `docs/DISCREPANCIES.md` until the spec is reconciled.

Run `scripts/check-docs.sh` before committing docs.

## Verification

The usual gate is:

```sh
npx tsc --noEmit
npm run lint
npm run build
```

For Go:

```sh
go vet ./...
go test ./...
```

The agent must not claim a visual check it did not perform. Ask for exact URLs,
clicks, and expected results.

## Related

- [Documentation conventions](../README.md)
- [Decision records](../decisions/README.md)
- [Local setup](local-setup.md)
