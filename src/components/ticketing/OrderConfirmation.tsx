'use client'

import Link from 'next/link'
import { useCallback, useEffect, useState, useSyncExternalStore } from 'react'

import { formatLongShowDate } from '@/lib/format'
import {
  getTicketingClient,
  usingMockTicketing,
} from '@/lib/ticketing/client'
import { formatPrice } from '@/lib/ticketing/format'
import { readOrderToken } from '@/lib/ticketing/order-session'
import type {
  OrderSeat,
  PublicOrder,
  ShowSummary,
} from '@/lib/ticketing/types'

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

function orderSeatLabel(seat: OrderSeat): string {
  return seat.label
    ? `${seat.label} (${seat.tier_name})`
    : `${seat.tier_name} admission`
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
  const [markingPaid, setMarkingPaid] = useState(false)
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

  async function markPaid() {
    if (!token || !client.devMarkPaid) {
      return
    }

    setMarkingPaid(true)
    try {
      setOrder(await client.devMarkPaid(orderID, token))
      setError(null)
    } catch (nextError) {
      setError(
        nextError instanceof Error
          ? nextError.message
          : 'We could not mark this order paid.',
      )
    } finally {
      setMarkingPaid(false)
    }
  }

  const canMarkPaid =
    usingMockTicketing &&
    Boolean(client.devMarkPaid) &&
    order?.status === 'pending' &&
    order.payment_method === 'bank_transfer'

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
                    <dd>{order.seats.map(orderSeatLabel).join(', ')}</dd>
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

                {order.payment_method === 'bank_transfer' &&
                order.bank_transfer ? (
                  <div className="tk-bank">
                    <h3>Pay by bank transfer</h3>
                    <p className="tk-muted">
                      Transfer {formatPrice(order.total_cents, order.currency)}{' '}
                      using the reference{' '}
                      <strong>{shortReference(order.order_id)}</strong>.
                    </p>
                    <dl className="tk-details">
                      <div>
                        <dt>Bank</dt>
                        <dd>{order.bank_transfer.bank_name}</dd>
                      </div>
                      <div>
                        <dt>Account name</dt>
                        <dd>{order.bank_transfer.account_name}</dd>
                      </div>
                      <div>
                        <dt>Account number</dt>
                        <dd>{order.bank_transfer.account_number}</dd>
                      </div>
                      <div>
                        <dt>Branch</dt>
                        <dd>{order.bank_transfer.branch}</dd>
                      </div>
                      <div>
                        <dt>Pay by</dt>
                        <dd>{formatLongShowDate(order.expires_at)}</dd>
                      </div>
                    </dl>
                    {order.bank_transfer.instructions ? (
                      <p className="tk-muted">
                        {order.bank_transfer.instructions}
                      </p>
                    ) : null}
                    {canMarkPaid ? (
                      <div className="tk-actions">
                        <button
                          className="tk-button tk-button--secondary"
                          disabled={markingPaid}
                          onClick={() => void markPaid()}
                          type="button"
                        >
                          {markingPaid
                            ? 'Marking paid…'
                            : 'Mark paid (dev)'}
                        </button>
                      </div>
                    ) : null}
                  </div>
                ) : null}

                {order.status === 'pending' ? (
                  <p className="tk-notice">
                    Your order is reserved. This page refreshes automatically
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

                        let href: string | null = null
                        try {
                          href = client.ticketHref(order, ticket)
                        } catch {
                          // The link appears as soon as the provider returns it.
                        }

                        const label =
                          seat?.label ??
                          (seat
                            ? `${seat.tier_name} admission`
                            : ticket.seat_code)

                        return (
                          <li key={ticket.ticket_id}>
                            {href ? (
                              <Link href={href}>
                                <span>
                                  <strong>{label}</strong>
                                  <span>{seat?.tier_name}</span>
                                </span>
                                <span>View ticket →</span>
                              </Link>
                            ) : (
                              <span className="tk-picker-row">
                                <span>
                                  <strong>{label}</strong>
                                  <span>{seat?.tier_name}</span>
                                </span>
                                <span>Link pending</span>
                              </span>
                            )}
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
