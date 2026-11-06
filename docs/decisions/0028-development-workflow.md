# 0028 — Development workflow: specs, plan mode, step gates

- **Status:** accepted
- **Covers:** `docs/`, `AGENTS.md`, `TICKETING.md`, `DESIGN.md`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0027](0027-lean-principles.md), `docs/README.md`

## Context

Large changes were landing without enough shared context, and long agent sessions drifted from the written design.

## Decision

Write the spec first. Use opencode plan mode before build mode. Break work into stages, stop after each, give verification steps, and wait for `next`. Tag the repository before a large refactor.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

More up-front writing and deliberate pauses. Changes are easier to review and roll back.

## Revisit when

Revisit if the step gates slow down small fixes; keep the gates for structural work.

## Related

[0027](0027-lean-principles.md), `docs/README.md`
