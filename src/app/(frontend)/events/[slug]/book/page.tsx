import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { cache } from 'react'

import { BookingShowPicker } from '@/components/ticketing/BookingShowPicker'
import { DefaultBooking } from '@/components/ticketing/DefaultBooking'
import { designs, plainDesign } from '@/designs'
import { getBookableShows } from '@/lib/event-view'
import { getEventView } from '@/lib/payload'

export const dynamic = 'force-dynamic'

type BookingPageProps = {
  params: Promise<{ slug: string }>
  searchParams: Promise<Record<string, string | string[] | undefined>>
}

const getEventBySlug = cache(getEventView)

export async function generateMetadata({
  params,
}: BookingPageProps): Promise<Metadata> {
  const { slug } = await params
  const event = await getEventBySlug(slug)

  return {
    title: event ? `Reserve seats · ${event.title}` : 'Booking',
  }
}

export default async function BookingPage({
  params,
  searchParams,
}: BookingPageProps) {
  const { slug } = await params
  const query = await searchParams
  const event = await getEventBySlug(slug)

  if (!event) {
    notFound()
  }

  const bookable = getBookableShows(event)
  const requested = typeof query.show === 'string' ? query.show : undefined
  const selected = requested
    ? (bookable.find((show) => show.id === requested) ?? null)
    : bookable.length === 1
      ? bookable[0]
      : null

  const design = designs[event.slug] ?? plainDesign
  const { Shell } = design
  const Booking = design.Booking ?? DefaultBooking

  return (
    <Shell event={event}>
      {selected ? (
        <Booking event={event} show={selected} />
      ) : (
        <BookingShowPicker event={event} shows={bookable} />
      )}
    </Shell>
  )
}
