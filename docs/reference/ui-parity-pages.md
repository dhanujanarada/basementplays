# UI parity: pages and event designs

- **Status:** implemented
- **Covers:** `src/app/(frontend)/`, `src/designs/`, `src/components/`
- **Last checked:** 8de3320

Part of the [UI parity checklist](../ui-parity.md). Check at desktop and mobile
widths unless stated. Legend: `[ ]` = check at both widths.

## 0. Global chrome (all pages)

- [ ] Sticky header at the top, 80 px tall, separated by a thin border and a
      translucent blurred site background.
- [ ] Header left: the site name from Site Settings (`groupName`, default
      "Basement Plays"), linked to `/`. Display/serif face, hover turns the
      accent colour.
- [ ] Header right: `Home`, `Events`, `About`, `Contact` in that order,
      uppercase, small, wide letter-spacing, muted colour; hover turns accent.
- [ ] Keyboard focus is visible on every header/footer link.
- [ ] Page background, text, muted text, accent and border colours follow the
      global CSS variables; no horizontal scrollbar at either width.
- [ ] Footer appears exactly once per page. On ordinary pages it is the global
      footer; on an event, booking, confirmation or ticket page it is rendered
      inside that production's design shell.
- [ ] Footer blocks: site name (display face) and tagline; **Explore** links
      Home/Events/About/Contact; **Contact** with email (mailto), WhatsApp
      (opens `wa.me` in a new tab), phone (tel) and location (opens maps when a
      maps URL exists, otherwise plain text); **Follow** social links when
      configured; copyright line `© <year> <groupName>. All rights reserved.`
- [ ] Empty optional fields hide their own element rather than leaving a blank
      row.
- [ ] Fonts, colours, spacing and hover transitions match the pre-revamp
      screenshots/recording.
- [ ] `prefers-reduced-motion` removes fade-up, hover zoom and smooth scrolling.
- [ ] Unknown event slugs show the site 404 page.

## 1. Home `/`

### Desktop

- [ ] Hero section fills at least 70 vh, with a faint grid texture behind it.
- [ ] Left: a pill badge reading **Next up** (or the group name when no show is
      upcoming).
- [ ] Next production: venue (run venue, then Site Settings location, then group
      name), large display-face title, tagline when present.
- [ ] Status badge (Open / Few left / Sold out / Cancelled) followed by the next
      show date and time in Asia/Colombo.
- [ ] Primary button **View production** links to `/events/<slug>`; secondary
      **All events** links to `/events`.
- [ ] Right: poster in a 2:3 frame with border, shadow and rounded-sm corners.
      Missing poster shows the grid texture, group name and title.
- [ ] No upcoming show: group name as the hero title, tagline, and a single
      **See what's on** button to `/events`; no poster.
- [ ] About teaser: eyebrow **About**, heading `A little about <groupName>`,
      one plain-text teaser paragraph (320 characters, ellipsis at a word
      boundary), secondary **Read more** link to `/about`.
- [ ] Upcoming shows: eyebrow **Upcoming shows**, heading **On stage next**,
      **All events** button.
- [ ] Card grid: 3 columns on desktop, 2 on tablet, 1 on mobile. Each card is one
      link to `/events/<slug>` and shows:
      - poster (or a textured fallback with group name and title);
      - status badge in the top-left over the poster (or **No upcoming shows**);
      - title in the display face;
      - formatted next-show date, or **Check back for future dates**.
- [ ] Card hover: border turns accent, poster zooms slightly, title turns
      accent; reduced motion disables the zoom.
- [ ] With no upcoming productions: the grid is replaced by a bordered
      paragraph, **No upcoming shows are listed yet. Check back soon.**

### Mobile

- [ ] Header compresses without overlap; nav items remain on one line.
- [ ] Hero stacks copy above the poster; poster is centred and stays inside the
      viewport.
- [ ] About teaser stacks heading above text; buttons do not overflow.
- [ ] Cards become a single column with the same order and content.

## 2. Events `/events`

- [ ] Page padding matches the rest of the site (large top/bottom space).
- [ ] Eyebrow **Programme**, H1 **Events**, description
      `Productions from <groupName>, with upcoming show dates and ticket links.`
- [ ] Right-aligned count `N production` / `N productions` when at least one
      production exists; hidden when the list is empty.
- [ ] Cards are the same `EventCard` as Home (poster, status badge, title, next
      date, whole-card link).
- [ ] Order: productions with an upcoming show first, earliest next show first;
      remaining productions alphabetically by title.
- [ ] Grid: 1 column mobile, 2 tablet, 3 desktop.
- [ ] Empty state (bordered paragraph):
      **No productions have been published yet.**
- [ ] Every card links to `/events/<slug>` and never directly to a booking or
      ticket URL.

## 3. Event page `/events/dear-daughter` (Dear Daughter design)

### Desktop

**Hero**

