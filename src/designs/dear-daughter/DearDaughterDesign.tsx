import { RichText } from '@payloadcms/richtext-lexical/react'
import { Caveat_Brush, Quicksand } from 'next/font/google'
import Image from 'next/image'
import Link from 'next/link'

import { Gallery, type GalleryImage } from '@/components/Gallery'
import { SiteFooter } from '@/components/SiteFooter'
import { Button } from '@/components/ui/Button'
import { Container } from '@/components/ui/Container'
import {
  formatLongShowDate,
  formatShowDate,
  formatShowTime,
  formatStatus,
} from '@/lib/format'
import { mediaImage } from '@/lib/media'
import {
  getUpcomingShow,
  sortRunsNewestFirst,
  sortShowsAscending,
} from '@/lib/shows'
import type { ProductionDesignProps } from '@/lib/types'

import './dear-daughter.css'

const displayFont = Caveat_Brush({
  subsets: ['latin'],
  variable: '--font-dd-display',
  weight: '400',
  display: 'swap',
})

const bodyFont = Quicksand({
  subsets: ['latin'],
  variable: '--font-dd-body',
  display: 'swap',
})

const siteTimeZone = 'Asia/Colombo'

const weekdayFormatter = new Intl.DateTimeFormat('en-GB', {
  timeZone: siteTimeZone,
  weekday: 'short',
})

const dayFormatter = new Intl.DateTimeFormat('en-GB', {
  day: '2-digit',
  timeZone: siteTimeZone,
})

const monthFormatter = new Intl.DateTimeFormat('en-GB', {
  month: 'short',
  timeZone: siteTimeZone,
})

type DateParts = {
  day: string
  full: string
  month: string
  weekday: string
}

function dateParts(value: string): DateParts {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return {
      day: '—',
      full: 'Date to be announced',
      month: 'TBA',
      weekday: '',
    }
  }

  return {
    day: dayFormatter.format(date),
    full: formatLongShowDate(value),
    month: monthFormatter.format(date),
    weekday: weekdayFormatter.format(date),
  }
}

