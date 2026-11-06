import { EventCard } from '@/components/EventCard'
import { Container } from '@/components/ui/Container'
import { getNextShow, type EventShow, type EventView } from '@/lib/event-view'
import { getAllEventViews, getSiteSettings } from '@/lib/payload'

export const dynamic = 'force-dynamic'

export const metadata = {
  title: 'Events',
}

type ProductionListing = {
  event: EventView
  nextShow: EventShow | null
}

export default async function EventsPage() {
  const [events, settings] = await Promise.all([
    getAllEventViews(),
    getSiteSettings(),
  ])
  const listings: ProductionListing[] = events.map((event) => ({
    event,
    nextShow: getNextShow(event),
  }))

  listings.sort((a, b) => {
    if (a.nextShow && b.nextShow) {
      return (
        new Date(a.nextShow.startsAt).getTime() -
        new Date(b.nextShow.startsAt).getTime()
      )
    }

    if (a.nextShow) {
      return -1
    }

    if (b.nextShow) {
      return 1
    }

    return a.event.title.localeCompare(b.event.title)
  })

  const groupName = settings.groupName || 'Basement Plays'

  return (
    <Container className="py-16 sm:py-24">
      <p className="text-xs font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
        Programme
      </p>
      <div className="mt-4 flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-5xl tracking-tight text-[var(--text)] sm:text-7xl">
            Events
          </h1>
          <p className="mt-5 max-w-xl text-base leading-7 text-[var(--muted)]">
            Productions from {groupName}, with upcoming show dates and ticket
            links.
          </p>
        </div>
        {listings.length > 0 ? (
          <p className="text-[10px] uppercase tracking-[0.2em] text-[var(--muted)]">
            {listings.length} {listings.length === 1 ? 'production' : 'productions'}
          </p>
        ) : null}
      </div>

      {listings.length > 0 ? (
        <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {listings.map(({ event, nextShow }) => (
            <EventCard
              event={event}
              groupName={groupName}
              key={event.id}
              nextShow={nextShow}
            />
          ))}
        </div>
      ) : (
        <p className="mt-12 border-t border-[var(--border)] pt-8 text-sm text-[var(--muted)]">
          No productions have been published yet.
        </p>
      )}
    </Container>
  )
}
