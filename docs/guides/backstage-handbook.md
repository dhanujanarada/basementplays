# Backstage handbook

- **Status:** partial
- **Covers:** `src/collections/Productions.ts`, `src/globals/SiteSettings.ts`, `src/payload.config.ts`
- **Last checked:** 8de3320

This guide is for the people who run the Basement Plays website. It explains the
**Backstage** admin panel. It does not cover writing code.

**Current limit:** marking a bank transfer paid and copying a ticket link are
planned staff pages. Until those pages exist, ask a developer to do those two
steps through the ticketing service. Everything else here works in Backstage.

## Log in

1. Open `/backstage`.
2. Use the account the team gave you.
3. After five failed attempts the account locks for ten minutes.

Members can edit Events, Media, and Site Settings. Only admins can manage users.

## Create an event

1. Open **Events**.
2. Click **Create New**.
3. In **Details**, fill in:
   - title and slug;
   - tagline, presented by, written by, directed by;
   - synopsis;
   - poster (required) and optional hero image.
4. Save.

The event appears in the Events list after saving. If it has no show dates yet,
its page shows the fallback date message, and home/listing cards show “Check back
for future dates” or “No upcoming shows”.

## Add show dates

Open the **Shows & tickets** tab.

1. Set the default **venue** for the event.
2. Add a row under **Shows**.
3. Set the date and time. The site displays it in Asia/Colombo.
4. Add an optional venue override if one performance is elsewhere.
5. Tick **Cancelled** when a performance is called off. Keep the row so the
   public page can show it as cancelled.

Do not delete a show that has sold tickets. Set it to cancelled instead.

## Set tiers and ticket settings

In the same tab, open the **Ticketing** group.

- **Payment mode:** bank transfer (default), online, or off.
  - Bank transfer: the buyer sees bank details and staff confirm the transfer.
  - Online: PayHere sandbox; live payments are not configured.
  - Off: the event shows the WhatsApp fallback instead of a booking flow.
- **Selection mode:** quantity (tier boxes, the default) or seats (numbered
  grid).
- **Stage position:** front, back, or none. Front puts the stage bar above the
  tiers; back puts it below and reverses the tier order.
- **Numbering** and **columns** only appear in seats mode.
- **Tiers:** add at least one. Each tier needs a name, a whole-rupee price, and a
  count. The first tier is nearest the stage.

Ticketing is active only when payment mode is not off and at least one tier has
a positive price and a positive count. If it is inactive, the event page uses
the WhatsApp fallback.

Changing tier names, prices, or display settings is safe. Changing tier count,
adding/removing a tier, or reordering tiers is a structural change. Once sales
or holds exist, the ticketing service should reject it and Backstage should warn
you. The warning is planned; do not assume a structural edit succeeded until the
service is wired up.

## Add people, partners, and gallery

**People** tab:

- Cast: name, role, optional photo.
- Crew: name and role.
- Partners: name, role, optional logo. No logo is fine; the site prints the name.

**Gallery** tab:

- Upload an image.
- Add a caption.
- Choose **On stage** or **Behind the scenes**.

Use meaningful alt text on every image. It is required by the Media collection
and read by screen readers.

## Site Settings

Open **Site Settings** to edit:

- group name, tagline, about text, about image;
- email, phone, WhatsApp number, location, maps URL;
- social links;
- bank transfer details.

The WhatsApp number is used by the fallback link. The bank details are meant for
the confirmation page. Keep them current before opening a bank-transfer event.

## Mark a bank transfer paid (planned)

Today there is no Backstage button for this. A developer uses the ticketing
internal API. The intended staff page will live under `/backstage-tools`, use
your Backstage login, and show the order list with a **Mark paid** action.

When it exists:

1. Open the show's orders.
2. Find the order by short reference.
3. Confirm the transfer arrived.
4. Mark it paid.
5. The buyer's confirmation page then reveals the ticket links.

## Copy a ticket link (planned)

Today the buyer sees ticket links on their confirmation page after payment. The
planned staff page will show the same links so staff can copy one when a buyer
loses their tab.

## What to do when something is wrong

- Wrong date or venue: edit the show and save. If the date changes, tell the
  ticketing developer; sold-seat inventory is tied to the show id.
- A performance is cancelled: tick **Cancelled**. Do not delete it.
- A show is sold out: this should be derived from the ticketing service, not
  typed by hand. Do not add a manual status field.
- Booking is not active: check payment mode is not off and at least one tier has
  a price and count above zero.
- The site shows the wrong contact details: edit Site Settings, then refresh.

## Related

- [Adding a new event](adding-a-new-event.md)
- [Operations](operations.md)
- [Events data model](../architecture/data-model.md)
