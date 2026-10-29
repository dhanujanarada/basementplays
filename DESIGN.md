# Theatre Group Website + CMS — Design Doc

## 1. Goal
A public website for a theatre group with a built-in CMS. Each production gets its own
"sub-site" style page with a custom look, show dates, cast/crew, a photo gallery, and
backstage photos. An admin panel manages content and per-production prop tracking.
The admin holds **content only**; each production's presentation lives in its own code
component under `src/designs/<slug>/`. Payments come later (out of scope for now).

**Immediate target:** a demo-ready basic site (listing + event pages + admin) by tomorrow.

## 2. Tech Stack
- Next.js (App Router) + TypeScript
- Payload CMS 3 running inside the same Next.js app (admin panel at /admin)
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

Content is always structured and editable in the admin. The admin does **not** store
colors, fonts, layout keys, or other presentation settings.

## 4. Data Model (Payload collections)

### Productions
A play (the creative work), which may be staged more than once.
- `title` (text, required)
- `slug` (text, unique, auto-generated from title)
- `tagline` (text)
- `synopsis` (rich text)
- `presentedBy` (text)
- `writtenBy` (text)
- `directedBy` (text)
- `poster` (upload -> Media, required)
- `heroImage` (upload -> Media)
- `partners` (optional array): `name` (text, required), `role` (text, e.g.
  "Location Partner" or "Promotional Partner"), `logo` (optional upload -> Media)
- `cast` (array): `name`, `role`, `photo` (optional upload)
- `crew` (array): `name`, `role`
- `gallery` (array of uploads -> Media): the polished show photos
- `backstage` (array): `image` (upload), `caption` (text): behind-the-scenes shots

Admin UI: organize fields into tabs: Details, Partners, People, Gallery, Backstage.
There is no theme group and no layout key.

### Runs
A season of a production, with its own dates and venue.
- `production` (relationship -> Productions, required)
- `label` (text, e.g. "2026 Season")
- `venue` (text)
- `shows` (array): `startsAt` (date+time), `status` (select: open, few-left,
  sold-out, cancelled), `price` (text), `seatsNote` (text, optional),
  `ticketUrl` (text, optional)

### Props
Per-production prop tracking, a backstage tool (not shown publicly).
- `production` (relationship -> Productions, required)
- `name` (text, required)
- `scene` (text)
- `status` (select: needed, sourced, ready, returned)
- `assignedTo` (text)
- `notes` (textarea)

### Media
Payload upload collection with required `alt` text. Generate resized variants
(thumbnail ~400px, card ~800px, full ~1920px).

### Users
Payload auth collection with roles: `admin`, `editor`. Editors can manage
productions/runs/media; props management visible to both.

Admin sidebar groups: **Content** (Productions, Runs, Media), **Backstage** (Props),
**System** (Users).

## 5. Folder Structure
```
src/
  app/
    (frontend)/
      layout.tsx              # global nav, footer, fonts
      page.tsx                # home
      events/page.tsx         # listing
      events/[slug]/page.tsx  # resolves slug -> design component
      archive/page.tsx        # past productions
      about/page.tsx
      contact/page.tsx
    (payload)/admin/...       # Payload-generated
  collections/
    Productions.ts  Runs.ts  Props.ts  Media.ts  Users.ts
  designs/
    index.ts                  # registry: production slug -> design component
    PlainDesign.tsx           # neutral fallback, shows all content
    dear-daughter/
      DearDaughterDesign.tsx  # self-contained production design
      dear-daughter.css       # design-owned colors and styles
  components/
    ui/                       # Button, Badge, Container
    EventCard.tsx
  lib/
    payload.ts                # ONLY place that talks to the CMS
    types.ts
    format.ts
  payload.config.ts
```
Rule: each production design is self-contained in `src/designs/<slug>/` and owns its
layout, styling, and assets. Designs may reuse shared `components/ui` pieces. The admin
stores content only.

## 6. Key Implementation Patterns

Design resolution in `events/[slug]/page.tsx`:
```tsx
const production = await getProduction(params.slug)
if (!production) notFound()
const Design = designs[production.slug] ?? PlainDesign
return <Design production={production} />
```