- [ ] Full-bleed stage-green header with hero photo, dark veil, damask wallpaper
      and decorative seal; all layers fall back gracefully when image/texture
      files are missing.
- [ ] Back link **← All events** to `/events`.
- [ ] Eyebrow: `presentedBy` or **A Basement production**.
- [ ] Title in Caveat Brush, tagline in Quicksand.
- [ ] Credits list only non-empty values: **Presented by**, **Written by**,
      **Directed by**.
- [ ] Hero actions: primary **Reserve seats** (or **See show dates** when there
      is no bookable next show) plus a text link **Show dates**.
- [ ] Below the actions: next-show long date and venue with a small decorative
      marker, shown only when a next show exists.
- [ ] Right: poster in a 2:3 frame with caption `venue · run label`; missing
      poster shows the design fallback.
- [ ] Partner strip only when partners exist: **With support from**, then each
      partner logo (or name as text) and role.

**Season / show dates**

- [ ] Torn-paper top edge and paper-coloured section.
- [ ] Eyebrow **Performances**, title **Show dates**, note
      `Join us at <venue>.` or the announced-soon fallback.
- [ ] Runs newest first; each run shows its label and venue.
- [ ] Each show row (ascending by date) shows weekday, day, month in a date
      block; time; price and seats note joined with ` · `; a status badge
      (Open / Few left / Sold out / Cancelled); and a **Reserve** button when the
      show is bookable.
- [ ] Cancelled rows are dimmed; sold-out/cancelled rows do not show a reserve
      button.
- [ ] Empty run: **No show dates have been announced for this run yet.**
      Empty event: **No show dates have been announced yet. Check back soon.**

**Booking card (right column on desktop; sticky beneath the header)**

- [ ] Eyebrow **Booking**, title **Reserve your seats**.
- [ ] Definition list: **Next show** long date; **Venue**; **Tickets** (price
      text); **Availability** (seats note) — each row only when its value
      exists.
- [ ] Primary **Reserve seats** when the next show is bookable; otherwise a note
      like `This performance is sold out.`
- [ ] Closing note: `Reservations are handled through the ticket link. Seats are
      confirmed directly by the company.`
- [ ] No next show: **Booking opens as soon as the next date is announced.**

**Rest of the page**

- [ ] Wood divider between the season section and synopsis.
- [ ] Synopsis section: eyebrow **The play**, title **Synopsis**, decorative
      sprig, rich text with the design typography; omitted when empty.
- [ ] Cast & crew: eyebrow **The company**, title **Cast & crew**; cast in two
      columns on wide screens with portrait (or initials fallback), name and
      role; crew column with name and role; empty column hidden.
- [ ] Gallery: eyebrow **Production stills**, title **Gallery**, note
      **Click any image to view it larger.** Masonry (2 columns mobile, 3
      desktop), captions under images, hover zoom.
- [ ] Clicking a gallery image opens the lightbox at that image; close button,
      Escape, swipe and arrow keys work; backdrop click closes.
- [ ] Outro: accent band, eyebrow **More from Basement Plays**, title
      **See what else is on.**, light **All events** button.
- [ ] Footer is the themed wood footer.
- [ ] Every link and button has a visible keyboard focus style.

### Mobile

- [ ] Hero content stacks; title/tagline/credits remain readable; poster centres.
- [ ] Partner strip wraps without overflow.
- [ ] Show rows stack date block, details and action; reserve button remains
      reachable and at least 32 px tall.
- [ ] Booking card appears after the show list, not sticky.
- [ ] **Mobile sticky bar** appears at the bottom only when the next show is
      bookable and the viewport is under 720 px:
      - label **Next show** and the formatted date on the left;
      - **Reserve** button on the right;
      - fixed above the safe-area inset, with the footer padded so content is
        not hidden;
      - hidden at 720 px and wider.
- [ ] Cast grid collapses to one column; gallery stays two columns.
- [ ] Outro stacks copy above the button.

## 4. Event page fallback (PlainDesign)

Used for productions without a custom design (the revamp adds a fictional
dev-only event here).

- [ ] Hero: optional background hero image at 30% opacity with a dark gradient;
      grid texture; **← All events**; status badge (or **No upcoming shows**)
      plus `presentedBy`; large serif title; tagline; credits; **See show dates**
      anchor and **Browse all events** button; poster on the right.
- [ ] Show dates: eyebrow **Performances**, title **Show dates**; runs newest
      first, shows ascending; each row shows long date, `price · seats note`,
      status badge, and a **Get tickets** button when bookable; cancelled rows
      dimmed.
- [ ] Partner strip **With support from** with logos/names and roles.
- [ ] Synopsis rich text with the neutral editorial typography.
- [ ] Cast & crew lists with round cast photos, names and roles.
- [ ] Gallery masonry + captions + lightbox.
- [ ] Outro **More from Basement Plays / See what else is on.** with **All
      events**.
- [ ] No mobile sticky bar in PlainDesign.
