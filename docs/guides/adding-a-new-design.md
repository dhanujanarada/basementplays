# Adding a new design

- **Status:** implemented
- **Covers:** `src/designs/`, `src/lib/types.ts`, `src/components/ticketing/SeasonBooking.tsx`, `docs/ui-parity.md`
- **Last checked:** 8de3320

A design is code, not CMS content. Each event gets a self-contained folder under
`src/designs/<slug>/`.

## Checklist

1. Create `src/designs/<slug>/<Name>Design.tsx` and a CSS file.
2. Export `Page` and `Shell`.
3. Optionally export `Booking`; otherwise `DefaultBooking` is used.
4. Register `{ Page, Shell, Booking? }` in `src/designs/index.ts` under the event
   slug.
5. Use the shared `SeasonBooking` for the show selector and the single booking
   card. Pass the design's card markup as `renderCard`.
6. Keep the card as the only reserve action. Hero and mobile links should scroll
   to `#booking`.
7. Put palette, fonts, textures, and layout in `Shell` and the design CSS. Do
   not add theme fields to Payload.
8. Define the selection variables on the shell root: `--seat-available`,
   `--seat-selected`, `--seat-held`, `--seat-sold`, `--seat-blocked`,
   `--seat-stroke`, `--seat-label`, `--stage-fill`.
9. Use `bookingHref`/`SeasonBooking`; never hard-code a ticket URL.
10. Check the event, booking, confirmation, and ticket pages at mobile and
    desktop widths.
11. Check `prefers-reduced-motion`.
12. Update `docs/ui-parity.md` if the design intentionally changes a behaviour.

## Contract

`Page` receives `{ event }`. `Shell` receives `{ event, children }`. `Booking`
receives `{ event, show }`. Only `EventView` is passed to designs; no design
imports Payload or reads a raw event row.

```tsx
// src/designs/index.ts — designs
'dear-daughter': {
  Page: DearDaughterPage,
  Shell: DearDaughterShell,
}
```

## Design-owned content

- Hero and poster treatment.
- Show selector presentation, using the shared state.
- Booking card presentation.
- Synopsis, cast/crew, partners, gallery, outro.
- Themed footer through `Shell`.

## Shared content

- Event data and derived statuses.
- Booking state, holds, orders, and tickets.
- Payment and confirmation logic.
- The footer component itself; the design only themes it.

## Testing

- `npm run lint`
- `npx tsc --noEmit`
- `npm run build`
- Event page with a next show, no upcoming show, cancelled show, and inactive
  ticketing.
- Booking page in quantity and seats modes.
- Confirmation and ticket pages.

## Related

- [Design contract](../decisions/0003-design-contract.md)
- [Event page booking](../decisions/0024-event-page-booking.md)
- [Site walkthrough](../code/site-walkthrough.md)
- [CSS patterns](../code/css-patterns.md)
