'use client'

import Link from 'next/link'
import { useCallback, useEffect, useState, useSyncExternalStore } from 'react'

import { formatLongShowDate } from '@/lib/format'
import { getTicketingClient } from '@/lib/ticketing/client'
import { formatPrice } from '@/lib/ticketing/format'
import { readOrderToken } from '@/lib/ticketing/order-session'
import type { PublicOrder, ShowSummary } from '@/lib/ticketing/types'

import './ticketing.css'

type OrderConfirmationProps = {
  orderID: string
  productionSlug: string
  productionTitle: string
}

const finalStatuses: PublicOrder['status'][] = [
  'paid',
  'expired',
  'cancelled',
  'needs_review',
]

function statusLabel(status: PublicOrder['status']): string {
  switch (status) {
    case 'paid':
      return 'Paid'
    case 'expired':
      return 'Expired'
    case 'cancelled':
      return 'Cancelled'
    case 'needs_review':
      return 'Checking payment'
    default:
      return 'Reserved'
  }
}

function statusTone(status: PublicOrder['status']): string {
  return status === 'paid' ? 'tk-badge tk-badge--ok' : 'tk-badge tk-badge--warn'
}

function shortReference(orderID: string): string {
  return orderID.replace(/-/g, '').slice(-6).toUpperCase()
}

const subscribeToOrderToken = () => () => {}
const getServerOrderToken = () => null

// The order token lives in sessionStorage, which is an external store. Reading
// it through useSyncExternalStore keeps the server snapshot (no token) and the
// client snapshot in sync without a setState-in-effect round trip.
function useOrderToken(orderID: string): string | null {
  return useSyncExternalStore(
    subscribeToOrderToken,
    () => readOrderToken(orderID),
    getServerOrderToken,
  )
}

