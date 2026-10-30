import { EventCard } from '@/components/EventCard'
import { Container } from '@/components/ui/Container'
import type { Production } from '@/payload-types'
import {
  getAllProductions,
  getRunsForProduction,
  getSiteSettings,
} from '@/lib/payload'
import { getUpcomingShow } from '@/lib/shows'
import type { ShowEntry } from '@/lib/types'

export const dynamic = 'force-dynamic'

export const metadata = {
  title: 'Events',
}

type ProductionListing = {
  nextShow: ShowEntry | null
  production: Production
}

export default async function EventsPage() {
  const [productions, settings] = await Promise.all([
    getAllProductions(),
    getSiteSettings(),
  ])
  const listings: ProductionListing[] = await Promise.all(
    productions.map(async (production) => {
      const runs = await getRunsForProduction(production.id)

      return {
        nextShow: getUpcomingShow(runs),
        production,
      }
    }),
  )

  listings.sort((a, b) => {
    if (a.nextShow && b.nextShow) {
      return (
        new Date(a.nextShow.show.startsAt).getTime() -
        new Date(b.nextShow.show.startsAt).getTime()
      )
    }

    if (a.nextShow) {
      return -1
    }

    if (b.nextShow) {
      return 1
    }

    return a.production.title.localeCompare(b.production.title)
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
          {listings.map(({ nextShow, production }) => (
            <EventCard
              groupName={groupName}
              key={production.id}
              nextShow={nextShow}
              production={production}
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
