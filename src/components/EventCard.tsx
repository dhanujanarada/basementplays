import Image from 'next/image'
import Link from 'next/link'

import { Badge } from '@/components/ui/Badge'
import type { Production } from '@/payload-types'
import { formatShowDate, formatStatus, statusTone } from '@/lib/format'
import { mediaImage } from '@/lib/media'
import type { ShowEntry } from '@/lib/types'

type EventCardProps = {
  nextShow: ShowEntry | null
  production: Production
}

export function EventCard({ nextShow, production }: EventCardProps) {
  const poster = mediaImage(production.poster, 'card')

  return (
    <Link
      className="group flex flex-col overflow-hidden rounded-sm border border-[var(--border)] bg-[var(--surface)] transition hover:border-[var(--accent)]/70 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--accent)]"
      href={`/events/${production.slug}`}
    >
      <div className="relative aspect-[2/3] overflow-hidden bg-[var(--surface-raised)]">
        {poster ? (
          <Image
            alt={poster.alt}
            className="object-cover transition duration-700 group-hover:scale-[1.03] motion-reduce:transition-none motion-reduce:group-hover:scale-100"
            fill
            sizes="(min-width: 1024px) 33vw, (min-width: 640px) 50vw, 100vw"
            src={poster.src}
          />
        ) : (
          <div className="site-grid flex h-full flex-col justify-between p-6">
            <span className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
              Basement Plays
            </span>
            <span className="font-display text-3xl leading-[0.95] text-[var(--text)]">
              {production.title}
            </span>
          </div>
        )}

        <div className="absolute left-4 top-4">
          {nextShow ? (
            <Badge tone={statusTone(nextShow.show.status)}>
              {formatStatus(nextShow.show.status)}
            </Badge>
          ) : (
            <Badge tone="muted">No upcoming shows</Badge>
          )}
        </div>
      </div>

      <div className="flex flex-1 flex-col gap-3 p-5">
        <h2 className="font-display text-2xl leading-tight text-[var(--text)] transition group-hover:text-[var(--accent)]">
          {production.title}
        </h2>
        <p className="mt-auto text-sm text-[var(--muted)]">
          {nextShow
            ? formatShowDate(nextShow.show.startsAt)
            : 'Check back for future dates'}
        </p>
      </div>
    </Link>
  )
}
