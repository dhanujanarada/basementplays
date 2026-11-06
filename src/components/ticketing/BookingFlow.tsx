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
  SeatMap as SeatMapData,
} from '@/lib/ticketing/types'
import { TicketingError } from '@/lib/ticketing/types'

import { SeatMap } from './SeatMap'

import './ticketing.css'

type BookingFlowProps = {
  maxSeats?: number
  productionSlug: string
  showId: string
  showTitle: string
  startsAt: string
  venue: string
}

type PaymentMethod = 'payhere' | 'bank_transfer'

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

export function BookingFlow({
  maxSeats = 6,
  productionSlug,
  showId,
  showTitle,
  startsAt,
  venue,
}: BookingFlowProps) {
  const router = useRouter()
  const client = getTicketingClient()

  const nameID = useId()
  const phoneID = useId()
  const emailID = useId()
  const paymentID = useId()

  const [map, setMap] = useState<SeatMapData | null>(null)
  const [selected, setSelected] = useState<string[]>([])
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
  const [paymentMethod, setPaymentMethod] =
    useState<PaymentMethod>('payhere')

  useEffect(() => {
    rememberShow(showId, {
      productionSlug,
      startsAt,
      title: showTitle,
      venue,
    })
  }, [productionSlug, showId, showTitle, startsAt, venue])

  const applyMap = useCallback((next: SeatMapData) => {
    setMap(next)
    setLoadError(null)
    setSelected((current) =>
      current.filter(
        (code) =>
          next.seats.find((seat) => seat.code === code)?.status ===
          'available',
      ),
    )
  }, [])

  const refresh = useCallback(async () => {
    try {
      const next = await client.getSeatMap(showId)
      applyMap(next)
    } catch (error) {
      setLoadError(messageFor(error))
    } finally {
      setLoading(false)
    }
  }, [applyMap, client, showId])

  useEffect(() => {
    let active = true

    client
      .getSeatMap(showId)
      .then((next) => {
        if (active) {
          applyMap(next)
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
  }, [applyMap, client, showId])

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

  async function handleHold() {
    if (selected.length === 0) {
      return
    }

    setBusy(true)
    setNotice(null)

    try {
      const result = await client.createHold(showId, selected, {
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
      if (
        error instanceof TicketingError &&
        error.code === 'seat_taken'
      ) {
        setNotice(
          'Someone else took one of those seats. We refreshed the map; your other choices are still selected.',
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

    setNotice('Seats released. You can choose again.')
    await refresh()
  }

  async function handleCheckout(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const orderToken = holdToken.current

    if (!hold || !orderToken) {
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
          payment_method: paymentMethod,
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
  const total = selectedSeats.reduce(
    (sum, seat) => sum + seat.price_cents,
    0,
  )

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

                <label className="tk-field" htmlFor={paymentID}>
                  <span>Payment</span>
                  <select
                    id={paymentID}
                    onChange={(event) =>
                      setPaymentMethod(
                        event.target.value as PaymentMethod,
                      )
                    }
                    value={paymentMethod}
                  >
                    <option value="payhere">
                      Pay online (sandbox)
                    </option>
                    <option value="bank_transfer">
                      Reserve and pay by bank transfer
                    </option>
                  </select>
                </label>

                <p className="tk-muted">
                  {paymentMethod === 'payhere'
                    ? 'The sandbox payment completes immediately in demo mode.'
                    : 'Your seats are held for 24 hours while we confirm the transfer.'}
                </p>

                <div className="tk-actions">
                  <button
                    className="tk-button"
                    disabled={busy || remaining === 0}
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
                    Change seats
                  </button>
                </div>
              </form>
            </section>
          ) : (
            <section className="tk-panel" aria-labelledby="seats-heading">
              <h2 id="seats-heading">Choose your seats</h2>
              <p className="tk-muted">
                Seats are held for 10 minutes once you continue. Pick up to{' '}
                {maxSeats}.
              </p>

              {loading ? <p className="tk-muted">Loading seats…</p> : null}
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
              {map && !loadError ? (
                <SeatMap
                  disabled={busy}
                  map={map}
                  onToggle={toggleSeat}
                  selected={selected}
                />
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
                Seats held for <strong>{formatCountdown(remaining)}</strong>
              </p>
            ) : null}

            {notice ? (
              <p className="tk-notice" role="status">
                {notice}
              </p>
            ) : null}

            {selectedSeats.length === 0 ? (
              <p className="tk-muted">No seats selected yet.</p>
            ) : (
              <ul className="tk-summary-list">
                {selectedSeats.map((seat) => (
                  <li key={seat.code}>
                    <span>
                      {seat.label ?? seat.code}{' '}
                      <span className="tk-muted">{seat.tier_name}</span>
                    </span>
                    <span>
                      {formatPrice(seat.price_cents, map?.currency)}
                    </span>
                  </li>
                ))}
              </ul>
            )}

            <div className="tk-total">
              <span>Total</span>
              <strong>{formatPrice(total, map?.currency)}</strong>
            </div>

            {!hold ? (
              <div className="tk-actions">
                <button
                  className="tk-button"
                  disabled={busy || selected.length === 0}
                  onClick={() => void handleHold()}
                  type="button"
                >
                  {busy ? 'Holding…' : 'Hold seats'}
                </button>
              </div>
            ) : null}
          </section>
        </aside>
      </div>
    </div>
  )
}
