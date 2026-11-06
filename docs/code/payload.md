# Payload CMS

- **Status:** implemented
- **Covers:** `src/payload.config.ts`, `src/collections/`, `src/globals/SiteSettings.ts`, `src/payload-types.ts`, `src/lib/payload.ts`
- **Last checked:** 8de3320

Payload 3 runs inside the Next.js app. It provides the admin panel, the content
API, authentication, uploads, and generated TypeScript types.

## Config

`src/payload.config.ts` sets the admin route, registers collections and globals,
uses the Lexical editor, and chooses the SQLite adapter when no database URL is
provided. `withPayload` wraps the Next config.

```ts
// src/payload.config.ts — buildConfig
routes: { admin: '/backstage' },
collections: [Users, Media, Productions],
globals: [SiteSettings],
db: sqliteAdapter({ client: { url: process.env.DATABASE_URL || 'file:./payload.db' } })
```

## Collections

- **Events** (`productions`) — one editor, four tabs: Details, Shows & tickets,
  People, Gallery. It contains title, slug, tagline, credits, synopsis, poster,
  hero image, shows, ticketing settings, cast, crew, partners, and gallery.
- **Media** — uploads with required alt text and three image sizes.
- **Users** — auth collection with `admin` and `member` roles, five login
  attempts, and a ten-minute lock.

There is no Runs collection and no Props collection.

## Site Settings

A global singleton with public identity, contact details, social links, and the
`bankTransfer` group. It feeds the header, footer, About, Contact, WhatsApp
fallback, and bank-transfer confirmation.

## Tabs and fields

Tabs are a UI grouping only; they do not change the stored document shape. The
`ticketing` group holds `paymentMode`, `selectionMode`, `stagePosition`,
`numbering`, `columns`, and `tiers`. `numbering` and `columns` have admin
conditions so they only show in seats mode.

```ts
// src/collections/Productions.ts — ticketing columns field
{
  name: 'columns',
  type: 'number',
  defaultValue: 10,
  min: 1,
  admin: {
    condition: (data) => data?.ticketing?.selectionMode === 'seats',
    description: 'Seat grid width in seats mode.',
  },
}
```

## Array row ids

Payload gives each array row an `id`. The shows array uses that id as the show id
in `?show=` and as the Go show id. The tiers array uses it as the tier id. The
current SQLite adapter was checked by reversing and renaming array rows; the ids
survived. Do not replace these ids with array indexes.

## Hooks

`Productions` has a `beforeValidate` hook that slugifies `title` and makes the
slug unique. There is no event save hook yet that syncs shows to Go. The target
sync hook is planned.

## Access control

- Events, Media, and Site Settings are publicly readable.
- Create/update/delete content requires a logged-in user.
- Users is readable and writable only by admins; the first-user flow is the one
  exception.
- No public registration.
- Admin navigation is Events, Media, Users, and Site Settings only.

## Admin route

The admin is at `/backstage`, not `/admin`. `robots.txt` disallows the path. The
site never links to it. Hiding is not security; the login and roles are.

## Data access rule

Only `src/lib/payload.ts` calls Payload on the frontend. It returns `EventView`,
never raw records. Components and designs do not import Payload.

## Related

- [Data model](../architecture/data-model.md)
- [Security](../architecture/security.md)
- [Site walkthrough](site-walkthrough.md)
