# Request lifecycles

- **Status:** partial
- **Covers:** `src/components/ticketing/`, `src/lib/ticketing/`, `ticketing/internal/httpapi/`, `ticketing/internal/service/`, `ticketing/internal/store/`
- **Last checked:** 8de3320

These traces connect the UI, the API, the service, and the database. PayHere is
planned; the other flows exist in code today.

## 1. Buyer reserves by bank transfer

1. The event page's `SeasonBooking` selects a show and sends the buyer to
   `/events/<slug>/book?show=<id>`.
2. `BookingFlow` calls `getShowSummary` and, in seats mode, `getSeatMap`.
3. The buyer selects seats. Quantity-mode selection is implemented in the
   frontend/mock; the current Go service only accepts a seat list.
4. `POST /v1/public/shows/{id}/holds` creates a pending order and returns the
   order id, expiry, seats, and a one-time `order_token`.
5. `BookingFlow` saves the token in `sessionStorage` and shows a countdown.
6. Checkout sends name, phone, email, and `payment_method=bank_transfer`.
7. The Go service sets the source to `manual`, stores buyer details, and extends
   the hold to `MANUAL_HOLD_HOURS` (24).
8. The confirmation page polls `GET /v1/public/orders/{id}` with
   `X-Order-Token`.
9. The mock shows Site Settings bank details. In the current Go API, bank
   details are not yet in the public order response.
10. Staff mark the order paid through the internal API.
11. The order becomes `paid`, seats become `sold`, tickets are issued.
12. The confirmation page sees the paid status and shows one ticket link per
    seat.

## 2. Buyer pays through PayHere (planned)

1. Checkout sends `payment_method=payhere`.
2. The service keeps the source `online`, stores buyer details, and extends the
   hold to `CHECKOUT_HOLD_MINUTES` (15).
3. The server builds a PayHere checkout hash and redirects to sandbox.
4. PayHere returns the buyer to the confirmation page.
5. The return URL does **not** mark the order paid. The page keeps polling.
6. PayHere calls `/payhere/notify`, which verifies the signature, merchant,
   amount, and currency, records the payment, and calls `MarkPaid`.
7. If seats were taken after expiry, the order becomes `needs_review`.

Current code has config guards only; no PayHere handler exists yet.

## 3. Notify callback (planned)

The callback must be idempotent on `provider_payment_id`, respond quickly, and
never log secrets. It checks the stored amount, not the client's amount. A
duplicate success notification must not issue duplicate tickets. The notify route
is not mounted in the current server.

## 4. Staff marks an order paid

1. Staff call `POST /v1/internal/orders/{id}/mark-paid` with the service token.
2. `Orders.MarkPaid` validates the optional payment method.
3. The store locks the order, tries to claim the seats if the order is expired,
   marks the order paid, sells the seats, and inserts one ticket per seat.
4. If any seat is gone, the order becomes `needs_review`.
5. Calling the endpoint again is idempotent for an already-paid order.

## 5. Hold expires

1. The claim query treats a hold as available when `held_until < now()`, so a
   new buyer can reclaim it without waiting for the sweeper.
2. Every 30 seconds the sweeper asks the store to expire overdue pending orders.
3. The store takes `pg_try_advisory_xact_lock`, marks expired orders, and frees
   only seats still pointing at those expired orders.
4. If another order already took the seat, the seat's `order_id` points
   elsewhere and is left alone.
5. The buyer's page sees `expired` on its next poll and tells them to choose
   again.

## Related

- [Ticketing walkthrough](ticketing-walkthrough.md)
- [State machines](../architecture/state-machines.md)
- [Payments](../architecture/payments.md)
