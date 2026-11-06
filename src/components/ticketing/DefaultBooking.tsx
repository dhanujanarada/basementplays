import { Container } from '@/components/ui/Container'
import { formatLongShowDate } from '@/lib/format'
import { MAX_SEATS_PER_ORDER } from '@/lib/ticketing/types'
import type { EventBookingProps } from '@/lib/types'

import { BookingFlow } from './BookingFlow'

export function DefaultBooking({ event, show }: EventBookingProps) {
  return (
    <>
      <header className="tk-booking-header">
        <Container>
          <p className="tk-eyebrow">Reserve seats</p>
          <h1>{event.title}</h1>
          <p>
            {formatLongShowDate(show.startsAt)}
            {show.venue ? ` · ${show.venue}` : ''}
          </p>
        </Container>
      </header>

      <Container>
        <BookingFlow
          event={event}
          maxSeats={MAX_SEATS_PER_ORDER}
          show={show}
        />
      </Container>
    </>
  )
}
