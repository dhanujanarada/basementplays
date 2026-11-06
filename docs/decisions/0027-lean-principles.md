# 0027 — Lean principles and the parity guardrail

- **Status:** accepted
- **Covers:** `package.json`, `tsconfig.json`, `docs/ui-parity.md`, `src/`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0025](0025-frontend-first-mock.md), [0028](0028-development-workflow.md), `docs/ui-parity.md`

## Context

The revamp touched a working UI. Cleanup without a baseline could silently remove behaviour, and duplicated code was already spreading across designs.

## Decision

Derive instead of store, use one component instead of copies, and delete replaced code. Use knip, depcheck, and a one-time bundle analysis. Keep `docs/ui-parity.md` as the no-regression checklist.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

Smaller surfaces and measurable bundle changes. Cleanup takes time and requires the parity checklist to be maintained.

## Revisit when

Revisit if the checklist becomes too broad to run for routine changes.

## Related

[0025](0025-frontend-first-mock.md), [0028](0028-development-workflow.md), `docs/ui-parity.md`
