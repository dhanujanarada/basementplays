# CSS patterns

- **Status:** implemented
- **Covers:** `src/app/(frontend)/globals.css`, `src/designs/`, `src/components/ticketing/ticketing.css`
- **Last checked:** 8de3320

The site uses Tailwind utilities for layout and plain CSS for each design's
identity. Shared ticketing UI is themed only through CSS custom properties.

## Custom properties

The neutral palette lives on `:root`. A design overrides the same names on its
root element, so shared UI inherits the design's colors without knowing which
design is active.

```css
/* src/designs/dear-daughter/dear-daughter.css — .dd-page */
--bg: #4f7a56;
--paper: #f1ead9;
--accent: #9a4b38;
--ink: #241f19;
--font-display: var(--dd-font-display);
```

The ticketing CSS uses fallbacks such as `var(--accent, #c9a227)`, so it still
renders when a design omits a variable.

## Shell theming

`Shell` owns the root element, font variables, textures, header, and footer.
Booking, confirmation, and ticket pages render inside it. This is why a ticket
page can look like the production without duplicating the booking layout.

## clamp()

`clamp(min, preferred, max)` gives fluid spacing and type without a pile of
breakpoints. It is used for section spacing, titles, and gaps.

```css
/* src/components/ticketing/ticketing.css — .tk */
padding-block: clamp(2.5rem, 6vw, 5rem);
```

## CSS grid

Grid handles the two-column season/booking layout, the booking summary, and the
seat map. `minmax(0, 1fr)` prevents long content from blowing out a column.

```css
/* src/components/ticketing/ticketing.css — .tk-layout */
grid-template-columns: minmax(0, 1fr);
/* at 900px: minmax(0, 1fr) 20rem */
```

The seat grid uses the configured column count as a custom property:

```css
grid-template-columns: repeat(var(--tk-columns, 10), minmax(2rem, 1fr));
```

## Torn edge and mask-image

The Dear Daughter torn edge is **not** an `mask-image` in the current code. It is
a layered background: an optional `/textures/torn-edge.png` plus a radial
gradient that paints the paper-colored scallop. The page still works when the
texture is missing.

```css
/* src/designs/dear-daughter/dear-daughter.css — .dd-torn */
background-image:
  url('/textures/torn-edge.png'),
  radial-gradient(circle at 50% 100%, var(--paper) 66%, transparent 68%);
```

`mask-image` appears on the neutral `.site-grid` texture, where it fades the
grid toward the bottom. If a future change moves the torn edge to a mask, this
page should be updated.

## Reduced motion

`prefers-reduced-motion` disables transitions, hover zoom, fade-up animation,
and smooth scrolling. Keep new motion behind the same guard.

```css
@media (prefers-reduced-motion: reduce) {
  .tk-seat, .tk-button, .tk-picker a { transition: none; }
}
```

## Related

- [Site walkthrough](site-walkthrough.md)
- [Design contract](../decisions/0003-design-contract.md)
- [UI parity checklist](../ui-parity.md)
