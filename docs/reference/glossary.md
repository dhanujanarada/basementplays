# Glossary

- **Status:** implemented
- **Covers:** `src/`, `ticketing/`, `docs/`
- **Last checked:** 8de3320

Terms are defined here once and linked from other documents.

| Term | Meaning |
| --- | --- |
| **Backstage** | The Payload admin panel at `/backstage`. Never a public gallery. |
| **Booking card** | The single reserve action on an event page, rendered by a design through `SeasonBooking`. |
| **Buyer** | The person holding seats and completing checkout. Buyers do not have accounts. |
| **Covers** | Metadata in a document listing the repository paths it describes. |
| **Design** | A self-contained visual component under `src/designs/<slug>/`, registered by event slug. |
| **Event** | A production in the Payload `productions` collection. The admin label is Events. |
| **EventView** | The normalized event shape from `src/lib/event-view.ts`; the only shape designs consume. |
| **Few left** | A derived availability badge when active ticketing has a small positive number of seats remaining. |
| **Hold** | A temporary claim on seats, stored as a pending order with an expiry. |
| **Internal API** | Go routes under `/v1/internal/*`, used server-to-server with a bearer token. |
| **Mock client** | In-browser ticketing implementation used for demos and tests. |
| **Order** | A hold plus buyer details. It can be pending, paid, expired, cancelled, or needs review. |
| **Order token** | Secret returned once with a hold; required to release, check out, or read an order. |
| **Payload** | The CMS running inside the Next.js app. |
| **Payment mode** | Per-event choice: bank transfer, online (PayHere), or off (WhatsApp fallback). |
| **Production** | The historical name for an event; it remains the Payload collection slug. |
| **Quantity mode** | Ticket selection by tier quantity; the server assigns the lowest available seats. |
| **QR token** | HMAC-signed token in a ticket URL, verified by the Go service. |
| **Run** | A previous grouping of shows. Runs were removed in favour of an event-level shows list. |
| **SeasonBooking** | Shared client component that owns show selection and the `?show=` URL state. |
| **Seats mode** | Ticket selection from a numbered seat grid. |
| **Shell** | The design component that provides palette, fonts, textures, header, and themed footer. |
| **Show** | One performance date in an event's `shows` array. Its row id is used in the URL and by Go. |
| **Stage position** | Whether the stage bar is above the tiers, below them, or absent. |
| **Tier** | A price/count bucket such as General or Front. Tiers generate seats. |
| **Ticketing service** | The Go API in `ticketing/` with its own Postgres schema. |
| **UI parity** | The checklist in `docs/ui-parity.md` that guards pre-revamp look and behaviour. |
| **WhatsApp fallback** | A `wa.me` link built from Site Settings when online ticketing is not active. |

## Related

- [Documentation index](../README.md)
- [Architecture overview](../architecture/overview.md)
- [UI parity checklist](../ui-parity.md)
