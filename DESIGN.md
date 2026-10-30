# Theatre Group Website + CMS — Design Doc

## 1. Goal
A public website for a theatre group with a built-in CMS. Each production gets its own
"sub-site" style page with a custom look, show dates, cast/crew, and a photo gallery.
The admin panel manages content and per-production prop tracking.

The admin panel is called **Backstage**. "Backstage" refers only to the admin panel —
never to public photo sections, gallery categories, or content types. Backstage holds
**content only**; each production's presentation lives in its own code component under
`src/designs/<slug>/`. Payments come later (out of scope for now).

**Immediate target:** a demo-ready basic site (listing + event pages + Backstage) by
tomorrow.

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
- `gallery` (array): `image` (upload -> Media, required), `caption` (text, optional)

There is no `theme` group, no `layoutKey`, no `backstage` array, no gallery
`category`, and no `testimonials` field. Do not add them back.

Admin UI tabs: Details, Partners, People, Gallery.

### Runs
A season of a production, with its own dates and venue.
- `production` (relationship -> Productions, required)
- `label` (text, e.g. "2026 Season")
- `venue` (text)
- `shows` (array): `startsAt` (date+time), `status` (select: open, few-left,
  sold-out, cancelled), `price` (text), `seatsNote` (text, optional),
  `ticketUrl` (text, optional)

### Props
Per-production prop tracking, an internal Backstage tool (not shown publicly).
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
Payload auth collection with roles: `admin`, `member`.
- Only admins can create, update, or delete users.
- Members can edit Productions, Runs, Media, Props, and Site Settings, but cannot
  manage users.
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
- `whatsappNumber` (text, international format, digits only)
- `locationText` (text)
- `mapsUrl` (text)
- `socialLinks` (array): `platform` (select: instagram, facebook, youtube, tiktok,
  x, other), `url` (text, required)

### Access control
- Public read: Productions, Runs, Media, Site Settings.
- Never publicly readable: Props, Users.
- Create/update/delete content: any logged-in user (admin or member).
- User management: admins only.
- Admin panel route: `/backstage`. `/admin` returns 404.
- `robots.txt` disallows `/backstage`.
- The public site never links to Backstage.

Admin sidebar groups: **Content** (Productions, Runs, Media, Site Settings),
**Production** (Props), **System** (Users).

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
    Productions.ts  Runs.ts  Props.ts  Media.ts  Users.ts
    access.ts
  globals/
    SiteSettings.ts
  designs/
    index.ts                  # registry: production slug -> design component
    PlainDesign.tsx           # neutral fallback, shows all content
    dear-daughter/
      DearDaughterDesign.tsx  # self-contained production design
      dear-daughter.css       # design-owned colors and styles
  components/
    ui/                       # Button, Badge, Container
    EventCard.tsx
    SiteHeader.tsx
    SiteFooter.tsx            # reusable themed footer
  lib/
    payload.ts                # ONLY place that talks to the CMS
    types.ts
    format.ts
  payload.config.ts
```
Rule: each production design is self-contained in `src/designs/<slug>/` and owns its
layout, styling, and assets. Designs may reuse shared `components/ui` pieces. Backstage
stores content only.

## 6. Key Implementation Patterns

Design resolution in `events/[slug]/page.tsx`:
```tsx
const production = await getProduction(slug)
if (!production) notFound()
const Design = designs[production.slug] ?? PlainDesign
return <Design production={production} runs={runs} />
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
relations arrive in one query. Export `getProduction(slug)`, `getAllProductions()`,
`getRunsForProduction(id)`, and `getSiteSettings()`.

The shared footer reads Site Settings and is styled with neutral fallbacks:
`--footer-bg` (`#111`), `--footer-text` (`#eee`), `--footer-accent` (`#c9a227`).
Production designs may override those variables on their root element.

Footer placement: the frontend layout renders `<SiteFooter />` for every page, but
`FooterGate` (a client component that checks `usePathname`) suppresses it on
`/events/[slug]`. Each production design renders `<SiteFooter />` inside its own root
element instead, so the design's CSS variables theme the footer. This keeps exactly
one footer per page. Header and footer both read Site Settings (`groupName`, contact
details, social links) so content stays in Backstage.

## 7. Pages

**Home:** the next production, a short About teaser with a "Read more" link to
`/about`, and the upcoming shows.

**Events listing:** poster-card grid of **all** productions. Each card links to
`/events/<slug>` and shows title, next show date, and a status badge. Missing posters
are handled gracefully. Upcoming first.

**Event page (the sub-site):** a production's registered design renders the page. The
fallback `PlainDesign` shows every content field:
1. Hero with title, tagline, credits, and poster or hero image
2. Show dates: date, time, venue, status, price, seats note, and ticket link
3. Synopsis
4. Cast and crew
5. Gallery: masonry grid, captions, and a lightbox
6. Partners
7. Link back to all events

All runs of a recurring production appear on the same production page, newest run first.

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
- Productions: remove `theme`, `layoutKey`, `backstage`, gallery `category`, and
  `testimonials`; add `presentedBy`, `writtenBy`, `directedBy`, `partners`; make
  `gallery` an array of `{ image, caption }`.
- Runs: `venue`, show `price`, `seatsNote`, and the four show statuses.
- Add the Site Settings global.
- Public read on Productions, Runs, Media, Site Settings; never on Props or Users.
- Move the admin panel to `/backstage`, title suffix " - Backstage", add `robots.txt`.
- Users roles `admin`/`member`, admin-only user management, login lockout.
- Remove all public links to Backstage.

### Step B: Seed content
- Run all seeders with `npm run seed` (admin users, Site Settings, Dear Daughter).
- Seed "Dear Daughter," with credits, lorem ipsum synopsis, 4 cast, 3 crew, one run
  and placeholder show, WhatsApp ticket URL, partners, and 8 captioned gallery images.
- Seed Site Settings with placeholder contact and social values.
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
- Hero with wallpaper and partner strip; cream section with torn edge, show details,
  and booking block; wood divider; synopsis, cast/crew, gallery with captions and
  lightbox; mobile sticky Reserve bar.
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
- Payments / checkout
- User accounts for the public
- Prop dashboard styling beyond the default Backstage UI
- Email, ticket QR codes, or booking logic
- SEO polish and analytics

## 11. Future Work (design so these are easy to add later)
- Archive page for past productions.
- Payment gateway (likely PayHere for Sri Lanka; verify availability). Add a `Bookings`
  collection with status pending/paid; the gateway webhook marks it paid.
- Link "Book tickets" buttons to the booking flow instead of an external `ticketUrl`.
- Move SQLite to Postgres and deploy.

## 12. Conventions
- TypeScript strict mode; shared types in `src/lib/types.ts`.
- Server components by default; client components only for lightbox/animation.
- Keep each production design self-contained in `src/designs/<slug>/`.
- Backstage holds content only: no colors, fonts, or layout keys in Payload.
- "Backstage" means the admin panel only. Never use it for photos, gallery categories,
  or public sections. Do not add testimonials.
- Do not access Payload outside `src/lib/payload.ts`.
