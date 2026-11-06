# UI parity checklist (pre-revamp baseline)

Purpose: the ticketing revamp changes only the behaviour listed in `TICKETING.md`
§14. Everything in this file is the baseline that must still look and behave the
same after each stage. The baseline is git tag `pre-revamp`.

How to verify:

1. Run the site with `npm run dev` and open the URL at **desktop** (1280 px or
   wider) and **mobile** (390 × 844, device toolbar on).
2. Check every box. If an item is intentionally replaced by the revamp, confirm
   the replacement named under **Allowed revamp difference** instead.
3. For booking checks, enable ticketing for a show in Backstage during the
   baseline; after R1 the seeded Dear Daughter event is already active.
4. Record the exact show id from the booking link (`?show=...`) before walking
   the booking and confirmation pages.

Legend: `[ ]` = check at both widths unless stated. "Desktop only" and
"Mobile only" are explicit.

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

## 5. Booking page `/events/<slug>/book`

### Entry

- [ ] Event page reserve action opens `/events/<slug>/book?show=<show id>`.
- [ ] With no `show` parameter and exactly one show on sale, that show is used.
- [ ] With no `show` parameter and several shows on sale, the **Which
      performance?** picker appears: header eyebrow **Reserve seats**, production
      title, `Choose a performance to see available seats.`; each row shows
      formatted date, `run label · venue`, and **Choose seats →** linking to the
      same page with `?show=<id>`.
- [ ] No shows on sale: the same picker page says **There are no performances on
      sale right now. Check the event page for updates.**
- [ ] Booking pages render inside the production's design shell (header, fonts,
      colours, themed footer).

### Booking header

- [ ] Compact design-themed strip above the booking panel with eyebrow
      **Reserve seats**, production title, long date and venue.
- [ ] Dear Daughter keeps its damask strip, torn edge, paper panel and terracotta
      primary buttons; PlainDesign keeps the neutral dark style.

### Seat selection (desktop)

- [ ] Left panel title **Choose your seats** and note
      `Seats are held for 10 minutes once you continue. Pick up to 6.`
- [ ] While loading: **Loading seats…**.
- [ ] Load error shows the message plus a **Try again** link-style button.
- [ ] Stage label (or **Stage**) above the tier sections.
- [ ] Each tier section: tier name, `price · N of M left`, seat grid with the
      configured column count, horizontal scroll instead of page overflow.
- [ ] Seat tiles: label or code; available, selected, held, sold, blocked states;
      selected has `aria-pressed`; unavailable tiles disabled; status is not
      conveyed by colour alone (legend present).
- [ ] Legend: Available, Selected, On hold, Sold with swatches.
- [ ] Selecting up to 6 seats updates the summary immediately; selecting a 7th
      shows `You can reserve up to 6 seats at a time.` and does not add it.
- [ ] Seat selection polls availability about every 10 seconds before a hold;
      if a selected seat is taken it is removed and a notice explains why.

### Summary (right column on desktop, stacked on mobile)

- [ ] Title **Your order**; show title, formatted date and venue.
- [ ] Empty selection: **No seats selected yet.**
- [ ] Selected list: seat label (or code), tier name, formatted price each.
- [ ] **Total** row uses the show currency.
- [ ] **Hold seats** disabled until at least one seat is selected; busy label
      **Holding…**.
- [ ] Hold success: countdown `Seats held for mm:ss` starts at 10 minutes; the
      selection panel is replaced by **Your details**.
- [ ] Hold expiry: notice `Your seat hold expired. Choose your seats again.`,
      selection clears, map refreshes.
- [ ] 409 seat taken: notice `Someone else took one of those seats. We
      refreshed the map; your other choices are still selected.`
- [ ] **Change seats** releases the hold, returns to selection, and shows
      `Seats released. You can choose again.`

### Buyer form

- [ ] Copy: `We only use these details to send your tickets and contact you
      about this booking.`
- [ ] Required **Full name**, required **Phone (WhatsApp)**, optional **Email**,
      and a **Payment** select containing **Pay online (sandbox)** and
      **Reserve and pay by bank transfer**.
- [ ] Helper text changes with the payment method: PayHere completes immediately
      in demo mode; bank transfer holds seats for 24 hours while the transfer is
      confirmed.
- [ ] **Confirm booking** disabled while busy or after expiry; busy label
      **Confirming…**.
- [ ] **Change seats** remains available while not busy.
- [ ] Submit stores the order token in `sessionStorage` and navigates to
      `/events/<slug>/book/confirmation?order=<order id>`.
