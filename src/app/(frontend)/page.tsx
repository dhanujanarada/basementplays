import Image from 'next/image'

import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Container } from '@/components/ui/Container'
import type { Production } from '@/payload-types'
import { formatShowDate, formatStatus, statusTone } from '@/lib/format'
import { mediaImage } from '@/lib/media'
import { getAllProductions, getRunsForProduction } from '@/lib/payload'
import { getUpcomingShow } from '@/lib/shows'
import type { ShowEntry } from '@/lib/types'

export const dynamic = 'force-dynamic'

type NextProduction = {
  nextShow: ShowEntry
  production: Production
}

export default async function HomePage() {
  const productions = await getAllProductions()
  const upcoming = await Promise.all(
    productions.map(async (production) => {
      const runs = await getRunsForProduction(production.id)
      const nextShow = getUpcomingShow(runs)

      return nextShow ? { nextShow, production } : null
    }),
  )

  const next = upcoming
    .filter((item): item is NextProduction => item !== null)
    .sort(
      (a, b) =>
        new Date(a.nextShow.show.startsAt).getTime() -
        new Date(b.nextShow.show.startsAt).getTime(),
    )[0]

  const poster = next ? mediaImage(next.production.poster, 'card') : null

  return (
    <section className="relative overflow-hidden border-b border-[var(--border)]">
      <div className="site-grid pointer-events-none absolute inset-0 opacity-40" />
      <Container className="relative grid min-h-[70vh] items-center gap-12 py-20 lg:grid-cols-[minmax(0,1fr)_minmax(280px,400px)] lg:py-24">
        <div className="max-w-3xl">
          <Badge tone="accent">{next ? 'Next up' : 'Basement Plays'}</Badge>

          {next ? (
            <>
              <p className="mt-8 text-xs font-medium uppercase tracking-[0.22em] text-[var(--muted)]">
                {next.nextShow.run.venue ?? 'Basement Plays'}
              </p>
              <h1 className="font-display mt-4 text-5xl leading-[0.95] tracking-tight text-[var(--text)] sm:text-7xl lg:text-8xl">
                {next.production.title}
              </h1>
              {next.production.tagline ? (
                <p className="mt-7 max-w-xl text-base leading-7 text-[var(--muted)] sm:text-lg">
                  {next.production.tagline}
                </p>
              ) : null}

              <div className="mt-8 flex flex-wrap items-center gap-4">
                <Badge tone={statusTone(next.nextShow.show.status)}>
                  {formatStatus(next.nextShow.show.status)}
                </Badge>
                <p className="text-sm text-[var(--muted)]">
                  {formatShowDate(next.nextShow.show.startsAt)}
                </p>
              </div>

              <div className="mt-10 flex flex-wrap items-center gap-4">
                <Button href={`/events/${next.production.slug}`}>
                  View production
                </Button>
                <Button href="/events" variant="secondary">
                  All events
                </Button>
              </div>
            </>
          ) : (
            <>
              <h1 className="font-display mt-8 text-5xl leading-[0.95] tracking-tight text-[var(--text)] sm:text-7xl lg:text-8xl">
                Small room.
                <br />
                Big stories.
              </h1>
              <p className="mt-8 max-w-xl text-base leading-7 text-[var(--muted)] sm:text-lg">
                An independent theatre group making intimate, ambitious work in
                the basement.
              </p>
              <div className="mt-10 flex flex-wrap items-center gap-4">
                <Button href="/events">See what&apos;s on</Button>
                <Button href="/admin" variant="secondary">
                  Open admin
                </Button>
              </div>
            </>
          )}
        </div>

        {next ? (
          <div className="relative mx-auto aspect-[2/3] w-full max-w-[400px] overflow-hidden rounded-sm border border-[var(--border)] bg-[var(--surface-raised)] shadow-2xl shadow-black/30 lg:mx-0 lg:ml-auto">
            {poster ? (
              <Image
                alt={poster.alt}
                className="object-cover"
                fill
                priority
                sizes="(min-width: 1024px) 400px, 80vw"
                src={poster.src}
              />
            ) : (
              <div className="site-grid flex h-full flex-col justify-between p-6">
                <span className="text-[10px] uppercase tracking-[0.22em] text-[var(--accent)]">
                  Basement Plays
                </span>
                <span className="font-display text-4xl leading-none text-[var(--text)]">
                  {next.production.title}
                </span>
              </div>
            )}
          </div>
        ) : null}
      </Container>
    </section>
  )
}
