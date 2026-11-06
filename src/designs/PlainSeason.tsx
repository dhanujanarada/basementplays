'use client'

import { PlainSectionHeading } from '@/designs/PlainSectionHeading'
import { Badge } from '@/components/ui/Badge'
import { Container } from '@/components/ui/Container'
import { ReserveButton } from '@/components/ticketing/ReserveButton'
import { SeasonBooking } from '@/components/ticketing/SeasonBooking'
import type { EventView } from '@/lib/event-view'
import {
  formatLongShowDate,
  formatPriceFrom,
  formatStatus,
  statusTone,
} from '@/lib/format'

type PlainSeasonProps = {
  event: EventView
}

export function PlainSeason({ event }: PlainSeasonProps) {
  return (
    <SeasonBooking
      event={event}
      renderCard={({ href, isNextShow, seasonEnded, show }) => (
        <aside
          aria-labelledby="booking-heading"
          className="scroll-mt-24 border border-[var(--border)] bg-[var(--surface)] p-6"
          id="booking"
        >
          <p className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
            Booking
          </p>
          <h3
            className="font-display mt-2 text-2xl text-[var(--text)]"
            id="booking-heading"
          >
            Reserve your seats
          </h3>

          {show ? (
            <>
              <dl className="mt-5 grid gap-3 border-t border-[var(--border)] pt-5">
                <div>
                  <dt className="text-[10px] uppercase tracking-[0.18em] text-[var(--accent)]">
                    {isNextShow ? 'Next show' : 'Selected show'}
                  </dt>
                  <dd className="mt-1 text-sm text-[var(--text)]">
                    {formatLongShowDate(show.startsAt)}
                  </dd>
                </div>
                {show.venue ? (
                  <div>
                    <dt className="text-[10px] uppercase tracking-[0.18em] text-[var(--accent)]">
                      Venue
                    </dt>
                    <dd className="mt-1 text-sm text-[var(--text)]">
                      {show.venue}
                    </dd>
                  </div>
                ) : null}
                {formatPriceFrom(show.tiers) ? (
                  <div>
                    <dt className="text-[10px] uppercase tracking-[0.18em] text-[var(--accent)]">
                      Tickets
                    </dt>
                    <dd className="mt-1 text-sm text-[var(--text)]">
                      {formatPriceFrom(show.tiers)}
                    </dd>
                  </div>
                ) : null}
              </dl>

              {seasonEnded ? (
                <p className="mt-4 text-xs leading-5 text-[var(--muted)]">
                  The season has ended. Booking opens as soon as the next date
                  is announced.
                </p>
              ) : null}

              {!href && !seasonEnded ? (
                <p className="mt-4 text-xs leading-5 text-[var(--muted)]">
                  This performance is {formatStatus(show.status).toLowerCase()}.
                </p>
              ) : null}

              <div className="mt-5">
                <ReserveButton
                  className="w-full"
                  event={event}
                  label="Reserve seats"
                  show={show}
                />
              </div>

              <p className="mt-4 text-xs leading-5 text-[var(--muted)]">
                Reservations are handled through the ticket link. Seats are
                confirmed directly by the company.
              </p>
            </>
          ) : (
            <p className="mt-4 text-sm leading-6 text-[var(--muted)]">
              Booking opens as soon as the next date is announced.
            </p>
          )}
        </aside>
      )}
    >
      {({ canSelect, card, isSelected, seasons, selectShow }) => (
        <section
          aria-labelledby="show-dates-heading"
          className="scroll-mt-24 border-b border-[var(--border)]"
          id="show-dates"
        >
          <Container className="py-20">
            <div className="grid gap-10 lg:grid-cols-[0.32fr_0.68fr]">
              <PlainSectionHeading
                eyebrow="Performances"
                headingId="show-dates-heading"
                title="Show dates"
              />

              <div className="space-y-10">
                {card}

                {seasons.length > 0 ? (
                  seasons.map((season) => (
                    <article
                      className="border-t border-[var(--border)] pt-8"
                      key={`${season.label}-${season.venue}`}
                    >
                      <div className="flex flex-wrap items-start justify-between gap-4">
                        <div>
                          <h3 className="font-display text-2xl text-[var(--text)]">
                            {season.label}
                          </h3>
                          {season.venue ? (
                            <p className="mt-1 text-sm text-[var(--muted)]">
                              {season.venue}
                            </p>
                          ) : null}
                        </div>
                        <p className="text-[10px] uppercase tracking-[0.18em] text-[var(--muted)]">
                          {season.shows.length}{' '}
                          {season.shows.length === 1 ? 'show' : 'shows'}
                        </p>
                      </div>

                      <ul className="mt-4 divide-y divide-[var(--border)]">
                        {season.shows.map((show) => {
                          const selectable = canSelect(show)
                          const selected = isSelected(show)
                          const price = formatPriceFrom(show.tiers)

                          return (
                            <li
                              className={`relative flex flex-col gap-4 py-5 sm:flex-row sm:items-center sm:justify-between ${
                                selected ? 'bg-[var(--surface-raised)]' : ''
                              } ${
                                show.status === 'past' ? 'opacity-40' : ''
                              }`}
                              key={show.id}
                            >
                              {selectable ? (
                                <button
                                  aria-label={`Select ${formatLongShowDate(show.startsAt)}`}
                                  aria-pressed={selected}
                                  className="absolute inset-0 z-10 rounded-sm transition hover:bg-[var(--accent)]/5 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--accent)]"
                                  onClick={() => selectShow(show.id)}
                                  type="button"
                                />
                              ) : null}

                              <div>
                                <p className="text-sm text-[var(--text)]">
                                  {formatLongShowDate(show.startsAt)}
                                </p>
                                {price ? (
                                  <p className="mt-1 text-xs text-[var(--muted)]">
                                    {price}
                                  </p>
                                ) : null}
                              </div>

                              <div className="flex flex-wrap items-center gap-3">
                                <Badge tone={statusTone(show.status)}>
                                  {formatStatus(show.status)}
                                </Badge>
                              </div>
                            </li>
                          )
                        })}
                      </ul>
                    </article>
                  ))
                ) : (
                  <p className="text-sm leading-7 text-[var(--muted)]">
                    No show dates have been announced yet. Check back soon.
                  </p>
                )}
              </div>
            </div>
          </Container>
        </section>
      )}
    </SeasonBooking>
  )
}
