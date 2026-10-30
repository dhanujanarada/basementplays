import { convertLexicalToPlaintext } from '@payloadcms/richtext-lexical/plaintext'
import Image from 'next/image'

import { EventCard } from '@/components/EventCard'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Container } from '@/components/ui/Container'
import type { Production } from '@/payload-types'
import { formatShowDate, formatStatus, statusTone } from '@/lib/format'
import { mediaImage } from '@/lib/media'
import {
  getAllProductions,
  getRunsForProduction,
  getSiteSettings,
} from '@/lib/payload'
import { getUpcomingShow } from '@/lib/shows'
import type { ShowEntry } from '@/lib/types'

export const dynamic = 'force-dynamic'

type NextProduction = {
  nextShow: ShowEntry
  production: Production
}

export default async function HomePage() {
  const [productions, settings] = await Promise.all([
    getAllProductions(),
    getSiteSettings(),
  ])
  const upcoming = (
    await Promise.all(
      productions.map(async (production) => {
        const runs = await getRunsForProduction(production.id)
        const nextShow = getUpcomingShow(runs)

        return nextShow ? { nextShow, production } : null
      }),
    )
  )
    .filter((item): item is NextProduction => item !== null)
    .sort(
      (a, b) =>
        new Date(a.nextShow.show.startsAt).getTime() -
        new Date(b.nextShow.show.startsAt).getTime(),
    )

  const next = upcoming[0]
  const poster = next ? mediaImage(next.production.poster, 'card') : null
  const groupName = settings.groupName || 'Basement Plays'
  const tagline =
    settings.tagline ??
    'Independent theatre making intimate, ambitious work in the basement.'
  const aboutTeaser = settings.aboutText
    ? teaser(convertLexicalToPlaintext({ data: settings.aboutText }))
    : tagline

  return (
    <>
      <section className="relative overflow-hidden border-b border-[var(--border)]">
        <div className="site-grid pointer-events-none absolute inset-0 opacity-40" />
        <Container className="relative grid min-h-[70vh] items-center gap-12 py-20 lg:grid-cols-[minmax(0,1fr)_minmax(280px,400px)] lg:py-24">
          <div className="max-w-3xl">
            <Badge tone="accent">{next ? 'Next up' : groupName}</Badge>

            {next ? (
              <>
                <p className="mt-8 text-xs font-medium uppercase tracking-[0.22em] text-[var(--muted)]">
                  {next.nextShow.run.venue ??
                    settings.locationText ??
                    groupName}
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
                  {groupName}
                </h1>
                <p className="mt-8 max-w-xl text-base leading-7 text-[var(--muted)] sm:text-lg">
                  {tagline}
                </p>
                <div className="mt-10 flex flex-wrap items-center gap-4">
                  <Button href="/events">See what&apos;s on</Button>
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
                    {groupName}
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

      <section className="border-b border-[var(--border)]">
        <Container className="grid gap-8 py-16 sm:py-20 lg:grid-cols-[0.35fr_0.65fr]">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
              About
            </p>
            <h2 className="font-display mt-3 text-4xl tracking-tight text-[var(--text)] sm:text-5xl">
              A little about {groupName}
            </h2>
          </div>
          <div>
            <p className="max-w-2xl text-base leading-8 text-[var(--muted)]">
              {aboutTeaser}
            </p>
            <Button className="mt-6" href="/about" variant="secondary">
              Read more
            </Button>
          </div>
        </Container>
      </section>

      <section>
        <Container className="py-16 sm:py-20">
          <div className="flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <p className="text-xs font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
                Upcoming shows
              </p>
              <h2 className="font-display mt-3 text-4xl tracking-tight text-[var(--text)] sm:text-5xl">
                On stage next
              </h2>
            </div>
            <Button href="/events" variant="secondary">
              All events
            </Button>
          </div>

          {upcoming.length > 0 ? (
            <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {upcoming.map(({ nextShow, production }) => (
                <EventCard
                  groupName={groupName}
                  key={production.id}
                  nextShow={nextShow}
                  production={production}
                />
              ))}
            </div>
          ) : (
            <p className="mt-10 border-t border-[var(--border)] pt-8 text-sm text-[var(--muted)]">
              No upcoming shows are listed yet. Check back soon.
            </p>
          )}
        </Container>
      </section>
    </>
  )
}

function teaser(text: string, limit = 320): string {
  const clean = text.replace(/\s+/g, ' ').trim()

  if (clean.length <= limit) {
    return clean
  }

  const cut = clean.slice(0, limit)
  const lastSpace = cut.lastIndexOf(' ')

  return `${lastSpace > 0 ? cut.slice(0, lastSpace) : cut}…`
}
