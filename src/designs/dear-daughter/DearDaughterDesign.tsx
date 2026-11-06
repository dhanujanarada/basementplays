import { RichText } from '@payloadcms/richtext-lexical/react'
import { Caveat_Brush, Quicksand } from 'next/font/google'
import Image from 'next/image'
import Link from 'next/link'

import { Gallery, type GalleryImage } from '@/components/Gallery'
import { SiteFooter } from '@/components/SiteFooter'
import { Button } from '@/components/ui/Button'
import { Container } from '@/components/ui/Container'
import { formatLongShowDate, formatShowDate } from '@/lib/format'
import { getNextShow, showSeasonLabel } from '@/lib/event-view'
import { mediaImage } from '@/lib/media'
import type { EventDesignProps, EventShellProps } from '@/lib/types'

import { DearDaughterSeason } from './DearDaughterSeason'
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

function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .map((part) => part[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
}

export function DearDaughterPage({ event }: EventDesignProps) {
  const poster = mediaImage(event.poster, 'full')
  const hero = mediaImage(event.heroImage, 'full') ?? poster
  const nextShow = getNextShow(event)
  const venue = nextShow?.venue ?? event.shows[0]?.venue ?? event.venue

  const gallery: GalleryImage[] = event.gallery.flatMap((item) => {
    const image = mediaImage(item.image, 'full')

    return image ? [{ ...image, caption: item.caption ?? undefined }] : []
  })

  const cast = event.cast
  const crew = event.crew
  const partners = event.partners

  const credits = [
    { label: 'Presented by', value: event.presentedBy },
    { label: 'Written by', value: event.writtenBy },
    { label: 'Directed by', value: event.directedBy },
  ].filter((credit) => Boolean(credit.value))

  return (
    <>
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
              {event.presentedBy ?? 'A Basement production'}
            </p>

            <h1 className="dd-title">{event.title}</h1>

            {event.tagline ? (
              <p className="dd-tagline">{event.tagline}</p>
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
              <Button className="dd-cta" href="#booking">
                Get tickets
              </Button>
              <a className="dd-text-link" href="#show-dates">
                Show dates
              </a>
            </div>

            {nextShow ? (
              <p className="dd-hero__next">
                <span aria-hidden />
                {formatLongShowDate(nextShow.startsAt)}
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
                  <span>{event.presentedBy ?? 'Basement Plays'}</span>
                  <strong>{event.title}</strong>
                </div>
              )}
            </div>
            <p className="dd-poster__caption">
              {venue ?? 'Basement Plays'}
              {nextShow ? ` · ${showSeasonLabel(nextShow.startsAt)}` : ''}
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

      <DearDaughterSeason event={event} />

      <div aria-hidden className="dd-wood-divider" />

      {event.synopsis ? (
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
              <RichText className="dd-richtext" data={event.synopsis} />
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

      {nextShow ? (
        <div className="dd-sticky-bar">
          <div className="dd-sticky-bar__copy">
            <span className="dd-sticky-bar__label">Next show</span>
            <span className="dd-sticky-bar__date">
              {formatShowDate(nextShow.startsAt)}
            </span>
          </div>
          <Button className="dd-cta dd-cta--small" href="#booking">
            Get tickets
          </Button>
        </div>
      ) : null}
    </>
  )
}

export function DearDaughterShell({
  children,
  event,
}: EventShellProps) {
  const nextShow = getNextShow(event)
  const hasBar = Boolean(nextShow)

  return (
    <article
      className={`dd-page ${displayFont.variable} ${bodyFont.variable}${
        hasBar ? ' dd-page--has-bar' : ''
      }`}
    >
      {children}
      <SiteFooter />
    </article>
  )
}
