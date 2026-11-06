'use client'

import { useRouter } from 'next/navigation'
import {
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
  type FormEvent,
} from 'react'

import type { EventShow, EventView } from '@/lib/event-view'
import { formatShowDate } from '@/lib/format'
import { getTicketingClient } from '@/lib/ticketing/client'
import {
  formatCountdown,
  formatPrice,
  secondsUntil,
} from '@/lib/ticketing/format'
import { rememberShow } from '@/lib/ticketing/mock'
import {
  clearOrderToken,
  saveOrderToken,
} from '@/lib/ticketing/order-session'
import type {
  Hold,
  HoldRequest,
  PaymentMode,
  SeatMap as SeatMapData,
  ShowSummary,
  TierSummary,
} from '@/lib/ticketing/types'
import { MAX_SEATS_PER_ORDER, TicketingError } from '@/lib/ticketing/types'

import { SeatMap } from './SeatMap'
import { StageBar } from './StageBar'

import './ticketing.css'

type BookingFlowProps = {
  event: EventView
  maxSeats?: number
  show: EventShow
}

type CheckoutPaymentMethod = 'payhere' | 'bank_transfer'

function idempotencyKey(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }

  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
}

function messageFor(error: unknown): string {
  if (error instanceof TicketingError) {
    return error.message
  }

  if (error instanceof Error) {
    return error.message
  }

  return 'Something went wrong. Please try again.'
}

function paymentDetails(paymentMode: PaymentMode): {
  help: string
  label: string
  method: CheckoutPaymentMethod | null
} {
  if (paymentMode === 'online') {
    return {
      help: 'The sandbox payment completes immediately in demo mode.',
      label: 'Pay online (sandbox)',
      method: 'payhere',
    }
  }

  if (paymentMode === 'bank_transfer') {
    return {
      help: 'Your tickets are held for 24 hours while we confirm the transfer. The bank details and deadline appear on the confirmation page.',
      label: 'Bank transfer',
      method: 'bank_transfer',
    }
  }

  return {
    help: 'This event is not accepting online bookings. Contact the box office on WhatsApp.',
    label: 'Not available',
    method: null,
  }
}

