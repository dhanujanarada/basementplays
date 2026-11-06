# Theatre Group Website + CMS — Design Doc

## 1. Goal
A public website for a theatre group with a built-in CMS. Each production gets its own
"sub-site" style page with a custom look, show dates, cast/crew, and a photo gallery.
The admin panel manages content and one Events editor that also holds each event's show
dates, ticketing settings, and bank-transfer details.

The admin panel is called **Backstage**. "Backstage" refers only to the admin panel —
never to public photo sections, gallery categories, or content types. Backstage holds
**content only**; each production's presentation lives in its own code component under
`src/designs/<slug>/`. Ticketing and payments are specified in `TICKETING.md`.

**Immediate target:** a demo-ready basic site (listing + event pages + Backstage) with
the mock-backed booking flow.

## 2. Tech Stack
- Next.js (App Router) + TypeScript
- Payload CMS 3 running inside the same Next.js app (admin panel at `/backstage`)
- Database: SQLite for local dev (via Payload's sqlite adapter, selected by
  `DATABASE_URL`); Postgres later for production
- Tailwind CSS, `next/font`, `next/image`
- Framer Motion (sparingly), `yet-another-react-lightbox` for galleries

**Fallback:** if Payload setup stalls for over an hour, replace `src/lib/payload.ts`
with a loader that reads local JSON files matching the schema below. Nothing else in
the frontend should change. Keep ALL CMS access inside `src/lib/payload.ts`.

## 3. Core Concept
One generic route renders every production from CMS-managed content. Presentation is
chosen in code by the production's slug:

1. **Custom design:** if `src/designs/<slug>/` exports a design component and it is
   registered in `src/designs/index.ts`, that component renders the production.
2. **Fallback:** otherwise `PlainDesign` renders a complete, neutral page with every
   content section.

Content is always structured and editable in Backstage. Backstage does **not** store
colors, fonts, layout keys, or other presentation settings.

## 4. Data Model (Payload collections and globals)

### Events (collection slug `productions`)
One production/event. The admin label is **Events**; the slug stays `productions`.
Admin UI tabs: Details, Shows & tickets, People, Gallery.

**Details**
- `title` (text, required)
- `slug` (text, unique, auto-generated from title)
- `tagline` (text)
- `synopsis` (rich text)
- `presentedBy` (text, default "A Basement production")
- `writtenBy` (text)
- `directedBy` (text)
- `poster` (upload -> Media, required)
- `heroImage` (upload -> Media, optional; kept from the existing design)

**Shows & tickets**
- `venue` (text): default venue for all this event's shows.
- `shows` (array):
  - `startsAt` (date+time, required)
  - `venue` (text, optional override)
  - `cancelled` (checkbox)
- `ticketing` (group):
  - `paymentMode` (select: bank_transfer [default], online, off)
  - `selectionMode` (select: quantity [default], seats)
  - `stagePosition` (select: front [default], back, none)
  - `numbering` (checkbox; shown only in seats mode)
  - `columns` (number, default 10; shown only in seats mode)
  - `tiers` (array): `name` (text, required), `price` (number, whole rupees,
    required), `count` (number, required, min 1). At least one tier.

**People**
- `cast` (array): `name`, `role`, `photo` (optional upload)
- `crew` (array): `name`, `role`
- `partners` (array): `name` (required), `role`, `logo` (optional upload)

**Gallery**
- `gallery` (array): `image` (upload -> Media, required), `caption` (text),
  `category` (select: On stage, Behind the scenes)

Rules:
- Show array row ids are the show ids used in `?show=` and in the Go service. Tier
  array row ids are the tier ids. Both are stable across rename and reorder.
- Do not add `ticketingEnabled`, `seatPlanKey`, `ticketUrl`, `price` text, `seatsNote`,
  or a manual `status` field. Ticketing state and statuses are derived:
  - active iff `paymentMode != off` and at least one tier has `price > 0` and
    `count > 0`;
  - cancelled from the checkbox, past from the date, sold out/few left from Go
    availability when ticketing is active.
- There is no `theme` group, no `layoutKey`, no `backstage` array, no testimonials.
  Do not add them back.

### Media
Payload upload collection with required `alt` text. Generate resized variants
(thumbnail ~400px, card ~800px, full ~1920px).

