# Payments

- **Status:** partial
- **Covers:** `ticketing/internal/service/orders.go`, `ticketing/internal/config/config.go`, `src/components/ticketing/BookingFlow.tsx`, `src/components/ticketing/OrderConfirmation.tsx`, `src/globals/SiteSettings.ts`
- **Last checked:** 8de3320

Two payment modes are represented in the product: bank transfer and PayHere.
The bank-transfer flow is implemented in Go; PayHere is still planned. The
frontend mock implements both so the UI can be demonstrated without a payment
provider.

## Bank transfer

1. The buyer chooses an event whose `paymentMode` is `bank_transfer`. The buyer
   does not choose the method in the form.
2. Checkout validates name, phone, and optional email. In Go, a bank-transfer
   checkout sets the order source to `manual` and extends the hold to
   `MANUAL_HOLD_HOURS` (24 by default).
3. The confirmation page shows the order status, short reference, and deadline.
   The mock also shows bank details from Site Settings. The current Go public
   order type does not yet include bank details or `payment_method`, despite the
   TypeScript contract expecting them.
4. Staff mark the order paid through the internal API. `MarkPaid` sells the
   seats, issues one ticket per seat, and is idempotent.
5. Once paid, the confirmation page shows one ticket link per seat. Staff can
   copy those links from the eventual staff page; the public page already has
   the links.

## PayHere sandbox

The current code has configuration guards only:

- `PAYHERE_SANDBOX` defaults to `true`.
- Starting with `PAYHERE_SANDBOX=false` is refused unless
  `PAYHERE_ALLOW_LIVE=true`.
- Merchant id and secret are loaded from the environment.

There is no `internal/payhere` package, no checkout hash, no notify callback,
and no re-check-payment endpoint in this repository yet. The target flow is:

1. Server builds the checkout hash and redirects to PayHere sandbox.
2. PayHere calls the notify URL. The service verifies the signature, merchant,
   amount, and currency against the stored order.
3. A success notification calls `MarkPaid`; a duplicate is ignored.
4. The return URL never marks an order paid. A staff “Re-check payment” action
   handles reconciliation.

## Order and ticket effects

- Bank transfer and PayHere both end in the same `MarkPaid` path: seats become
  sold, tickets become issued.
- If payment arrives after the hold expires and the seats are gone, the order
  becomes `needs_review`.
- Cancelling a paid order voids tickets and releases seats according to the
  state machine.

## Secrets and trust

Payment amounts come from stored seat prices, never from the client. Secrets are
server-only. The browser may know the merchant id but never the merchant secret.
No payment secret is logged.

## Related

- [State machines](state-machines.md)
- [Security](security.md)
- [API contract](api-contract.md)
- `docs/DISCREPANCIES.md` (D5) will record the current PayHere gap.