export function BookingFlow({
  event,
  maxSeats = MAX_SEATS_PER_ORDER,
  show,
}: BookingFlowProps) {
  const router = useRouter()
  const client = getTicketingClient()
  const showId = show.id
  const productionSlug = event.slug
  const showTitle = event.title
  const startsAt = show.startsAt
  const venue = show.venue

  const nameID = useId()
  const phoneID = useId()
  const emailID = useId()

  const [summary, setSummary] = useState<ShowSummary | null>(null)
  const [map, setMap] = useState<SeatMapData | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [quantities, setQuantities] = useState<Record<string, number>>({})
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState<string | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  const [hold, setHold] = useState<Hold | null>(null)
  const holdToken = useRef<string | null>(null)
  const [remaining, setRemaining] = useState(0)

  const [buyerName, setBuyerName] = useState('')
  const [buyerPhone, setBuyerPhone] = useState('')
  const [buyerEmail, setBuyerEmail] = useState('')

  const selectionMode =
    summary?.selection_mode ?? event.ticketing.selectionMode
  const stagePosition =
    summary?.stage_position ?? event.ticketing.stagePosition
  const payment = paymentDetails(
    summary?.payment_mode ?? event.ticketing.paymentMode,
  )

  useEffect(() => {
    rememberShow(showId, {
      bankTransfer: event.bankTransfer
        ? {
            account_name: event.bankTransfer.accountName ?? '',
            account_number: event.bankTransfer.accountNumber ?? '',
            bank_name: event.bankTransfer.bankName ?? '',
            branch: event.bankTransfer.branch ?? '',
            instructions: event.bankTransfer.instructions ?? '',
          }
        : null,
      columns: event.ticketing.columns,
      numbering: event.ticketing.numbering,
      paymentMode: event.ticketing.paymentMode,
      productionSlug,
      selectionMode: event.ticketing.selectionMode,
      stagePosition: event.ticketing.stagePosition,
      startsAt,
      tiers: event.ticketing.tiers.map((tier) => ({
        count: tier.count,
        id: tier.id,
        name: tier.name,
        price_cents: tier.price * 100,
      })),
      title: showTitle,
      venue,
    })
  }, [
    event.bankTransfer,
    event.ticketing.columns,
    event.ticketing.numbering,
    event.ticketing.paymentMode,
    event.ticketing.selectionMode,
    event.ticketing.stagePosition,
    event.ticketing.tiers,
    productionSlug,
    showId,
    showTitle,
    startsAt,
    venue,
  ])

  const applyAvailability = useCallback(
    (nextSummary: ShowSummary, nextMap?: SeatMapData) => {
      setSummary(nextSummary)
      setLoadError(null)

      if (nextSummary.selection_mode === 'seats') {
        if (nextMap) {
          setMap(nextMap)
          setSelected((current) =>
            current.filter(
              (code) =>
                nextMap.seats.find((seat) => seat.code === code)?.status ===
                'available',
            ),
          )
        }
        return
      }

      setMap(null)
      setQuantities((current) => {
        const next: Record<string, number> = {}

        for (const [tierID, quantity] of Object.entries(current)) {
          const tier = nextSummary.tiers.find(
            (candidate) => candidate.tier_id === tierID,
          )
          const safe = Math.min(quantity, tier?.seats_remaining ?? 0)

          if (safe > 0) {
            next[tierID] = safe
          }
        }

        return next
      })
    },
    [],
  )

  const refresh = useCallback(async () => {
    try {
      const next = await client.getShowSummary(showId)
      const nextMap =
        next.selection_mode === 'seats'
          ? await client.getSeatMap(showId)
          : undefined

      applyAvailability(next, nextMap)
    } catch (error) {
      setLoadError(messageFor(error))
    } finally {
      setLoading(false)
    }
  }, [applyAvailability, client, showId])

  useEffect(() => {
    let active = true

    client
      .getShowSummary(showId)
      .then(async (next) => {
        const nextMap =
          next.selection_mode === 'seats'
            ? await client.getSeatMap(showId)
            : undefined

        if (active) {
          applyAvailability(next, nextMap)
        }
      })
      .catch((error) => {
        if (active) {
          setLoadError(messageFor(error))
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })

    return () => {
      active = false
    }
  }, [applyAvailability, client, showId])

  useEffect(() => {
    if (hold) {
      return
    }

    const timer = window.setInterval(() => {
      void refresh()
    }, 10_000)

    return () => window.clearInterval(timer)
  }, [hold, refresh])

  useEffect(() => {
    if (!hold) {
      return
    }

    const tick = () => {
      const left = secondsUntil(hold.expires_at)
      setRemaining(left)

      if (left === 0) {
        holdToken.current = null
        clearOrderToken(hold.order_id)
        setHold(null)
        setSelected([])
        setQuantities({})
        setNotice('Your seat hold expired. Choose your seats again.')
        void refresh()
      }
    }

    tick()
    const timer = window.setInterval(tick, 1_000)

    return () => window.clearInterval(timer)
  }, [hold, refresh])

  function toggleSeat(code: string) {
    if (hold) {
      return
    }

    setNotice(null)
    setSelected((current) => {
      if (current.includes(code)) {
        return current.filter((value) => value !== code)
      }
      if (current.length >= maxSeats) {
        setNotice(`You can reserve up to ${maxSeats} seats at a time.`)
        return current
      }

      return [...current, code]
    })
  }

  function changeQuantity(tier: TierSummary, delta: number) {
    if (hold || busy) {
      return
    }

    setNotice(null)

    const currentQuantity = quantities[tier.tier_id] ?? 0
    const total = Object.values(quantities).reduce(
      (sum, quantity) => sum + quantity,
      0,
    )
    const limit = Math.min(tier.seats_remaining, maxSeats)
    const nextQuantity = Math.max(
      0,
      Math.min(currentQuantity + delta, limit),
    )

    if (delta > 0 && total >= maxSeats) {
      setNotice(`You can reserve up to ${maxSeats} tickets at a time.`)
      return
    }

    if (delta > 0 && nextQuantity === currentQuantity) {
      setNotice('There are no more tickets left in this tier.')
      return
    }

    const next = { ...quantities }
    if (nextQuantity > 0) {
      next[tier.tier_id] = nextQuantity
    } else {
      delete next[tier.tier_id]
    }

    setQuantities(next)
  }

  async function handleHold() {
    const request: HoldRequest =
      selectionMode === 'quantity'
        ? {
            tiers: Object.entries(quantities)
              .filter(([, quantity]) => quantity > 0)
              .map(([tier_id, quantity]) => ({ quantity, tier_id })),
          }
        : { seats: selected }

    const requested =
      'tiers' in request
        ? request.tiers.reduce((sum, item) => sum + item.quantity, 0)
        : request.seats.length

    if (requested === 0) {
      return
    }

    setBusy(true)
    setNotice(null)

    try {
      const result = await client.createHold(showId, request, {
        idempotencyKey: idempotencyKey(),
      })

      if (!result.order_token) {
        throw new Error('The hold response did not include an order token.')
      }

      holdToken.current = result.order_token
      saveOrderToken(result.order_id, result.order_token)
      setRemaining(secondsUntil(result.expires_at))
      setHold(result)
    } catch (error) {
      if (error instanceof TicketingError && error.code === 'seat_taken') {
        setNotice(
          'Someone else took one of those seats. We refreshed the map; your other choices are still selected.',
        )
        await refresh()
      } else if (
        error instanceof TicketingError &&
        error.code === 'insufficient_inventory'
      ) {
        setNotice(
          'Not enough tickets are left in those tiers. We refreshed availability; your other choices are still selected.',
        )
        await refresh()
      } else {
        setNotice(messageFor(error))
      }
    } finally {
      setBusy(false)
    }
  }

  async function releaseHold() {
    const current = hold
    const token = holdToken.current

    holdToken.current = null
    setHold(null)

    if (current && token) {
      clearOrderToken(current.order_id)
      try {
        await client.releaseOrder(current.order_id, token)
      } catch {
        // The hold may already have expired; the map refresh is the source of
        // truth either way.
      }
    }

    setNotice(
      selectionMode === 'quantity'
        ? 'Tickets released. You can choose again.'
        : 'Seats released. You can choose again.',
    )
    await refresh()
  }

  async function handleCheckout(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const orderToken = holdToken.current

    if (!hold || !orderToken || !payment.method) {
      return
    }
    if (remaining === 0) {
      setNotice('Your seat hold expired. Choose your seats again.')
      return
    }

    setBusy(true)
    setNotice(null)

    try {
      const order = await client.checkout(
        hold.order_id,
        orderToken,
        {
          buyer_name: buyerName,
          buyer_email: buyerEmail,
          buyer_phone: buyerPhone,
          payment_method: payment.method,
        },
        { idempotencyKey: idempotencyKey() },
      )

      saveOrderToken(order.order_id, orderToken)
      router.push(
        `/events/${productionSlug}/book/confirmation?order=${order.order_id}`,
      )
    } catch (error) {
      setNotice(messageFor(error))
    } finally {
      setBusy(false)
    }
  }

  const selectedSeats =
    map?.seats.filter((seat) => selected.includes(seat.code)) ?? []
  const quantityLines = summary
    ? summary.tiers
        .map((tier) => ({
          quantity: quantities[tier.tier_id] ?? 0,
          tier,
        }))
        .filter((line) => line.quantity > 0)
    : []
  const selectedCount =
    selectionMode === 'quantity'
      ? quantityLines.reduce((sum, line) => sum + line.quantity, 0)
      : selected.length
  const total = hold
    ? hold.total_cents
    : selectionMode === 'quantity'
      ? quantityLines.reduce(
          (sum, line) => sum + line.tier.price_cents * line.quantity,
          0,
        )
      : selectedSeats.reduce((sum, seat) => sum + seat.price_cents, 0)
  const currency = summary?.currency ?? 'LKR'
  const orderedTiers =
    stagePosition === 'back' && summary
      ? [...summary.tiers].reverse()
      : (summary?.tiers ?? [])

  return (
    <div className="tk">
      <div className="tk-layout">
        <div>
          {hold ? (
            <section className="tk-panel" aria-labelledby="details-heading">
              <h2 id="details-heading">Your details</h2>
              <p className="tk-muted">
                We only use these details to send your tickets and contact you
                about this booking.
              </p>

              <form onSubmit={handleCheckout}>
                <label className="tk-field" htmlFor={nameID}>
                  <span>Full name</span>
                  <input
                    autoComplete="name"
                    id={nameID}
                    maxLength={120}
                    onChange={(event) => setBuyerName(event.target.value)}
                    required
                    type="text"
                    value={buyerName}
                  />
                </label>

                <label className="tk-field" htmlFor={phoneID}>
                  <span>Phone (WhatsApp)</span>
                  <input
                    autoComplete="tel"
                    id={phoneID}
                    maxLength={32}
                    onChange={(event) => setBuyerPhone(event.target.value)}
                    required
                    type="tel"
                    value={buyerPhone}
                  />
                </label>

                <label className="tk-field" htmlFor={emailID}>
                  <span>Email (optional)</span>
                  <input
                    autoComplete="email"
                    id={emailID}
                    maxLength={254}
                    onChange={(event) => setBuyerEmail(event.target.value)}
                    type="email"
                    value={buyerEmail}
                  />
                </label>

                <div className="tk-payment-choice">
                  <span>Payment</span>
                  <strong>{payment.label}</strong>
                  <p className="tk-muted">{payment.help}</p>
                </div>

                <div className="tk-actions">
                  <button
                    className="tk-button"
                    disabled={busy || remaining === 0 || !payment.method}
                    type="submit"
                  >
                    {busy ? 'Confirming…' : 'Confirm booking'}
                  </button>
                  <button
                    className="tk-button tk-button--secondary"
                    disabled={busy}
                    onClick={() => void releaseHold()}
                    type="button"
                  >
                    {selectionMode === 'quantity'
                      ? 'Change tickets'
                      : 'Change seats'}
                  </button>
                </div>
              </form>
            </section>
          ) : (
            <section className="tk-panel" aria-labelledby="seats-heading">
              <h2 id="seats-heading">
                {selectionMode === 'quantity'
                  ? 'Choose your tickets'
                  : 'Choose your seats'}
              </h2>
              <p className="tk-muted">
                {selectionMode === 'quantity'
                  ? `Tickets are held for 10 minutes once you continue. Pick up to ${maxSeats}.`
                  : `Seats are held for 10 minutes once you continue. Pick up to ${maxSeats}.`}
              </p>

              {loading ? (
                <p className="tk-muted">Loading availability…</p>
              ) : null}
              {loadError ? (
                <p className="tk-error" role="alert">
                  {loadError}{' '}
                  <button
                    className="tk-button tk-button--link"
                    onClick={() => void refresh()}
                    type="button"
                  >
                    Try again
                  </button>
                </p>
              ) : null}

              {!loading && !loadError && selectionMode === 'seats' && map ? (
                <SeatMap
                  disabled={busy}
                  map={map}
                  onToggle={toggleSeat}
                  selected={selected}
                />
              ) : null}

              {!loading &&
              !loadError &&
              selectionMode === 'quantity' &&
              summary ? (
                <div className="tk-quantity">
                  <StageBar position={stagePosition} />

                  {orderedTiers.map((tier) => {
                    const quantity = quantities[tier.tier_id] ?? 0
                    const soldOut = tier.seats_remaining <= 0
                    const canAdd =
                      !soldOut &&
                      quantity < tier.seats_remaining &&
                      selectedCount < maxSeats

                    return (
                      <section
                        className={`tk-tier-box${
                          soldOut ? ' is-sold-out' : ''
                        }`}
                        key={tier.tier_id}
                      >
                        <div className="tk-tier-box__head">
                          <div>
                            <h3>{tier.name}</h3>
                            <p className="tk-tier-meta">
                              {formatPrice(tier.price_cents, currency)} ·{' '}
                              {tier.seats_remaining} of {tier.seats_total} left
                            </p>
                          </div>

                          {soldOut ? (
                            <p className="tk-tier-box__sold">Sold out</p>
                          ) : (
                            <div className="tk-stepper">
                              <button
                                aria-label={`Remove one ${tier.name} ticket`}
                                disabled={busy || quantity === 0}
                                onClick={() => changeQuantity(tier, -1)}
                                type="button"
                              >
                                −
                              </button>
                              <span aria-live="polite">{quantity}</span>
                              <button
                                aria-label={`Add one ${tier.name} ticket`}
                                disabled={busy || !canAdd}
                                onClick={() => changeQuantity(tier, 1)}
                                type="button"
                              >
                                +
                              </button>
                            </div>
                          )}
                        </div>
                      </section>
                    )
                  })}

                  {stagePosition === 'back' ? (
                    <StageBar position="back" />
                  ) : null}
                </div>
              ) : null}
            </section>
          )}
        </div>

        <aside className="tk-summary">
          <section className="tk-panel" aria-labelledby="summary-heading">
            <h2 id="summary-heading">Your order</h2>

            <p className="tk-muted">
              {showTitle} · {formatShowDate(startsAt)}
              {venue ? ` · ${venue}` : ''}
            </p>

            {hold ? (
              <p className="tk-countdown">
                {selectionMode === 'quantity' ? 'Tickets' : 'Seats'} held for{' '}
                <strong>{formatCountdown(remaining)}</strong>
              </p>
            ) : null}

            {notice ? (
              <p className="tk-notice" role="status">
                {notice}
              </p>
            ) : null}

            {hold ? (
              <ul className="tk-summary-list">
                {hold.seats.map((seat) => (
                  <li key={seat.code}>
                    <span>
                      {seat.label ? (
                        <>
                          Seat {seat.label}{' '}
                          <span className="tk-muted">{seat.tier_name}</span>
                        </>
                      ) : (
                        `${seat.tier_name} admission`
                      )}
                    </span>
                    <span>{formatPrice(seat.price_cents, currency)}</span>
                  </li>
                ))}
              </ul>
            ) : selectedCount === 0 ? (
              <p className="tk-muted">
                {selectionMode === 'quantity'
                  ? 'No tickets selected yet.'
                  : 'No seats selected yet.'}
              </p>
            ) : selectionMode === 'quantity' ? (
              <ul className="tk-summary-list">
                {quantityLines.map(({ quantity, tier }) => (
                  <li key={tier.tier_id}>
                    <span>
                      {quantity} × {tier.name}
                    </span>
                    <span>
                      {formatPrice(tier.price_cents * quantity, currency)}
                    </span>
                  </li>
                ))}
              </ul>
            ) : (
              <ul className="tk-summary-list">
                {selectedSeats.map((seat) => (
                  <li key={seat.code}>
                    <span>
                      {seat.label ?? seat.code}{' '}
                      <span className="tk-muted">{seat.tier_name}</span>
                    </span>
                    <span>{formatPrice(seat.price_cents, currency)}</span>
                  </li>
                ))}
              </ul>
            )}

            <div className="tk-total">
              <span>Total</span>
              <strong>{formatPrice(total, currency)}</strong>
            </div>

            {!hold ? (
              <div className="tk-actions">
                <button
                  className="tk-button"
                  disabled={busy || selectedCount === 0}
                  onClick={() => void handleHold()}
                  type="button"
                >
                  {busy
                    ? 'Holding…'
                    : selectionMode === 'quantity'
                      ? 'Hold tickets'
                      : 'Hold seats'}
                </button>
              </div>
            ) : null}
          </section>
        </aside>
      </div>
    </div>
  )
}