### Users
Payload auth collection with roles: `admin`, `member`.
- Only admins can create, update, or delete users.
- Members can edit Events, Media, and Site Settings, but cannot manage users.
- No public registration.
- Login lockout via `maxLoginAttempts` and `lockTime`.
- The first admin user is seeded from environment variables (or Payload's first-user
  flow); credentials are never hard-coded.

### Site Settings (global)
Publicly readable singleton, sidebar label "Site Settings":
- `groupName` (text, required)
- `tagline` (text)
- `aboutText` (rich text)
- `aboutImage` (optional upload -> Media)
- `email` (email)
- `phone` (text)
- `whatsappNumber` (text, international format, digits only) — used by the fallback
- `locationText` (text)
- `mapsUrl` (text)
- `socialLinks` (array): `platform` (select: instagram, facebook, youtube, tiktok,
  x, other), `url` (text, required)
- `bankTransfer` (group):
  - `bankName` (text)
  - `accountName` (text)
  - `accountNumber` (text)
  - `branch` (text)
  - `instructions` (textarea)

### Access control
- Public read: Events, Media, Site Settings.
- Never publicly readable: Users.
- Create/update/delete content: any logged-in user (admin or member).
- User management: admins only.
- Admin panel route: `/backstage`. `/admin` returns 404.
- `robots.txt` disallows `/backstage`.
- The public site never links to Backstage.

Admin navigation: **Events**, **Media**, **Users**, **Site Settings**, in that order.
Nothing else.

## 5. Folder Structure
```
src/
  app/
    (frontend)/
      layout.tsx              # global header, footer, fonts
      page.tsx                # home
      events/page.tsx         # listing
      events/[slug]/page.tsx  # resolves slug -> design component
      about/page.tsx
      contact/page.tsx
    (payload)/backstage/...   # Payload admin panel
    robots.ts                 # disallow /backstage
  collections/
    Productions.ts          # admin label "Events"; slug stays productions
    Media.ts  Users.ts
    access.ts  hooks/slug.ts
  globals/
    SiteSettings.ts
  designs/
    index.ts                  # registry: production slug -> { Page, Shell, Booking? }
    PlainDesign.tsx           # neutral fallback (PlainPage + PlainShell)
    dear-daughter/
      DearDaughterDesign.tsx  # DearDaughterPage + DearDaughterShell
      dear-daughter.css       # design-owned colors and styles
  components/
    ui/                       # Button, Badge, Container
    EventCard.tsx
    SiteHeader.tsx
    SiteFooter.tsx            # reusable themed footer
    ticketing/                # SeasonBooking, BookingFlow, SeatMap, confirmation, ticket
  lib/
    payload.ts                # ONLY place that talks to the CMS; returns EventView
    event-view.ts             # EventView types + derived show status/active rules
    types.ts
    format.ts
    ticketing/                # public contract, mock client, HTTP client
  payload.config.ts
```
Rule: each production design is self-contained in `src/designs/<slug>/` and owns its
layout, styling, and assets. Designs may reuse shared `components/ui` pieces. Backstage
stores content only.

## 6. Key Implementation Patterns

Design resolution in `events/[slug]/page.tsx`:
```tsx
const event = await getEventView(slug)
if (!event) notFound()
const design = designs[event.slug] ?? plainDesign
const { Page, Shell } = design

return (
  <Shell event={event}>
    <Page event={event} />
  </Shell>
)
```

Registry in `designs/index.ts`:
```ts
export const designs: Partial<Record<string, ProductionDesign>> = {
  'dear-daughter': {
    Page: DearDaughterPage,
    Shell: DearDaughterShell,
  },
}

export const plainDesign: ProductionDesign = {
  Page: PlainPage,
  Shell: PlainShell,
}
```

Colors, fonts, and layout live in the design's `Shell` and its CSS, not in Payload.
Each custom design defines its own CSS variables on its root element.

### Adding a new production

- [ ] Create `src/designs/<slug>/<Name>Design.tsx` and its CSS file.
- [ ] Export a `Page` that receives the normalized `EventView` and a `Shell` that
      renders the root element, CSS variables, fonts, textures, and themed
      `<SiteFooter />`.
- [ ] Optionally export a `Booking` layout; otherwise the shared `DefaultBooking`
      renders inside the design's `Shell`.
- [ ] Register `{ Page, Shell, Booking? }` in `src/designs/index.ts`. Designs not
      registered fall back to `plainDesign`.
- [ ] Use the shared `SeasonBooking` for the show selector and the single booking card;
      pass the design's existing card markup as a render prop. The card is the only
      reserve action.
