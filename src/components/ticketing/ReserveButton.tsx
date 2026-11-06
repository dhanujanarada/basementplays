import type { Production } from '@/payload-types'

import { Button } from '@/components/ui/Button'
import { formatStatus } from '@/lib/format'
import { bookingHref, isTicketingEnabled } from '@/lib/ticketing/booking'
import type { Show } from '@/lib/types'

type ReserveButtonProps = {
  className?: string
  // The design's existing label, kept when ticketing is disabled.
  label: string
  production: Production
  show: Show | null | undefined
  variant?: 'primary' | 'secondary' | 'ghost'
}

// Every Reserve/Book button goes through this component so there is exactly
// one place that decides between the ticketing flow and the external ticket
// link.
export function ReserveButton({
  className = '',
  label,
  production,
  show,
  variant = 'primary',
}: ReserveButtonProps) {
  const href = bookingHref(production, show)
  const ticketing = isTicketingEnabled(show)

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
      {...(ticketing ? {} : { rel: 'noreferrer', target: '_blank' })}
      variant={variant}
    >
      {ticketing ? 'Reserve seats' : label}
    </Button>
  )
}
