'use client'

import { useSearchParams } from 'next/navigation'
import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

import {
  getNextShow,
  groupShowsBySeason,
  sortShowsAscending,
  type EventSeason,
  type EventShow,
  type EventView,
} from '@/lib/event-view'
import {
  bookingHref,
  isExternalBookingHref,
} from '@/lib/ticketing/booking'

type SeasonCardRenderState = {
  event: EventView
  external: boolean
  href: string | null
  isNextShow: boolean
  seasonEnded: boolean
  show: EventShow | null
}

type SeasonBookingRenderState = {
  canSelect: (show: EventShow) => boolean
  card: ReactNode
  event: EventView
  isSelected: (show: EventShow) => boolean
  seasonEnded: boolean
  seasons: EventSeason[]
  selectedShow: EventShow | null
  selectedShowId: string | null
  selectShow: (showID: string) => void
  shows: EventShow[]
}

type SeasonBookingProps = {
  event: EventView
  renderCard: (state: SeasonCardRenderState) => ReactNode
  children: (state: SeasonBookingRenderState) => ReactNode
}

function latestShow(event: EventView): EventShow | null {
  return sortShowsAscending(event.shows).at(-1) ?? null
}

export function SeasonBooking({
  event,
  renderCard,
  children,
}: SeasonBookingProps) {
  const searchParams = useSearchParams()
  const nextShow = useMemo(() => getNextShow(event), [event])
  const fallbackShow = nextShow ?? latestShow(event)
  const requestedShowID = searchParams.get('show')
  const [selectedShowId, setSelectedShowId] = useState<string | null>(() => {
    const requested = requestedShowID
      ? event.shows.find((show) => show.id === requestedShowID)
      : null

    return (requested ?? fallbackShow)?.id ?? null
  })

  const selectedShow =
    event.shows.find((show) => show.id === selectedShowId) ?? fallbackShow
  const seasonEnded = nextShow === null

  useEffect(() => {
    if (!selectedShowId || typeof window === 'undefined') {
      return
    }

    const url = new URL(window.location.href)
    if (url.searchParams.get('show') === selectedShowId) {
      return
    }

    url.searchParams.set('show', selectedShowId)
    window.history.replaceState(
      null,
      '',
      `${url.pathname}${url.search}${url.hash}`,
    )
  }, [selectedShowId])

  const selectShow = useCallback(
    (showID: string) => {
      const show = event.shows.find((candidate) => candidate.id === showID)

      if (!show || show.status === 'past') {
        return
      }

      setSelectedShowId(show.id)
    },
    [event.shows],
  )

  const canSelect = useCallback(
    (show: EventShow) => show.status !== 'past',
    [],
  )
  const isSelected = useCallback(
    (show: EventShow) => show.id === selectedShow?.id,
    [selectedShow?.id],
  )

  const href = bookingHref(event, selectedShow)
  const card = renderCard({
    event,
    external: href ? isExternalBookingHref(href) : false,
    href,
    isNextShow: Boolean(
      selectedShow && nextShow && selectedShow.id === nextShow.id,
    ),
    seasonEnded,
    show: selectedShow,
  })

  return children({
    canSelect,
    card,
    event,
    isSelected,
    seasonEnded,
    seasons: groupShowsBySeason(event.shows),
    selectedShow,
    selectedShowId,
    selectShow,
    shows: sortShowsAscending(event.shows),
  })
}
