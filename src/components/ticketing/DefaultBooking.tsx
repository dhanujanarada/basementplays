import { Container } from '@/components/ui/Container'
import { formatLongShowDate } from '@/lib/format'
import { MAX_SEATS_PER_ORDER } from '@/lib/ticketing/types'
import type { ProductionBookingProps } from '@/lib/types'

import { BookingFlow } from './BookingFlow'

export function DefaultBooking({
  production,
  run,
  show,
}: ProductionBookingProps) {
  return (
    <>
      <header className="tk-booking-header">
        <Container>
          <p className="tk-eyebrow">Reserve seats</p>
          <h1>{production.title}</h1>
          <p>
            {formatLongShowDate(show.startsAt)}
            {run.venue ? ` · ${run.venue}` : ''}
          </p>
        </Container>
      </header>

      <Container>
        <BookingFlow
          maxSeats={MAX_SEATS_PER_ORDER}
          productionSlug={production.slug}
          showId={show.id ?? ''}
          showTitle={production.title}
          startsAt={show.startsAt}
          venue={run.venue ?? ''}
        />
      </Container>
    </>
  )
}
