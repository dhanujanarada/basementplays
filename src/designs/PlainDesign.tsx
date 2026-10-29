import { RichText } from '@payloadcms/richtext-lexical/react'
import Image from 'next/image'
import Link from 'next/link'

import { Gallery } from '@/components/Gallery'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Container } from '@/components/ui/Container'
import {
  formatLongShowDate,
  formatStatus,
  statusTone,
} from '@/lib/format'
import { mediaImage, type MediaImage } from '@/lib/media'
import { getUpcomingShow, sortRunsNewestFirst, sortShowsAscending } from '@/lib/shows'
import type { ProductionDesignProps } from '@/lib/types'

export function PlainDesign({ production, runs }: ProductionDesignProps) {
  const poster = mediaImage(production.poster, 'full')
  const hero = mediaImage(production.heroImage, 'full')
  const nextShow = getUpcomingShow(runs)
  const orderedRuns = sortRunsNewestFirst(runs)

  const gallery = (production.gallery ?? [])
    .map((image) => mediaImage(image, 'full'))
    .filter((image): image is MediaImage => image !== null)

  const backstage = (production.backstage ?? []).flatMap((item) => {
    const image = mediaImage(item.image, 'card')

    return image ? [{ ...item, image }] : []
  })

  const credits = [
    { label: 'Presented by', value: production.presentedBy },
    { label: 'Written by', value: production.writtenBy },
    { label: 'Directed by', value: production.directedBy },
  ].filter((credit) => Boolean(credit.value))

  return (
    <article>
      <header className="relative isolate overflow-hidden border-b border-[var(--border)]">
        {hero ? (
          <div className="absolute inset-0 -z-20">
            <Image
              alt=""
              className="object-cover opacity-30"
              fill
              priority
              sizes="100vw"
              src={hero.src}
            />
            <div className="absolute inset-0 bg-gradient-to-b from-[var(--bg)]/50 via-[var(--bg)]/85 to-[var(--bg)]" />
          </div>
        ) : null}
        <div className="site-grid pointer-events-none absolute inset-0 -z-10 opacity-40" />

        <Container className="py-16 sm:py-20 lg:py-28">
          <Link
            className="text-xs font-medium uppercase tracking-[0.18em] text-[var(--muted)] transition hover:text-[var(--accent)]"
            href="/events"
          >
            &larr; All events
          </Link>

          <div className="mt-10 grid gap-12 lg:grid-cols-[minmax(0,1fr)_minmax(260px,360px)] lg:items-end">
            <div>
              <div className="flex flex-wrap items-center gap-3">
                {nextShow ? (
                  <Badge tone={statusTone(nextShow.show.status)}>
                    {formatStatus(nextShow.show.status)}
                  </Badge>
                ) : (
                  <Badge tone="muted">No upcoming shows</Badge>
                )}
                {production.presentedBy ? (
                  <span className="text-xs uppercase tracking-[0.18em] text-[var(--muted)]">
                    {production.presentedBy}
                  </span>
                ) : null}
              </div>

              <h1 className="font-display mt-7 max-w-3xl text-5xl leading-[0.92] tracking-tight text-[var(--text)] sm:text-7xl">
                {production.title}
              </h1>

              {production.tagline ? (
                <p className="mt-6 max-w-2xl text-lg leading-8 text-[var(--muted)]">
                  {production.tagline}
                </p>
              ) : null}

              {credits.length > 0 ? (
                <dl className="mt-9 grid gap-x-8 gap-y-4 sm:grid-cols-2">
                  {credits.map((credit) => (
                    <div key={credit.label}>
                      <dt className="text-[10px] font-medium uppercase tracking-[0.2em] text-[var(--accent)]">
                        {credit.label}
                      </dt>
                      <dd className="mt-1 text-sm text-[var(--text)]">
                        {credit.value}
                      </dd>
                    </div>
                  ))}
                </dl>
              ) : null}

              <div className="mt-10 flex flex-wrap gap-3">
                <Button href="#show-dates">See show dates</Button>
                <Button href="/events" variant="secondary">
                  Browse all events
                </Button>
              </div>
            </div>

            <div className="relative mx-auto aspect-[2/3] w-full max-w-[360px] overflow-hidden rounded-sm border border-[var(--border)] bg-[var(--surface-raised)] shadow-2xl shadow-black/30 lg:mx-0">
              {poster ? (
                <Image
                  alt={poster.alt}
                  className="object-cover"
                  fill
                  priority
                  sizes="(min-width: 1024px) 360px, 80vw"
                  src={poster.src}
                />
              ) : (
                <div className="site-grid flex h-full flex-col justify-between p-6">
                  <span className="text-[10px] uppercase tracking-[0.22em] text-[var(--accent)]">
                    Basement Plays
                  </span>
                  <span className="font-display text-4xl leading-none text-[var(--text)]">
                    {production.title}
                  </span>
                </div>
              )}
            </div>
          </div>
        </Container>
      </header>

      <section
        aria-labelledby="show-dates-heading"
        className="scroll-mt-24 border-b border-[var(--border)]"
        id="show-dates"
      >
        <Container className="py-20">
          <div className="grid gap-10 lg:grid-cols-[0.32fr_0.68fr]">
            <SectionHeading
              eyebrow="Performances"
              headingId="show-dates-heading"
              title="Show dates"
            />

            <div>
              {orderedRuns.length > 0 ? (
                <div className="space-y-10">
                  {orderedRuns.map((run) => {
                    const shows = sortShowsAscending(run.shows ?? [])

                    return (
                      <article
                        className="border-t border-[var(--border)] pt-8 first:border-t-0 first:pt-0"
                        key={run.id}
                      >
                        <div className="flex flex-wrap items-start justify-between gap-4">
                          <div>
                            <h3 className="font-display text-2xl text-[var(--text)]">
                              {run.label}
                            </h3>
                            {run.venue ? (
                              <p className="mt-1 text-sm text-[var(--muted)]">
                                {run.venue}
                              </p>
                            ) : null}
                          </div>
                          <p className="text-[10px] uppercase tracking-[0.18em] text-[var(--muted)]">
                            {shows.length} {shows.length === 1 ? 'show' : 'shows'}
                          </p>
                        </div>

                        <ul className="mt-4 divide-y divide-[var(--border)]">
                          {shows.map((show, index) => {
                            const canBook =
                              Boolean(show.ticketUrl) &&
                              show.status !== 'sold-out' &&
                              show.status !== 'cancelled'

                            return (
                              <li
                                className={`flex flex-col gap-4 py-5 sm:flex-row sm:items-center sm:justify-between ${
                                  show.status === 'cancelled'
                                    ? 'opacity-60'
                                    : ''
                                }`}
                                key={show.id ?? `${show.startsAt}-${index}`}
                              >
                                <div>
                                  <p className="text-sm text-[var(--text)]">
                                    {formatLongShowDate(show.startsAt)}
                                  </p>
                                  <p className="mt-1 text-xs text-[var(--muted)]">
                                    {show.price}
                                    {show.seatsNote
                                      ? ` · ${show.seatsNote}`
                                      : ''}
                                  </p>
                                </div>

                                <div className="flex flex-wrap items-center gap-3">
                                  <Badge tone={statusTone(show.status)}>
                                    {formatStatus(show.status)}
                                  </Badge>
                                  {canBook ? (
                                    <Button
                                      className="px-4 py-2 text-xs"
                                      href={show.ticketUrl ?? '#'}
                                      rel="noreferrer"
                                      target="_blank"
                                      variant="secondary"
                                    >
                                      Get tickets
                                    </Button>
                                  ) : null}
                                </div>
                              </li>
                            )
                          })}
                        </ul>
                      </article>
                    )
                  })}
                </div>
              ) : (
                <p className="text-sm leading-7 text-[var(--muted)]">
                  No show dates have been announced yet. Check back soon.
                </p>
              )}
            </div>
          </div>
        </Container>
      </section>

      {production.partners && production.partners.length > 0 ? (
        <section className="border-b border-[var(--border)] bg-[var(--surface)]/40">
          <Container className="flex flex-col gap-6 py-10 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--muted)]">
              With support from
            </p>
            <ul className="flex flex-wrap items-center gap-x-8 gap-y-4">
              {production.partners.map((partner, index) => {
                const logo = mediaImage(partner.logo, 'thumbnail')

                return (
                  <li
                    className="flex items-center gap-3"
                    key={partner.id ?? `${partner.name}-${index}`}
                  >
                    {logo ? (
                      <Image
                        alt={logo.alt}
                        className="max-h-10 w-auto object-contain"
                        height={logo.height}
                        src={logo.src}
                        width={logo.width}
                      />
                    ) : null}
                    <div>
                      <p className="text-sm text-[var(--text)]">
                        {partner.name}
                      </p>
                      {partner.role ? (
                        <p className="text-[10px] uppercase tracking-[0.16em] text-[var(--muted)]">
                          {partner.role}
                        </p>
                      ) : null}
                    </div>
                  </li>
                )
              })}
            </ul>
          </Container>
        </section>
      ) : null}

      {production.synopsis ? (
        <section
          aria-labelledby="synopsis-heading"
          className="border-b border-[var(--border)]"
        >
          <Container className="py-20">
            <div className="grid gap-10 lg:grid-cols-[0.32fr_0.68fr]">
              <SectionHeading
                eyebrow="The play"
                headingId="synopsis-heading"
                title="Synopsis"
              />
              <RichText
                className="max-w-2xl text-base leading-8 text-[var(--muted)] [&_a]:text-[var(--accent)] [&_a]:underline [&_blockquote]:border-l-2 [&_blockquote]:border-[var(--accent)] [&_blockquote]:pl-5 [&_h2]:font-display [&_h2]:text-3xl [&_h2]:text-[var(--text)] [&_h3]:font-display [&_h3]:text-2xl [&_h3]:text-[var(--text)] [&_li]:mb-2 [&_ol]:mb-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:mb-5 [&_p:last-child]:mb-0 [&_ul]:mb-5 [&_ul]:list-disc [&_ul]:pl-5"
                data={production.synopsis}
              />
            </div>
          </Container>
        </section>
      ) : null}

      {production.cast?.length || production.crew?.length ? (
        <section
          aria-labelledby="people-heading"
          className="border-b border-[var(--border)] bg-[var(--surface)]/35"
        >
          <Container className="py-20">
            <SectionHeading
              eyebrow="The company"
              headingId="people-heading"
              title="Cast & crew"
            />

            <div className="mt-10 grid gap-12 md:grid-cols-2">
              {production.cast && production.cast.length > 0 ? (
                <div>
                  <h3 className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
                    Cast
                  </h3>
                  <ul className="mt-5 divide-y divide-[var(--border)]">
                    {production.cast.map((member, index) => {
                      const photo = mediaImage(member.photo, 'thumbnail')

                      return (
                        <li
                          className="flex items-center gap-4 py-4 first:pt-0"
                          key={member.id ?? `${member.name}-${index}`}
                        >
                          {photo ? (
                            <Image
                              alt={photo.alt}
                              className="h-12 w-12 rounded-full object-cover"
                              height={48}
                              src={photo.src}
                              width={48}
                            />
                          ) : null}
                          <div>
                            <p className="text-sm text-[var(--text)]">
                              {member.name}
                            </p>
                            {member.role ? (
                              <p className="mt-1 text-xs text-[var(--muted)]">
                                {member.role}
                              </p>
                            ) : null}
                          </div>
                        </li>
                      )
                    })}
                  </ul>
                </div>
              ) : null}

              {production.crew && production.crew.length > 0 ? (
                <div>
                  <h3 className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
                    Crew
                  </h3>
                  <ul className="mt-5 divide-y divide-[var(--border)]">
                    {production.crew.map((member, index) => (
                      <li
                        className="py-4 first:pt-0"
                        key={member.id ?? `${member.name}-${index}`}
                      >
                        <p className="text-sm text-[var(--text)]">
                          {member.name}
                        </p>
                        {member.role ? (
                          <p className="mt-1 text-xs text-[var(--muted)]">
                            {member.role}
                          </p>
                        ) : null}
                      </li>
                    ))}
                  </ul>
                </div>
              ) : null}
            </div>
          </Container>
        </section>
      ) : null}

      {gallery.length > 0 ? (
        <section
          aria-labelledby="gallery-heading"
          className="border-b border-[var(--border)]"
        >
          <Container className="py-20">
            <SectionHeading
              eyebrow="Production stills"
              headingId="gallery-heading"
              title="Gallery"
            />
            <div className="mt-10">
              <Gallery images={gallery} />
            </div>
          </Container>
        </section>
      ) : null}

      {backstage.length > 0 ? (
        <section
          aria-labelledby="backstage-heading"
          className="border-b border-[var(--border)] bg-[#17130f]"
        >
          <Container className="py-20">
            <SectionHeading
              eyebrow="Behind the scenes"
              headingId="backstage-heading"
              title="Backstage"
            />
            <div className="mt-10 grid gap-6 md:grid-cols-2">
              {backstage.map((item, index) => (
                <figure
                  className="overflow-hidden rounded-sm border border-[#3a2f24] bg-[#211a14]"
                  key={item.id ?? `${item.image.src}-${index}`}
                >
                  <div className="relative aspect-[4/3]">
                    <Image
                      alt={item.image.alt}
                      className="object-cover"
                      fill
                      sizes="(min-width: 768px) 50vw, 100vw"
                      src={item.image.src}
                    />
                  </div>
                  {item.caption ? (
                    <figcaption className="px-5 py-4 text-sm leading-6 text-[#c9bba9]">
                      {item.caption}
                    </figcaption>
                  ) : null}
                </figure>
              ))}
            </div>
          </Container>
        </section>
      ) : null}

      <section>
        <Container className="flex flex-col items-start gap-5 py-16 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-[10px] uppercase tracking-[0.2em] text-[var(--accent)]">
              More from Basement Plays
            </p>
            <p className="font-display mt-2 text-3xl text-[var(--text)]">
              See what else is on.
            </p>
          </div>
          <Button href="/events" variant="secondary">
            All events
          </Button>
        </Container>
      </section>
    </article>
  )
}

function SectionHeading({
  eyebrow,
  headingId,
  title,
}: {
  eyebrow: string
  headingId: string
  title: string
}) {
  return (
    <div>
      <p className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
        {eyebrow}
      </p>
      <h2
        className="font-display mt-3 text-4xl text-[var(--text)] sm:text-5xl"
        id={headingId}
      >
        {title}
      </h2>
    </div>
  )
}
