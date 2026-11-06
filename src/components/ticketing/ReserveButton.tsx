import { Button } from '@/components/ui/Button'
import { formatStatus } from '@/lib/format'
import type { EventShow, EventView } from '@/lib/event-view'
import {
  bookingHref,
  isExternalBookingHref,
} from '@/lib/ticketing/booking'

type ReserveButtonProps = {
  className?: string
  event: EventView
  // The design's existing label, kept when the event falls back to WhatsApp.
  label: string
  show: EventShow | null | undefined
  variant?: 'primary' | 'secondary' | 'ghost'
}

// Every Reserve/Book button goes through this component so there is exactly
// one place that decides between the ticketing flow and the WhatsApp fallback.
export function ReserveButton({
  className = '',
  event,
  label,
  show,
  variant = 'primary',
}: ReserveButtonProps) {
  const href = bookingHref(event, show)
  const external = href ? isExternalBookingHref(href) : false

  if (!href) {
    const status = show ? formatStatus(show.status) : 'Unavailable'

    return (
      <span
        aria-disabled="true"
        className={`inline-flex items-center justify-center gap-2 rounded-full border border-[var(--border)] px-5 py-3 text-sm font-medium tracking-wide text-[var(--muted)] opacity-70 ${className}`}
      >
        {status}
      </span>
    )
  }

  return (
    <Button
      className={className}
      href={href}
      {...(external ? { rel: 'noreferrer', target: '_blank' } : {})}
      variant={variant}
    >
      {event.ticketing.active ? 'Reserve seats' : label}
    </Button>
  )
}