- [ ] Define the selection variables in `Shell`: `--seat-available`,
      `--seat-selected`, `--seat-held`, `--seat-sold`, `--seat-blocked`,
      `--seat-stroke`, `--seat-label`, `--stage-fill`.
- [ ] Use `bookingHref`/`SeasonBooking` for the reserve link. Never hard-code a ticket
      URL or booking route in a design.
- [ ] Check the production page, booking page, confirmation page, ticket page, and
      mobile sticky bar at mobile and desktop widths, with `prefers-reduced-motion`
      enabled.

Data fetching in `lib/payload.ts`: use `payload.find` with `depth: 2` so media and
relations arrive in one query. Export `getAllEventViews()`, `getEventView(slug)`, and
`getSiteSettings()`. `EventView` flattens the event's shows, derives status and ticketing
active state, and is the ONLY data shape designs consume; designs never import a Run.
`getAllProductions`/`getRunsForProduction` and the `runs` relationship disappear with the
Runs collection.

The shared footer reads Site Settings and is styled with neutral fallbacks:
`--footer-bg` (`#111`), `--footer-text` (`#eee`), `--footer-accent` (`#c9a227`).
Production designs may override those variables on their root element.

Footer placement: the frontend layout renders `<SiteFooter />` for every page, but
`FooterGate` (a client component that checks `usePathname`) suppresses it on the
production page (`/events/[slug]`), the booking and confirmation routes, and the
ticket page (`/t/[token]`). Each of those routes renders inside the production's
`Shell`, which includes `<SiteFooter />` inside its own root element, so the
design's CSS variables theme the footer. This keeps exactly one footer per page.
Header and footer both read Site Settings (`groupName`, contact details, social
links) so content stays in Backstage.

## 7. Pages

**Home:** the next production, a short About teaser with a "Read more" link to
`/about`, and the upcoming shows.

**Events listing:** poster-card grid of **all** productions. Each card links to
`/events/<slug>` and shows title, next show date, and a status badge. Missing posters
are handled gracefully. Upcoming first.

**Event page (the sub-site):** a production's registered design renders the page. The
fallback `PlainDesign` shows every content field:
1. Hero with title, tagline, credits, and poster or hero image
2. Show dates as a selector: date, time, venue, status, price/tier display, and
   availability; past shows are dimmed and not selectable, cancelled shows are marked
3. Synopsis
4. Cast and crew
5. Gallery: masonry grid, captions, and a lightbox
6. Partners
7. Link back to all events

The event page uses the shared `SeasonBooking` client: the next upcoming show is
selected by default (or the most recent show with a "season ended" state), selecting a
show updates the card and `?show=`, and the booking card is the ONLY reserve action.
Hero CTAs and the mobile sticky bar are "Get tickets" anchors that scroll to the card.

**Booking, confirmation, and ticket pages:** `/events/[slug]/book`,
`/events/[slug]/book/confirmation`, and `/t/[token]` render inside the production's
`Shell` so they inherit its palette, fonts, and footer. A design may provide a custom
`Booking` component; otherwise the shared `DefaultBooking` is used. The event's
`paymentMode` decides the flow: bank transfer (details, reference, deadline, staff
marks paid), online (PayHere sandbox), or off (WhatsApp fallback). See `TICKETING.md`
for the data contract and the mock-backed demo flow.

**About:** full About content from Site Settings (`aboutText`, `aboutImage`).

**Contact:** contact details from Site Settings (email, WhatsApp button, location with
an "Open in Google Maps" link, social links). No contact form.

## 8. Design System
- Custom designs own their palette and typography in their own CSS. Do not put theme
  values in the CMS.
- `PlainDesign` fallback: dark, editorial, quiet. bg `#0d0d0f`, text `#f2efe9`, accent
  muted gold (`#c9a227`). Serif display + Inter body via `next/font`.
- Dear Daughter design: palette `--bg #4f7a56`, `--bg-deep #3d6247`,
  `--paper #f1ead9`, `--accent #9a4b38`, `--wood #a8794a`; footer variables
  `--footer-bg: var(--wood)`, `--footer-text: var(--paper)`,
  `--footer-accent: var(--paper)`. Fonts: Caveat Brush (display), Quicksand (body).
  Wallpaper `/public/textures/damask.png`, torn paper edge
  `/public/textures/torn-edge.png`; both need graceful fallbacks.
