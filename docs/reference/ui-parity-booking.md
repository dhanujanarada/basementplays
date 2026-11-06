# UI parity: booking, confirmation, and tickets

- **Status:** implemented
- **Covers:** `src/app/(frontend)/events/[slug]/book/`, `src/app/(frontend)/t/[token]/`, `src/components/ticketing/`
- **Last checked:** 8de3320

Part of the [UI parity checklist](../ui-parity.md). Check at desktop and mobile
widths unless stated. Legend: `[ ]` = check at both widths.

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
