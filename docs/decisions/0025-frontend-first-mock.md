# 0025 — Frontend-first development with a mock ticket API

- **Status:** accepted
- **Covers:** `src/lib/ticketing/types.ts`, `src/lib/ticketing/mock.ts`, `src/lib/ticketing/client.ts`
- **Last checked:** 8de3320
- **Implementation:** done
- **Date:** 2026-11
- **Related:** [0009](0009-separate-go-service.md), [0013](0013-api-routing.md), `src/lib/ticketing/types.ts`

## Context

The booking UI needed to be designed and demonstrated before the Go service was deployed. The production contract still needs to be stable.

## Decision

Develop the frontend against a mock client behind one interface. Treat the TypeScript types file as the contract the Go API must match. Keep the mock behind a dynamic import so it is not in first-load or HTTP-mode production bundles.

## Alternatives considered

- No alternatives were recorded in the source plan.

## Consequences and trade-offs accepted

Fast UI work and demoable flows without Go. The mock and HTTP clients can drift unless the types stay authoritative.

## Revisit when

Revisit when the HTTP client is the default and the mock becomes a test fixture only.

## Related

[0009](0009-separate-go-service.md), [0013](0013-api-routing.md), `src/lib/ticketing/types.ts`
