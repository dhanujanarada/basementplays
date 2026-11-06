'use client'

import { ReserveButton } from '@/components/ticketing/ReserveButton'
import { SeasonBooking } from '@/components/ticketing/SeasonBooking'
import { Container } from '@/components/ui/Container'
import type { EventView } from '@/lib/event-view'
import {
  formatLongShowDate,
  formatPriceFrom,
  formatShowTime,
  formatStatus,
} from '@/lib/format'

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

function dateParts(value: string) {
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

type DearDaughterSeasonProps = {
  event: EventView
}

export function DearDaughterSeason({ event }: DearDaughterSeasonProps) {
  return (
    <SeasonBooking
      event={event}
      renderCard={({ href, isNextShow, seasonEnded, show }) => (
        <aside
          aria-labelledby="dd-booking-heading"
          className="dd-booking"
          id="booking"
        >
          <p className="dd-eyebrow">Booking</p>
          <h3 className="dd-booking__title" id="dd-booking-heading">
            Reserve your seats
          </h3>

          {show ? (
            <>
              <dl className="dd-booking__meta">
                <div>
                  <dt>{isNextShow ? 'Next show' : 'Selected show'}</dt>
                  <dd>{formatLongShowDate(show.startsAt)}</dd>
                </div>
                {show.venue ? (
                  <div>
                    <dt>Venue</dt>
                    <dd>{show.venue}</dd>
                  </div>
                ) : null}
                {formatPriceFrom(show.tiers) ? (
                  <div>
                    <dt>Tickets</dt>
                    <dd>{formatPriceFrom(show.tiers)}</dd>
                  </div>
                ) : null}
              </dl>

              {seasonEnded ? (
                <p className="dd-booking__note">
                  The season has ended. Booking opens as soon as the next date
                  is announced.
                </p>
              ) : null}

              {!href && !seasonEnded ? (
                <p className="dd-booking__note">
                  This performance is {formatStatus(show.status).toLowerCase()}.
                </p>
              ) : null}

              <ReserveButton
                className="dd-cta dd-booking__button"
                event={event}
                label="Reserve seats"
                show={show}
              />

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
      )}
    >
      {({
        canSelect,
        card,
        isSelected,
        seasons,
        selectedShow,
        selectShow,
      }) => (
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
                  {selectedShow?.venue
                    ? `Join us at ${selectedShow.venue}.`
                    : event.venue
                      ? `Join us at ${event.venue}.`
                      : 'Dates and venue will be announced soon.'}
                </p>
              </div>

              {seasons.length > 0 ? (
                <div className="dd-runs">
                  {seasons.map((season) => (
                    <div
                      className="dd-run"
                      key={`${season.label}-${season.venue}`}
                    >
                      <div className="dd-run__head">
                        <h3>{season.label}</h3>
                        {season.venue ? <p>{season.venue}</p> : null}
                      </div>

                      <ul className="dd-shows">
                        {season.shows.map((show) => {
                          const parts = dateParts(show.startsAt)
                          const meta = formatPriceFrom(show.tiers)
                          const selectable = canSelect(show)
                          const selected = isSelected(show)

                          return (
                            <li
                              aria-label={parts.full}
                              className={`dd-show dd-show--${show.status}${
                                selected ? ' is-selected' : ''
                              }`}
                              key={show.id}
                            >
                              {selectable ? (
                                <button
                                  aria-label={`Select ${parts.full}`}
                                  aria-pressed={selected}
                                  className="dd-show__select"
                                  onClick={() => selectShow(show.id)}
                                  type="button"
                                />
                              ) : null}

                              <div className="dd-show__date">
                                <span className="dd-show__weekday">
                                  {parts.weekday}
                                </span>
                                <span className="dd-show__day">{parts.day}</span>
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
                              </div>
                            </li>
                          )
                        })}
                      </ul>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="dd-empty">
                  No show dates have been announced yet. Check back soon.
                </p>
              )}
            </div>

            {card}
          </Container>
        </section>
      )}
    </SeasonBooking>
  )
}
