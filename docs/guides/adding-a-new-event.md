# Adding a new event

- **Status:** implemented
- **Covers:** `src/collections/Productions.ts`, `src/globals/SiteSettings.ts`, `src/designs/index.ts`
- **Last checked:** 8de3320

This is the content checklist. For the click-by-click admin tour, see the
[Backstage handbook](backstage-handbook.md).

## Before you start

- Decide the event title, slug, tagline, credits, and synopsis.
- Have the poster ready. A hero image is optional.
- Collect cast, crew, and partner details.
- Have gallery images and captions.
- Decide the venue, show dates, tiers, payment mode, selection mode, and stage
  position.
- Decide whether the event needs a custom design. If not, `PlainDesign` is the
  fallback.

## Create the event

1. In Backstage, open **Events** → **Create New**.
2. Fill **Details**. Use a clear title; the slug is generated but can be edited.
3. Fill **Shows & tickets**:
   - default venue;
   - show dates and optional venue overrides;
   - cancelled flags where needed;
   - payment mode, selection mode, stage position;
   - numbering and columns only in seats mode;
   - at least one tier with name, rupee price, and count.
4. Fill **People** and **Gallery**.
5. Save.

## After saving

- Check the event page at `/events/<slug>`.
- Check the show selector and that `?show=` updates.
- Check the booking card is the only reserve action.
- Check the fallback if payment mode is off.
- Check mobile width.
- If the event should use a custom design, follow
  [Adding a new design](adding-a-new-design.md). Until it is registered, the
  event uses `PlainDesign`.

## Things not to do

- Do not add a manual status, ticketing-enabled checkbox, ticket URL, price text,
  or seats note. Those fields were removed.
- Do not delete a show with orders. Mark it cancelled.
- Do not change tier count, add/remove tiers, or reorder tiers after sales
  without expecting a `seating_locked` rejection.
- Do not paste real bank details into documentation or test fixtures.

## Related

- [Backstage handbook](backstage-handbook.md)
- [Adding a new design](adding-a-new-design.md)
- [Events data model](../architecture/data-model.md)