export function OrderConfirmation({
  orderID,
  productionSlug,
  productionTitle,
}: OrderConfirmationProps) {
  const client = getTicketingClient()
  const token = useOrderToken(orderID)
  const [order, setOrder] = useState<PublicOrder | null>(null)
  const [summary, setSummary] = useState<ShowSummary | null>(null)
  const [error, setError] = useState<string | null>(null)
  const showID = order?.show_id

  useEffect(() => {
    if (!showID) {
      return
    }

    let active = true
    void client
      .getShowSummary(showID)
      .then((next) => {
        if (active) {
          setSummary(next)
        }
      })
      .catch(() => {
        // The order is still useful without the extra show details.
      })

    return () => {
      active = false
    }
  }, [client, showID])

  const load = useCallback(async () => {
    if (!token) {
      return
    }

    try {
      const next = await client.getOrder(orderID, token)
      setOrder(next)
      setError(null)
    } catch (nextError) {
      setError(
        nextError instanceof Error
          ? nextError.message
          : 'We could not load this order.',
      )
    }
  }, [client, orderID, token])

  useEffect(() => {
    if (!token) {
      return
    }

    const orderToken = token
    let active = true
    let timer: number | undefined

    function schedule() {
      timer = window.setTimeout(poll, 3_000)
    }

    function apply(next: PublicOrder) {
      if (!active) {
        return
      }

      setOrder(next)
      setError(null)

      if (!finalStatuses.includes(next.status)) {
        schedule()
      }
    }

    function fail(nextError: unknown) {
      if (!active) {
        return
      }

      setError(
        nextError instanceof Error
          ? nextError.message
          : 'We could not load this order.',
      )
      schedule()
    }

    function poll() {
      client.getOrder(orderID, orderToken).then(apply).catch(fail)
    }

    poll()

    return () => {
      active = false
      if (timer !== undefined) {
        window.clearTimeout(timer)
      }
    }
  }, [client, orderID, token])

  return (
    <div className="tk">
      <div className="tk-layout">
        <div>
          <section className="tk-panel">
            <p className="tk-eyebrow">Booking confirmation</p>
            <h2>{summary?.title || productionTitle}</h2>
            {summary ? (
              <p className="tk-muted">
                {formatLongShowDate(summary.starts_at)}
                {summary.venue ? ` · ${summary.venue}` : ''}
              </p>
            ) : token ? (
              <p className="tk-muted">Loading show details…</p>
            ) : null}

            {!token ? (
              <p className="tk-notice" role="status">
                We cannot find this order in this browser. If you completed a
                booking, open the confirmation link from the same tab or contact
                the box office with reference{' '}
                <strong>{shortReference(orderID)}</strong>.
              </p>
            ) : null}

            {error ? (
              <p className="tk-notice" role="status">
                {error}{' '}
                <button
                  className="tk-button tk-button--link"
                  onClick={() => void load()}
                  type="button"
                >
                  Try again
                </button>
              </p>
            ) : null}

            {!order && token && !error ? (
              <p className="tk-muted">Loading your order…</p>
            ) : null}

            {order ? (
              <>
                <p>
                  <span className={statusTone(order.status)}>
                    {statusLabel(order.status)}
                  </span>
                </p>

                <dl className="tk-details">
                  <div>
                    <dt>Order reference</dt>
                    <dd>{shortReference(order.order_id)}</dd>
                  </div>
                  <div>
                    <dt>Seats</dt>
                    <dd>
                      {order.seats
                        .map(
                          (seat) =>
                            `${seat.label ?? seat.code} (${seat.tier_name})`,
                        )
                        .join(', ')}
                    </dd>
                  </div>
                  <div>
                    <dt>Total</dt>
                    <dd>{formatPrice(order.total_cents, order.currency)}</dd>
                  </div>
                  <div>
                    <dt>Status</dt>
                    <dd>{statusLabel(order.status)}</dd>
                  </div>
                </dl>

                {order.status === 'pending' ? (
                  <p className="tk-notice">
                    Your seats are reserved. This page refreshes automatically
                    while we wait for payment confirmation.
                  </p>
                ) : null}

                {order.status === 'needs_review' ? (
                  <p className="tk-notice">
                    We received your payment but one of the seats was taken.
                    The box office will contact you to arrange a refund or
                    another seat.
                  </p>
                ) : null}

                {order.status === 'expired' ? (
                  <p className="tk-notice">
                    This hold expired before payment was confirmed. No seats are
                    reserved.
                  </p>
                ) : null}

                {order.status === 'paid' ? (
                  <div>
                    <h3>Your tickets</h3>
                    <p className="tk-muted">
                      Show this page at the door. Each ticket has its own QR
                      code.
                    </p>
                    <ul className="tk-picker">
                      {order.tickets.map((ticket) => {
                        const seat = order.seats.find(
                          (candidate) => candidate.code === ticket.seat_code,
                        )

                        return (
                          <li key={ticket.ticket_id}>
                            <Link href={client.ticketHref(order, ticket)}>
                              <span>
                                <strong>
                                  {seat?.label ?? ticket.seat_code}
                                </strong>
                                <span>{seat?.tier_name}</span>
                              </span>
                              <span>View ticket →</span>
                            </Link>
                          </li>
                        )
                      })}
                    </ul>
                  </div>
                ) : null}
              </>
            ) : null}

            <div className="tk-actions">
              <Link className="tk-button" href={`/events/${productionSlug}`}>
                Back to event
              </Link>
              <Link
                className="tk-button tk-button--secondary"
                href={`/events/${productionSlug}/book`}
              >
                Book another performance
              </Link>
            </div>
          </section>
        </div>

        <aside className="tk-summary">
          <section className="tk-panel">
            <h2>Keep this reference</h2>
            <p className="tk-muted">
              Quote <strong>{shortReference(orderID)}</strong> if you contact us
              about this booking.
            </p>
          </section>
        </aside>
      </div>
    </div>
  )
}