- Layout: max content width ~1200px, generous whitespace, full-bleed hero and gallery.
- Motion: subtle fade-up on scroll, slow hover zoom on posters. Respect
  `prefers-reduced-motion`.
- Mobile-first. Gallery: 2 columns on mobile, 3-4 on desktop. Use `next/image` with
  `sizes`. Lightbox must support swipe and keyboard.
- Accessibility: required alt text on media; check text/background contrast.

## 9. Scope for the Demo (build in this order)

### Step A: Schema and Backstage access
- Events collection (slug `productions`): Details, Shows & tickets, People, Gallery
  tabs as described in section 4; remove Runs and Props entirely.
- Remove manual ticketing fields (`ticketingEnabled`, `seatPlanKey`, `ticketUrl`,
  `price` text, `seatsNote`, manual status). Ticketing state is derived.
- Add the Site Settings global, including the `bankTransfer` group.
- Public read on Events, Media, Site Settings; never on Users.
- Move the admin panel to `/backstage`, title suffix " - Backstage", add `robots.txt`.
- Users roles `admin`/`member`, admin-only user management, login lockout.
- Backstage navigation: Events, Media, Users, Site Settings. Remove all public links
  to Backstage.

### Step B: Seed content
- Run all seeders with `npm run seed` (admin users, Site Settings, Dear Daughter).
- Seed "Dear Daughter," with credits, lorem ipsum synopsis, 4 cast, 3 crew, three
  clearly-marked placeholder future show dates, venue "The Black Bean Cafe, Kandy",
  one "General" tier at Rs. 2500 × 50, quantity mode, stage front, bank transfer,
  partners, and 8 captioned gallery images.
- Add a fictional dev-only event on `PlainDesign` if none exists.
- Seed Site Settings with placeholder contact, social, and bank-transfer values.
- Re-runnable without duplicates; print reminders for placeholder values.
- Generate placeholder artwork locally. Do not download anything.

### Step C: Routing, shared layout, About, Contact, footer
- Header nav: Home, Events, About, Contact.
- Footer used on every page, including inside production designs; all text from Site
  Settings; neutral CSS-variable fallbacks.
- About and Contact pages.
- Home: next production, About teaser, upcoming shows.
- Events listing and `[slug]` route as described in section 7.
- `PlainDesign` shows every content field.

### Step D: Dear Daughter design
- Self-contained folder with its own CSS and the palette/fonts from section 8.
- Hero with wallpaper and partner strip; cream section with torn edge, show selector,
  and booking card; wood divider; synopsis, cast/crew, gallery with captions and
  lightbox; mobile sticky **Get tickets** bar.
- The show-dates list is the selector and the booking card is the only reserve action.
- Respect `prefers-reduced-motion`.
- Texture files are supplied later by the user; the design must fall back gracefully.

Implementation: `src/designs/dear-daughter/DearDaughterDesign.tsx` loads Caveat Brush
and Quicksand through `next/font`, sets the palette on `.dd-page`, and renders
`<SiteFooter />` inside the design root. `dear-daughter.css` owns every rule. Optional
textures live at `public/textures/damask.png` and `public/textures/torn-edge.png`; both
are layered over CSS gradients/scallops so a missing file never breaks the page.
Partner logos are uploaded per partner in Backstage; when a partner has no logo the
design prints the partner name as text.

## 10. Out of Scope (do not build now)
- User accounts for the public
- Prop tracking (the Props collection is removed)
- Seat-plan templates, a seat-plan builder, or plan keys
- SEO polish and analytics

Ticketing (payments, QR tickets, the Go service) has its own spec in
`TICKETING.md`. The public booking flow currently runs against an in-browser mock.

## 11. Future Work (design so these are easy to add later)
- Real PayHere integration and the Go ticketing service (`TICKETING.md`).
- Archive page for past productions.
- Move SQLite to Postgres and deploy.

## 12. Conventions
- TypeScript strict mode; shared types in `src/lib/types.ts`.
- Server components by default; client components only for lightbox/animation and the
  booking flow.
- Keep each production design self-contained in `src/designs/<slug>/`.
- Designs consume only the normalized `EventView`; derive statuses instead of storing
  them.
- Keep one booking card per event and one reserve link source (`bookingHref`).
- Backstage holds content only: no colors, fonts, or layout keys in Payload.
- "Backstage" means the admin panel only. Never use it for photos, gallery categories,
  or public sections. Do not add testimonials.
- Do not access Payload outside `src/lib/payload.ts`.