function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .map((part) => part[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
}

export function DearDaughterDesign({
  production,
  runs,
}: ProductionDesignProps) {
  const poster = mediaImage(production.poster, 'full')
  const hero = mediaImage(production.heroImage, 'full') ?? poster
  const nextShow = getUpcomingShow(runs)
  const orderedRuns = sortRunsNewestFirst(runs)
  const venue = nextShow?.run.venue ?? orderedRuns[0]?.venue

  const gallery: GalleryImage[] = (production.gallery ?? []).flatMap((item) => {
    const image = mediaImage(item.image, 'full')

    return image ? [{ ...image, caption: item.caption ?? undefined }] : []
  })

  const cast = production.cast ?? []
  const crew = production.crew ?? []
  const partners = production.partners ?? []

  const credits = [
    { label: 'Presented by', value: production.presentedBy },
    { label: 'Written by', value: production.writtenBy },
    { label: 'Directed by', value: production.directedBy },
  ].filter((credit) => Boolean(credit.value))

  const canBookNext =
    Boolean(nextShow?.show.ticketUrl) &&
    nextShow?.show.status !== 'sold-out' &&
    nextShow?.show.status !== 'cancelled'
  const nextTicketUrl = canBookNext ? (nextShow?.show.ticketUrl ?? null) : null
  const heroCtaHref = nextTicketUrl ?? '#show-dates'
  const heroCtaLabel = nextTicketUrl ? 'Reserve seats' : 'See show dates'

  return (
    <article
      className={`dd-page ${displayFont.variable} ${bodyFont.variable}${
        nextTicketUrl ? ' dd-page--has-bar' : ''
      }`}
    >
      <header className="dd-hero">
        {hero ? (
          <Image
            alt=""
            className="dd-hero__photo"
            fill
            priority
            sizes="100vw"
            src={hero.src}
          />
        ) : null}
        <div aria-hidden className="dd-wallpaper" />
        <div aria-hidden className="dd-hero__veil" />
        <span aria-hidden className="dd-hero__seal" />

        <Container className="dd-hero__inner">
          <div className="dd-hero__copy">
            <Link className="dd-back-link" href="/events">
              &larr; All events
            </Link>

            <p className="dd-eyebrow">
              {production.presentedBy ?? 'A Basement production'}
            </p>

            <h1 className="dd-title">{production.title}</h1>

            {production.tagline ? (
              <p className="dd-tagline">{production.tagline}</p>
            ) : null}

            {credits.length > 0 ? (
              <dl className="dd-credits">
                {credits.map((credit) => (
                  <div key={credit.label}>
                    <dt>{credit.label}</dt>
                    <dd>{credit.value}</dd>
                  </div>
                ))}
              </dl>
            ) : null}

            <div className="dd-hero__actions">
              <Button
                className="dd-cta"
                href={heroCtaHref}
                {...(nextTicketUrl
                  ? { rel: 'noreferrer', target: '_blank' }
                  : {})}
              >
                {heroCtaLabel}
              </Button>
              <a className="dd-text-link" href="#show-dates">
                Show dates
              </a>
            </div>

            {nextShow ? (
              <p className="dd-hero__next">
                <span aria-hidden />
                {formatLongShowDate(nextShow.show.startsAt)}
                {venue ? ` · ${venue}` : ''}
              </p>
            ) : null}
          </div>

          <div className="dd-poster-wrap">
            <div className="dd-poster">
              {poster ? (
                <Image
                  alt={poster.alt}
                  fill
                  priority
                  sizes="(min-width: 1024px) 360px, 80vw"
                  src={poster.src}
                />
              ) : (
                <div className="dd-poster__fallback">
                  <span>{production.presentedBy ?? 'Basement Plays'}</span>
                  <strong>{production.title}</strong>
                </div>
              )}
            </div>
            <p className="dd-poster__caption">
              {venue ?? 'Basement Plays'}
              {nextShow ? ` · ${nextShow.run.label}` : ''}
            </p>
          </div>
        </Container>

        {partners.length > 0 ? (
          <Container className="dd-partner-strip">
            <p className="dd-partner-strip__label">With support from</p>
            <ul className="dd-partner-strip__list">
              {partners.map((partner, index) => {
                const logo = mediaImage(partner.logo, 'thumbnail')

                return (
                  <li
                    className="dd-partner-strip__item"
                    key={partner.id ?? `${partner.name}-${index}`}
                  >
                    {logo ? (
                      <Image
                        alt={logo.alt}
                        className="dd-partner-strip__logo"
                        height={logo.height}
                        src={logo.src}
                        width={logo.width}
                      />
                    ) : (
                      <span className="dd-partner-strip__name">
                        {partner.name}
                      </span>
                    )}
                    {partner.role ? (
                      <span className="dd-partner-strip__role">
                        {partner.role}
                      </span>
                    ) : null}
                  </li>
                )
              })}
            </ul>
          </Container>
        ) : null}
      </header>

      <section
        aria-labelledby="dd-show-dates-heading"
        className="dd-season"
        id="show-dates"
      >
        <div aria-hidden className="dd-torn" />

        <Container className="dd-season__inner">
          <div className="dd-season__details">
            <div className="dd-section-intro">
              <p className="dd-eyebrow">Performances</p>
              <h2 className="dd-section-title" id="dd-show-dates-heading">
                Show dates
              </h2>
              <p className="dd-section-note">
                {venue
                  ? `Join us at ${venue}.`
                  : 'Dates and venue will be announced soon.'}
              </p>
            </div>

            {orderedRuns.length > 0 ? (
              <div className="dd-runs">
                {orderedRuns.map((run) => {
                  const shows = sortShowsAscending(run.shows ?? [])

                  return (
                    <div className="dd-run" key={run.id}>
                      <div className="dd-run__head">
                        <h3>{run.label}</h3>
                        {run.venue ? <p>{run.venue}</p> : null}
                      </div>

                      {shows.length > 0 ? (
                        <ul className="dd-shows">
                          {shows.map((show, index) => {
                            const parts = dateParts(show.startsAt)
                            const canBook =
                              Boolean(show.ticketUrl) &&
                              show.status !== 'sold-out' &&
                              show.status !== 'cancelled'
                            const meta = [show.price, show.seatsNote]
                              .filter(Boolean)
                              .join(' · ')

                            return (
                              <li
                                aria-label={parts.full}
                                className={`dd-show dd-show--${show.status}`}
                                key={show.id ?? `${show.startsAt}-${index}`}
                              >
                                <div className="dd-show__date">
                                  <span className="dd-show__weekday">
                                    {parts.weekday}
                                  </span>
                                  <span className="dd-show__day">
                                    {parts.day}
                                  </span>
                                  <span className="dd-show__month">
                                    {parts.month}
                                  </span>
                                </div>

                                <div className="dd-show__details">
                                  <p className="dd-show__time">
                                    {formatShowTime(show.startsAt)}
                                  </p>
                                  {meta ? (
                                    <p className="dd-show__price">{meta}</p>
                                  ) : null}
                                </div>

                                <div className="dd-show__action">
                                  <span
                                    className={`dd-status dd-status--${show.status}`}
                                  >
                                    {formatStatus(show.status)}
                                  </span>
                                  {canBook ? (
                                    <Button
                                      className="dd-ticket-button"
                                      href={show.ticketUrl ?? '#'}
                                      rel="noreferrer"
                                      target="_blank"
                                      variant="secondary"
                                    >
                                      Reserve
                                    </Button>
                                  ) : null}
                                </div>
                              </li>
                            )
                          })}
                        </ul>
                      ) : (
                        <p className="dd-empty">
                          No show dates have been announced for this run yet.
                        </p>
                      )}
                    </div>
                  )
                })}
              </div>
            ) : (
              <p className="dd-empty">
                No show dates have been announced yet. Check back soon.
              </p>
            )}
          </div>

          <aside aria-labelledby="dd-booking-heading" className="dd-booking">
            <p className="dd-eyebrow">Booking</p>
            <h3 className="dd-booking__title" id="dd-booking-heading">
              Reserve your seats
            </h3>

            {nextShow ? (
              <>
                <dl className="dd-booking__meta">
                  <div>
                    <dt>Next show</dt>
                    <dd>{formatLongShowDate(nextShow.show.startsAt)}</dd>
                  </div>
                  {venue ? (
                    <div>
                      <dt>Venue</dt>
                      <dd>{venue}</dd>
                    </div>
                  ) : null}
                  {nextShow.show.price ? (
                    <div>
                      <dt>Tickets</dt>
                      <dd>{nextShow.show.price}</dd>
                    </div>
                  ) : null}
                  {nextShow.show.seatsNote ? (
                    <div>
                      <dt>Availability</dt>
                      <dd>{nextShow.show.seatsNote}</dd>
                    </div>
                  ) : null}
                </dl>

                {nextTicketUrl ? (
                  <Button
                    className="dd-cta dd-booking__button"
                    href={nextTicketUrl}
                    rel="noreferrer"
                    target="_blank"
                  >
                    Reserve seats
                  </Button>
                ) : (
                  <p className="dd-booking__note">
                    This performance is{' '}
                    {formatStatus(nextShow.show.status).toLowerCase()}.
                  </p>
                )}

                <p className="dd-booking__note">
                  Reservations are handled through the ticket link. Seats are
                  confirmed directly by the company.
                </p>
              </>
            ) : (
              <p className="dd-empty">
                Booking opens as soon as the next date is announced.
              </p>
            )}
          </aside>
        </Container>
      </section>

      <div aria-hidden className="dd-wood-divider" />

      {production.synopsis ? (
        <section
          aria-labelledby="dd-synopsis-heading"
          className="dd-synopsis"
        >
          <Container className="dd-synopsis__grid">
            <div className="dd-synopsis__aside">
              <p className="dd-eyebrow">The play</p>
              <h2 className="dd-section-title" id="dd-synopsis-heading">
                Synopsis
              </h2>
              <span aria-hidden className="dd-sprig" />
            </div>
            <div className="dd-synopsis__body">
              <RichText className="dd-richtext" data={production.synopsis} />
            </div>
          </Container>
        </section>
      ) : null}

      {cast.length > 0 || crew.length > 0 ? (
        <section aria-labelledby="dd-company-heading" className="dd-company">
          <Container>
            <div className="dd-section-intro dd-section-intro--center">
              <p className="dd-eyebrow">The company</p>
              <h2 className="dd-section-title" id="dd-company-heading">
                Cast &amp; crew
              </h2>
            </div>

            <div className="dd-company__grid">
              {cast.length > 0 ? (
                <div className="dd-company__column">
                  <h3>Cast</h3>
                  <ul className="dd-company__list dd-company__list--cast">
                    {cast.map((member, index) => {
                      const photo = mediaImage(member.photo, 'thumbnail')

                      return (
                        <li
                          className="dd-person"
                          key={member.id ?? `${member.name}-${index}`}
                        >
                          <span className="dd-person__portrait">
                            {photo ? (
                              <Image
                                alt={photo.alt}
                                height={48}
                                src={photo.src}
                                width={48}
                              />
                            ) : (
                              initials(member.name)
                            )}
                          </span>
                          <div>
                            <p className="dd-person__name">{member.name}</p>
                            {member.role ? (
                              <p className="dd-person__role">{member.role}</p>
                            ) : null}
                          </div>
                        </li>
                      )
                    })}
                  </ul>
                </div>
              ) : null}

              {crew.length > 0 ? (
                <div className="dd-company__column">
                  <h3>Crew</h3>
                  <ul className="dd-company__list">
                    {crew.map((member, index) => (
                      <li
                        className="dd-person"
                        key={member.id ?? `${member.name}-${index}`}
                      >
                        <div>
                          <p className="dd-person__name">{member.name}</p>
                          {member.role ? (
                            <p className="dd-person__role">{member.role}</p>
                          ) : null}
                        </div>
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
        <section aria-labelledby="dd-gallery-heading" className="dd-gallery">
          <Container>
            <div className="dd-section-intro dd-gallery__head">
              <p className="dd-eyebrow">Production stills</p>
              <h2 className="dd-section-title" id="dd-gallery-heading">
                Gallery
              </h2>
              <p className="dd-section-note">
                Click any image to view it larger.
              </p>
            </div>
            <Gallery images={gallery} />
          </Container>
        </section>
      ) : null}

      <section className="dd-outro">
        <Container className="dd-outro__inner">
          <div>
            <p className="dd-eyebrow">More from Basement Plays</p>
            <p className="dd-outro__title">See what else is on.</p>
          </div>
          <Button className="dd-cta dd-cta--light" href="/events">
            All events
          </Button>
        </Container>
      </section>

      {nextTicketUrl && nextShow ? (
        <div className="dd-sticky-bar">
          <div className="dd-sticky-bar__copy">
            <span className="dd-sticky-bar__label">Next show</span>
            <span className="dd-sticky-bar__date">
              {formatShowDate(nextShow.show.startsAt)}
            </span>
          </div>
          <Button
            className="dd-cta dd-cta--small"
            href={nextTicketUrl}
            rel="noreferrer"
            target="_blank"
          >
            Reserve
          </Button>
        </div>
      ) : null}

      <SiteFooter />
    </article>
  )
}
