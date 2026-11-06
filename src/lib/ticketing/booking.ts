import type { Production, Run } from '@/payload-types'

import type { Show } from '@/lib/types'

export type BookableShow = {
  run: Run
  show: Show
}

export function isTicketingEnabled(show: Show | null | undefined): boolean {
  return show?.ticketingEnabled === true
}

export function isBookableShow(show: Show | null | undefined): boolean {
  if (!show) {
    return false
  }

  if (show.status === 'sold-out' || show.status === 'cancelled') {
    return false
  }

  return Boolean(show.ticketUrl) || isTicketingEnabled(show)
}

// The single source of truth for every Reserve/Book link. With ticketing on,
// point at the site's booking route; otherwise keep the existing external
// ticket URL (currently WhatsApp).
export function bookingHref(
  production: Production,
  show: Show | null | undefined,
): string | null {
  if (!show) {
    return null
  }

  if (show.status === 'sold-out' || show.status === 'cancelled') {
    return null
  }

  if (isTicketingEnabled(show)) {
    return show.id
      ? `/events/${production.slug}/book?show=${show.id}`
      : null
  }

  return show.ticketUrl ?? null
}

function showTime(show: Show): number {
  return new Date(show.startsAt).getTime()
}

export function getBookableShows(
  runs: Run[],
  now: Date = new Date(),
): BookableShow[] {
  const bookable: BookableShow[] = []

  for (const run of runs) {
    for (const show of run.shows ?? []) {
      const startsAt = showTime(show)

      if (
        !show.id ||
        !isTicketingEnabled(show) ||
        show.status === 'sold-out' ||
        show.status === 'cancelled' ||
        Number.isNaN(startsAt) ||
        startsAt < now.getTime()
      ) {
        continue
      }

      bookable.push({ run, show })
    }
  }

  return bookable.sort((a, b) => showTime(a.show) - showTime(b.show))
}
