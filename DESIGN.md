# Theatre Group Website + CMS — Design Doc

## 1. Goal
A public website for a theatre group with a built-in CMS. Each production gets its own
"sub-site" style page with a custom look, show dates, cast/crew, a photo gallery, and
backstage photos. An admin panel manages content and per-production prop tracking.
Payments come later (out of scope for now).

**Immediate target:** a demo-ready basic site (listing + event pages + admin) by tomorrow.

## 2. Tech Stack
- Next.js (App Router) + TypeScript
- Payload CMS 3 running inside the same Next.js app (admin panel at /admin)
- Database: SQLite for local dev (via Payload's sqlite adapter); Postgres later for production
- Tailwind CSS, `next/font`, `next/image`
- Framer Motion (sparingly), `yet-another-react-lightbox` for galleries

**Fallback:** if Payload setup stalls for over an hour, replace `src/lib/payload.ts`
with a loader that reads local JSON files matching the schema below. Nothing else in
the frontend should change. Keep ALL CMS access inside `src/lib/payload.ts`.

## 3. Core Concept
One generic route renders every production. Presentation is chosen per production:

1. **Default layout + CMS theme:** theme fields (colors, fonts) are stored in the CMS and
   injected as CSS variables. No code needed for a new show.
2. **Bespoke layout:** a production sets `layoutKey` to a custom React component, which
   receives the same data shape as the default layout and can look completely different.

Content is always structured and editable in the CMS. Only presentation can be custom.

## 4. Data Model (Payload collections)

### Productions
A play (the creative work), which may be staged more than once.
- `title` (text, required)
- `slug` (text, unique, auto-generated from title)
- `tagline` (text)
- `synopsis` (rich text)
- `poster` (upload -> Media, required)
- `heroImage` (upload -> Media)
- `layoutKey` (select; options match keys in the layout registry; default `default`)
- `theme` (group): `bg`, `text`, `accent` (hex colors), `displayFont` (select from a
  fixed list of loaded fonts)
- `cast` (array): `name`, `role`, `photo` (optional upload)
- `crew` (array): `name`, `role`
- `gallery` (array of uploads -> Media): the polished show photos
- `backstage` (array): `image` (upload), `caption` (text): behind-the-scenes shots

Admin UI: organize fields into tabs: Details, Theme, People, Gallery, Backstage.

### Runs
A season of a production, with its own dates and venue.
- `production` (relationship -> Productions, required)
- `label` (text, e.g. "2026 Season")
- `venue` (text)
- `shows` (array): `startsAt` (date+time), `status` (select: open, sold-out, cancelled),
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
      events/[slug]/page.tsx  # resolves production -> layout from registry
      archive/page.tsx        # past productions
      about/page.tsx
      contact/page.tsx
    (payload)/admin/...       # Payload-generated
  collections/
    Productions.ts  Runs.ts  Props.ts  Media.ts  Users.ts
  layouts/
    index.ts                  # registry: layoutKey -> component
    DefaultLayout.tsx
    crucible/CrucibleLayout.tsx   # example bespoke layout (own folder + css)
  components/
    ui/                       # Button, Badge, Container
    sections/                 # Hero, ShowDates, Synopsis, CastCrew, Gallery, Backstage
    EventCard.tsx
    ThemeProvider.tsx
  lib/
    payload.ts                # ONLY place that talks to the CMS
    types.ts
    format.ts
  payload.config.ts
```
Rule: **sections** are reusable blocks; **layouts** decide which sections appear, in what
order, and how they look. Bespoke layouts may reuse or replace sections.

## 6. Key Implementation Patterns

Layout resolution in `events/[slug]/page.tsx`:
```tsx
const production = await getProduction(params.slug)
if (!production) notFound()
const Layout = layouts[production.layoutKey] ?? layouts.default
return (
  <ThemeProvider theme={production.theme}>
    <Layout production={production} />
  </ThemeProvider>
)
```

Registry in `layouts/index.ts`:
```ts
export const layouts = {
  default: DefaultLayout,
  crucible: CrucibleLayout,
}
```

Theme injection in `ThemeProvider.tsx`:
```tsx
<div style={{
  '--bg': theme.bg, '--text': theme.text,
  '--accent': theme.accent, '--font-display': theme.displayFont,
} as React.CSSProperties}>{children}</div>
```
All themed styling must use `var(--bg)`, `var(--accent)`, etc., never hard-coded colors.

Data fetching in `lib/payload.ts`: use `payload.find` with `depth: 2` so gallery
images, runs, and relations arrive in one query. Export `getProduction(slug)`,
`getAllProductions()`, and `getRunsForProduction(id)`.

## 7. Pages

**Home:** hero for the current/upcoming production, short "about us" blurb,
upcoming shows row, past-posters strip linking to archive.

**Events listing:** poster-card grid. Each card shows title, next show date, and a
status badge (Upcoming / Now showing / Past). Upcoming first.

**Event page (the sub-site), default layout order:**
1. Full-screen hero (hero image or poster, title, tagline, primary "Book tickets" button)
2. Show dates: list of date, time, venue, status, and a button per row
3. Synopsis
4. Cast and crew
5. Gallery: masonry grid + lightbox
6. Backstage: captioned photos, visually distinct from the gallery
7. Link back to all events

All runs of a recurring production appear on the same production page, newest run first.

**Archive:** past productions as a poster grid.

## 8. Design System
- Mood: dark, editorial, quiet. Posters and photos provide the color.
- Fallback palette: bg `#0d0d0f`, text `#f2efe9`, accent muted gold (`#c9a227`).
  Productions override via theme.
- Type: serif display (Playfair Display or Cormorant) + Inter for body, loaded via
  `next/font`.
- Layout: max content width ~1200px, generous whitespace, full-bleed hero and gallery.
- Motion: subtle fade-up on scroll, slow hover zoom on posters. Heavier motion only in
  bespoke layouts.
- Mobile-first. Gallery: 2 columns on mobile, 3-4 on desktop. Use `next/image` with
  `sizes`. Lightbox must support swipe and keyboard.
- Accessibility: required alt text on media; check text/background contrast.

## 9. Scope for the Demo (build in this order)
1. Scaffold Next.js + Payload + Tailwind, fonts, global layout (nav/footer).
2. Create the collections in section 4; confirm /admin works with a seeded admin user.
3. Seed 2 productions with very different themes (e.g. dark gothic vs bright comedy),
   each with a run, several shows, gallery images, and backstage images. Use
   placeholder images if real ones aren't available.
4. Build `lib/payload.ts`, the events listing, and the `[slug]` route with the registry.
5. Build `DefaultLayout` and its sections (Hero, ShowDates, Gallery first).
6. If time remains: one bespoke layout registered as a third production.

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
- Keep each bespoke layout self-contained in `src/layouts/<name>/`.
- Do not access Payload outside `src/lib/payload.ts`.