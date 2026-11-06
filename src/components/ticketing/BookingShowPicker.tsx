import Link from 'next/link'

import { Container } from '@/components/ui/Container'
import type { EventShow, EventView } from '@/lib/event-view'
import { showSeasonLabel } from '@/lib/event-view'
import { formatShowDate } from '@/lib/format'

import './ticketing.css'

type BookingShowPickerProps = {
  event: EventView
  shows: EventShow[]
}

export function BookingShowPicker({
  event,
  shows,
}: BookingShowPickerProps) {
  return (
    <>
      <header className="tk-booking-header">
        <Container>
          <p className="tk-eyebrow">Reserve seats</p>
          <h1>{event.title}</h1>
          <p>Choose a performance to see available seats.</p>
        </Container>
      </header>

      <Container>
        <div className="tk">
          <div className="tk-panel">
            <h2>Which performance?</h2>

            {shows.length === 0 ? (
              <p className="tk-muted">
                There are no performances on sale right now. Check the event page
                for updates.
              </p>
            ) : (
              <ul className="tk-picker">
                {shows.map((show) => (
                  <li key={show.id}>
                    <Link href={`/events/${event.slug}/book?show=${show.id}`}>
                      <span>
                        <strong>{formatShowDate(show.startsAt)}</strong>
                        <span>
                          {[showSeasonLabel(show.startsAt), show.venue]
                            .filter(Boolean)
                            .join(' · ')}
                        </span>
                      </span>
                      <span>Choose seats →</span>
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </Container>
    </>
  )
}