- [ ] Validation, API errors and expired holds appear in a visible status
      notice rather than a browser alert.
- [ ] Mobile: form fields and buttons are full-width/comfortable, no overlap
      with the themed footer.

## 6. Confirmation `/events/<slug>/book/confirmation?order=<id>`

- [ ] Missing `order` or unknown production → 404.
- [ ] Header **Booking confirmation**, production/show title, long date and venue
      when the summary is available.
- [ ] If the order token is missing from this browser: notice explaining that the
      order cannot be found, including the short reference
      `last 6 of order UUID, uppercase`.
- [ ] While loading: **Loading your order…**; load error shows the message and
      **Try again**.
- [ ] Status badge: Reserved (pending), Paid, Expired, Cancelled, or Checking
      payment (needs_review).
- [ ] Details: **Order reference**, **Seats** (`label (tier)` or `code (tier)`,
      comma-separated), **Total**, **Status**.
- [ ] Pending: notice that seats are reserved and the page refreshes
      automatically; polling about every 3 seconds until a final status.
- [ ] Needs review: text explaining payment was received but a seat was taken and
      the box office will contact the buyer.
- [ ] Expired: text explaining the hold expired and no seats are reserved.
- [ ] Paid: **Your tickets** heading, note about showing the page at the door,
      and one link per ticket showing seat label/code, tier and **View ticket →**.
- [ ] Aside **Keep this reference** repeats the short reference.
- [ ] Actions: **Back to event** → `/events/<slug>`; **Book another
      performance** → `/events/<slug>/book`.
- [ ] Mobile stacks the aside below the main panel; all text wraps without
      overflow.

## 7. Ticket page `/t/<token>`

- [ ] Invalid or unknown token → 404; page metadata is `noindex, nofollow`.
- [ ] Renders inside the production's design shell and themed footer.
- [ ] Eyebrow **Your ticket**, show title (or fallback), long date and venue.
- [ ] Badge **Valid** for `issued`, otherwise the raw status (used/void) with a
      warning tone.
- [ ] Details: **Admission** (ticket description), **Price**, **Order
      reference**, **Ticket status**.
- [ ] QR block: SVG QR code on white, `QR code unavailable` fallback, and the
      ticket URL printed underneath.
- [ ] Invalid ticket: notice that it cannot be admitted and to contact the box
      office.
- [ ] Valid ticket: note to present the QR at the door; each seat has its own
      ticket.
- [ ] No buyer name is shown anywhere.
- [ ] QR remains scannable at mobile width; content does not overflow.

## 8. Cross-cutting behaviour

- [ ] No console errors or hydration warnings on any checked page.
- [ ] No request to the ticketing API when ticketing is inactive.
- [ ] Every visible date uses Asia/Colombo and the existing English formats
      (`Sat, 14 Nov 2026, 7:00 pm` and
      `Saturday, 14 November 2026 at 7:00 pm`).
- [ ] Money displays as `Rs. 2,500` / `Rs. 2,500.00` exactly as before (whole
      rupees in Payload, integer cents in the API/Go).
- [ ] The WhatsApp fallback remains available when ticketing is off or not
      active; it opens `wa.me/<digits>` with a prefilled message.
- [ ] No new paid service or external script is introduced.

## 9. Allowed revamp differences (do not treat as parity failures)

These are the only intentional changes; everything else above must stay.

- **R1 data/content:** Dear Daughter is re-seeded with three placeholder future
  shows, venue **The Black Bean Cafe, Kandy**, one **General** tier at Rs. 2,500
  × 50, quantity mode, stage front, bank transfer. A fictional dev-only event may
  appear on the fallback design. The Backstage editor becomes one **Events**
  editor with four tabs (Details, Shows & tickets, People, Gallery); Runs and
  Props disappear.
- **R2 booking:** quantity mode uses tier boxes with steppers and a stage bar
  instead of the seat grid; seats mode keeps the seat grid and gains the stage
  bar. The payment method is decided by the event (bank transfer, online, or
  off), so the buyer no longer chooses it in the form. Bank-transfer
  confirmation gains bank details, a short reference, a deadline, polling and a
  ticket link when paid.
- **R3 event page:** the show-dates list becomes a selector; the booking card is
  the only reserve action; the hero CTA and mobile sticky bar become **Get
  tickets** anchors that scroll to the card; the selected show is reflected in
  `?show=` with `history.replaceState` and restored on load; past shows are
  dimmed and not selectable; cancelled shows are marked.
- **R4 cleanup:** no visual/behavioural change is allowed. Only dead code,
  dependencies, duplicate/unused CSS and bundle loading may change.
