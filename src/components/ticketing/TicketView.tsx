import { Container } from '@/components/ui/Container'
import { formatLongShowDate } from '@/lib/format'
import { formatPrice } from '@/lib/ticketing/format'
import type { Ticket } from '@/lib/ticketing/types'

import './ticketing.css'

type TicketViewProps = {
  qrSvg: string | null
  ticket: Ticket
}

export function TicketView({ qrSvg, ticket }: TicketViewProps) {
  const valid = ticket.status === 'issued'

  return (
    <div className="tk">
      <Container>
        <section className="tk-panel">
          <div className="tk-ticket">
            <div>
              <p className="tk-eyebrow">Your ticket</p>
              <h1>{ticket.show_title || 'Basement Plays ticket'}</h1>
              <p className="tk-muted">
                {formatLongShowDate(ticket.starts_at)}
                {ticket.venue ? ` · ${ticket.venue}` : ''}
              </p>

              <p>
                <span
                  className={
                    valid
                      ? 'tk-badge tk-badge--ok'
                      : 'tk-badge tk-badge--warn'
                  }
                >
                  {valid ? 'Valid' : ticket.status}
                </span>
              </p>

              <dl className="tk-details">
                <div>
                  <dt>Admission</dt>
                  <dd>{ticket.description}</dd>
                </div>
                <div>
                  <dt>Price</dt>
                  <dd>{formatPrice(ticket.price_cents)}</dd>
                </div>
                <div>
                  <dt>Order reference</dt>
                  <dd>{ticket.order_reference}</dd>
                </div>
                <div>
                  <dt>Ticket status</dt>
                  <dd>{ticket.status}</dd>
                </div>
              </dl>

              {!valid ? (
                <p className="tk-notice">
                  This ticket is {ticket.status} and cannot be admitted. Contact
                  the box office if you think this is a mistake.
                </p>
              ) : null}

              <p className="tk-muted">
                Present this QR code at the door. Each seat has its own ticket.
              </p>
            </div>

            <div className="tk-qr">
              {qrSvg ? (
                <div dangerouslySetInnerHTML={{ __html: qrSvg }} />
              ) : (
                <p>QR code unavailable</p>
              )}
              <p>{ticket.qr_url}</p>
            </div>
          </div>
        </section>
      </Container>
    </div>
  )
}
