import { formatLongShowDate } from '@/lib/format'
import type { EventShow, EventView } from '@/lib/event-view'

// The single source of truth for every Reserve/Book link. With ticketing on,
// point at the site's booking route; otherwise use the WhatsApp fallback.
export function bookingHref(
  event: EventView,
  show: EventShow | null | undefined,
): string | null {
  if (!show || show.status === 'cancelled' || show.status === 'past') {
    return null
  }

  if (event.ticketing.active) {
    return show.id
      ? `/events/${event.slug}/book?show=${show.id}`
      : null
  }

  return whatsAppHref(event, show)
}

function whatsAppHref(
  event: EventView,
  show?: EventShow | null,
): string | null {
  const number = event.whatsappNumber?.replace(/\D/g, '')

  if (!number) {
    return null
  }

  const date = show ? ` on ${formatLongShowDate(show.startsAt)}` : ''
  const message = `Hi ${event.presentedBy ?? 'Basement Plays'}! I would like to book tickets for ${event.title}${date}.`

  return `https://wa.me/${number}?text=${encodeURIComponent(message)}`
}

export function isExternalBookingHref(href: string): boolean {
  return href.startsWith('http')
}
