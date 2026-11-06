# UI parity checklist (pre-revamp baseline)

- **Status:** implemented
- **Covers:** `docs/reference/ui-parity-pages.md`, `docs/reference/ui-parity-booking.md`
- **Last checked:** 8de3320

Purpose: the ticketing revamp changes only the behaviour listed in `TICKETING.md`
§14. Everything in this checklist is the baseline that must still look and behave
the same after each stage. The baseline is git tag `pre-revamp`.

The checklist is split so no page grows past the documentation limit:

- [Pages, event pages, and fallback design](reference/ui-parity-pages.md)
- [Booking, confirmation, ticket, and cross-cutting checks](reference/ui-parity-booking.md)

How to verify:

1. Run the site with `npm run dev` and open the URL at **desktop** (1280 px or
   wider) and **mobile** (390 × 844, device toolbar on).
2. Check every box. If an item is intentionally replaced by the revamp, confirm
   the replacement named under **Allowed revamp difference** instead.
3. For booking checks, use the seeded Dear Daughter event, which is active.
4. Record the exact show id from the booking link (`?show=...`) before walking
   the booking and confirmation pages.
5. Legend: `[ ]` = check at both widths unless stated. "Desktop only" and
   "Mobile only" are explicit.
