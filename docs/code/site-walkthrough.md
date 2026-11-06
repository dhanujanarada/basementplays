# Site walkthrough

- **Status:** implemented
- **Covers:** `src/app/(frontend)/`, `src/designs/`, `src/components/`, `src/lib/payload.ts`, `src/lib/event-view.ts`, `src/lib/ticketing/`
- **Last checked:** 8de3320

The public site is the Next.js App Router application at the repository root.
Payload runs inside it. The booking and ticket routes are the only parts that
talk to the ticketing client.

## Routes

| Route | Shape | Main code |
| --- | --- | --- |
| `/` | server page | `src/app/(frontend)/page.tsx` |
| `/events` | server page | `src/app/(frontend)/events/page.tsx` |
| `/events/[slug]` | server page + design components | `events/[slug]/page.tsx`, `src/designs/` |
| `/events/[slug]/book` | server shell + client flow | `book/page.tsx`, `DefaultBooking`, `BookingFlow` |
| `/events/[slug]/book/confirmation` | server shell + client confirmation | `confirmation/page.tsx`, `OrderConfirmation` |
| `/t/[token]` | server shell + ticket view | `t/[token]/page.tsx`, `TicketView` |
| `/backstage` | Payload admin | `src/app/(payload)/` |

`/about` and `/contact` are content pages fed by Site Settings.

## Content path

`src/lib/payload.ts` is the only module that calls Payload. It returns
`EventView`, not raw records. `EventView` is built in `src/lib/event-view.ts`:
shows are flattened, venues are resolved, statuses are derived, and the
ticketing settings are attached.

```ts
// src/app/(frontend)/events/[slug]/page.tsx — EventPage
const design = designs[event.slug] ?? plainDesign
const { Page, Shell } = design

return (
  <Shell event={event}>
    <Page event={event} />
  </Shell>
)
```

The page chooses a design by slug. If no design is registered, `PlainDesign`
renders the neutral fallback.

## Design contract

A design exports `Page`, `Shell`, and optionally `Booking` from
`src/lib/types.ts`:

- `Page` owns the event content.
- `Shell` owns the root element, CSS variables, fonts, textures, header, and
  themed footer. It is reused by the event, booking, confirmation, and ticket
  routes.
- `Booking` is optional; `DefaultBooking` is the default layout.

Dear Daughter uses `DearDaughterSeason` and `PlainDesign` uses `PlainSeason`.
Both are client wrappers around the shared `SeasonBooking`.

## SeasonBooking

`SeasonBooking` is a client component. It owns:

- the selected show;
- the next-show fallback and the season-ended state;
- `?show=` on load and `history.replaceState` on change;
- the only reserve link (`bookingHref`).

It does not own the card markup. Each design passes `renderCard` and a children
render prop. The card remains design-specific, but the state is shared.

```tsx
// src/components/ticketing/SeasonBooking.tsx — SeasonBooking
const searchParams = useSearchParams()
const nextShow = useMemo(() => getNextShow(event), [event])
const fallbackShow = nextShow ?? latestShow(event)
const requestedShowID = searchParams.get('show')
```

Past shows are not selectable. Cancelled shows remain visible and marked. The
booking card is the only reserve action; hero and mobile-bar links scroll to it.

## Ticket client and mock

`src/lib/ticketing/client.ts` is a lazy proxy. It dynamically imports the mock
or HTTP implementation on first use, so the mock is not in the first-load route
bundles. The shared contract is `src/lib/ticketing/types.ts`.

The mock keeps holds, orders, and buyer details in `sessionStorage`. It supports
quantity and seats modes, stage positions, payment modes, and test scenarios
such as `sold-out` and `hold-expired`.

## Backstage integration

Backstage is the Payload admin at `/backstage`. It is not linked from the public
site. Fields are content and ticketing configuration; the event save hook that
would sync each show to Go is not implemented yet.

## Related

- [TypeScript, React, and Next.js](typescript-react-next.md)
- [Payload](payload.md)
- [CSS patterns](css-patterns.md)
- [Architecture overview](../architecture/overview.md)
