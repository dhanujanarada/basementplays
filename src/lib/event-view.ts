import type { Production, SiteSetting } from '@/payload-types'

export type EventTier = {
  count: number
  id: string
  name: string
  price: number
}

export type EventShowStatus =
  | 'open'
  | 'few-left'
  | 'sold-out'
  | 'cancelled'
  | 'past'

export type EventShow = {
  cancelled: boolean
  id: string
  priceFrom: number | null
  seatsRemaining: number | null
  startsAt: string
  status: EventShowStatus
  tiers: EventTier[]
  venue: string
}

type EventTicketing = {
  active: boolean
  columns: number
  numbering: boolean
  paymentMode: Production['ticketing']['paymentMode']
  selectionMode: Production['ticketing']['selectionMode']
  stagePosition: Production['ticketing']['stagePosition']
  tiers: EventTier[]
}

export type EventView = {
  bankTransfer: NonNullable<SiteSetting['bankTransfer']> | null
  cast: NonNullable<Production['cast']>
  crew: NonNullable<Production['crew']>
  directedBy: string | null
  gallery: NonNullable<Production['gallery']>
  heroImage: Production['heroImage']
  id: number
  partners: NonNullable<Production['partners']>
  poster: Production['poster']
  presentedBy: string | null
  shows: EventShow[]
  slug: string
  synopsis: Production['synopsis']
  tagline: string | null
  ticketing: EventTicketing
  title: string
  venue: string | null
  whatsappNumber: string | null
  writtenBy: string | null
}

export type EventSeason = {
  label: string
  shows: EventShow[]
  venue: string
}

const FEW_LEFT_THRESHOLD = 10

export function deriveTicketingActive(
  paymentMode: Production['ticketing']['paymentMode'],
  tiers: EventTier[],
): boolean {
  return (
    paymentMode !== 'off' &&
    tiers.some((tier) => tier.price > 0 && tier.count > 0)
  )
}

export function deriveShowStatus(
  show: Pick<EventShow, 'cancelled' | 'seatsRemaining' | 'startsAt'>,
  ticketingActive: boolean,
  now: Date = new Date(),
): EventShowStatus {
  if (show.cancelled) {
    return 'cancelled'
  }

  const startsAt = new Date(show.startsAt).getTime()
  if (!Number.isNaN(startsAt) && startsAt < now.getTime()) {
    return 'past'
  }

  if (ticketingActive && show.seatsRemaining !== null) {
    if (show.seatsRemaining <= 0) {
      return 'sold-out'
    }
    if (show.seatsRemaining <= FEW_LEFT_THRESHOLD) {
      return 'few-left'
    }
  }

  return 'open'
}

export function sortShowsAscending(shows: EventShow[]): EventShow[] {
  return [...shows].sort(
    (a, b) => new Date(a.startsAt).getTime() - new Date(b.startsAt).getTime(),
  )
}

export function getNextShow(
  event: EventView,
  now: Date = new Date(),
): EventShow | null {
  return (
    sortShowsAscending(event.shows).find(
      (show) =>
        show.status !== 'cancelled' &&
        show.status !== 'past' &&
        new Date(show.startsAt).getTime() >= now.getTime(),
    ) ?? null
  )
}

export function getBookableShows(
  event: EventView,
  now: Date = new Date(),
): EventShow[] {
  if (!event.ticketing.active) {
    return []
  }

  return sortShowsAscending(event.shows).filter(
    (show) =>
      show.status !== 'cancelled' &&
      show.status !== 'past' &&
      show.status !== 'sold-out' &&
      show.seatsRemaining !== 0 &&
      new Date(show.startsAt).getTime() >= now.getTime(),
  )
}

const seasonYearFormatter = new Intl.DateTimeFormat('en-GB', {
  timeZone: 'Asia/Colombo',
  year: 'numeric',
})

export function showSeasonLabel(startsAt: string): string {
  const date = new Date(startsAt)

  return Number.isNaN(date.getTime())
    ? 'Season to be announced'
    : `${seasonYearFormatter.format(date)} Season`
}

export function groupShowsBySeason(shows: EventShow[]): EventSeason[] {
  const groups = new Map<string, EventSeason>()

  for (const show of shows) {
    const label = showSeasonLabel(show.startsAt)
    const key = `${label}|${show.venue}`
    const group = groups.get(key)

    if (group) {
      group.shows.push(show)
    } else {
      groups.set(key, {
        label,
        shows: [show],
        venue: show.venue,
      })
    }
  }

  return [...groups.values()]
    .map((group) => ({
      ...group,
      shows: sortShowsAscending(group.shows),
    }))
    .sort((a, b) => latestTime(b.shows) - latestTime(a.shows))
}

function latestTime(shows: EventShow[]): number {
  return Math.max(
    ...shows.map((show) => {
      const time = new Date(show.startsAt).getTime()
      return Number.isNaN(time) ? 0 : time
    }),
  )
}