Registry in `designs/index.ts`:
```ts
export const designs = {
  'dear-daughter': DearDaughterDesign,
}
```

Colors, fonts, and layout live in the design component and its CSS, not in Payload.
Each custom design defines its own CSS variables on its root element.

Data fetching in `lib/payload.ts`: use `payload.find` with `depth: 2` so media and
relations arrive in one query. Export `getProduction(slug)`, `getAllProductions()`, and
`getRunsForProduction(id)`.

## 7. Pages

**Home:** hero for the next production, short "about us" blurb, upcoming shows row.

**Events listing:** poster-card grid of **all** productions. Each card links to
`/events/<slug>` and shows title, next show date, and a status badge. Missing posters
are handled gracefully. Upcoming first.

**Event page (the sub-site):** a production's registered design renders the page. The
fallback `PlainDesign` shows the full content:
1. Hero with title, tagline, credits, and poster or hero image
2. Show dates: date, time, venue, status, price, seats note, and ticket link
3. Synopsis
4. Cast and crew
5. Gallery: masonry grid + lightbox
6. Backstage: captioned photos, visually distinct from the gallery
7. Link back to all events

All runs of a recurring production appear on the same production page, newest run first.

**Archive:** past productions as a poster grid.

## 8. Design System
- Custom designs own their palette and typography in their own CSS. Do not put theme
  values in the CMS.
- `PlainDesign` fallback: dark, editorial, quiet. bg `#0d0d0f`, text `#f2efe9`, accent
  muted gold (`#c9a227`). Serif display + Inter body via `next/font`.
- Layout: max content width ~1200px, generous whitespace, full-bleed hero and gallery.
- Motion: subtle fade-up on scroll, slow hover zoom on posters. Respect
  `prefers-reduced-motion`.
- Mobile-first. Gallery: 2 columns on mobile, 3-4 on desktop. Use `next/image` with
  `sizes`. Lightbox must support swipe and keyboard.
- Accessibility: required alt text on media; check text/background contrast.

## 9. Scope for the Demo (build in this order)
1. Scaffold Next.js + Payload + Tailwind, fonts, global layout (nav/footer).
2. Create the collections in section 4; confirm /admin works with a seeded admin user.
3. Seed the real production "Dear Daughter," with its run, show, partners, gallery, and
   backstage images. Generate placeholder images locally if real ones are unavailable.
4. Build `lib/payload.ts`, the events listing, the `[slug]` route, and the designs registry.
5. Build `PlainDesign` with every content section.
6. Build the `dear-daughter` design in its own folder.
7. Optional: a second fictional production that uses `PlainDesign`.

### Dear Daughter seed
- `npm run seed:dear-daughter` seeds the real production; `npm run seed` runs the
  admin seed first and then this one. The script is idempotent and creates no duplicate
  production, run, or media records.
- Placeholder artwork is generated locally with `sharp` from gradients and text. No
  external or copyrighted assets are downloaded.
- The seeded show date is a **placeholder** until the real date is set in the admin.
  The seed prints a loud reminder with the placeholder date.
- Partner logos are optional and seeded without images. Designs must show the partner
  name as text when no logo exists.
- The older fictional demo data remains available via `npm run seed:demo`.

## 10. Out of Scope (do not build now)
- Payments / checkout
- User accounts for the public
- Prop dashboard styling beyond the default admin UI
- Email, ticket QR codes, or booking logic
- SEO polish and analytics

## 11. Future Work (design so these are easy to add later)
- Payment gateway (likely PayHere for Sri Lanka; verify availability). Add a `Bookings`
  collection with status pending/paid; the gateway webhook marks it paid.
- Link "Book tickets" buttons to the booking flow instead of an external `ticketUrl`.
- Move SQLite to Postgres and deploy.

## 12. Conventions
- TypeScript strict mode; shared types in `src/lib/types.ts`.
- Server components by default; client components only for lightbox/animation.
- Keep each production design self-contained in `src/designs/<slug>/`.
- The admin holds content only: no colors, fonts, or layout keys in Payload.
- Do not access Payload outside `src/lib/payload.ts`.