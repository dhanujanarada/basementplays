import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { cache } from 'react'

import { BookingShowPicker } from '@/components/ticketing/BookingShowPicker'
import { DefaultBooking } from '@/components/ticketing/DefaultBooking'
import { designs, plainDesign } from '@/designs'
import { getProduction, getRunsForProduction } from '@/lib/payload'
import { getBookableShows } from '@/lib/ticketing/booking'

export const dynamic = 'force-dynamic'

type BookingPageProps = {
  params: Promise<{ slug: string }>
  searchParams: Promise<Record<string, string | string[] | undefined>>
}

const getProductionBySlug = cache(getProduction)

export async function generateMetadata({
  params,
}: BookingPageProps): Promise<Metadata> {
  const { slug } = await params
  const production = await getProductionBySlug(slug)

  return {
    title: production ? `Reserve seats · ${production.title}` : 'Booking',
  }
}

export default async function BookingPage({
  params,
  searchParams,
}: BookingPageProps) {
  const { slug } = await params
  const query = await searchParams
  const production = await getProductionBySlug(slug)

  if (!production) {
    notFound()
  }

  const runs = await getRunsForProduction(production.id)
  const bookable = getBookableShows(runs)
  const requested = typeof query.show === 'string' ? query.show : undefined
  const selected = requested
    ? (bookable.find((entry) => entry.show.id === requested) ?? null)
    : bookable.length === 1
      ? bookable[0]
      : null

  const design = designs[production.slug] ?? plainDesign
  const { Shell } = design
  const Booking = design.Booking ?? DefaultBooking

  return (
    <Shell production={production} runs={runs}>
      {selected ? (
        <Booking
          production={production}
          run={selected.run}
          runs={runs}
          show={selected.show}
        />
      ) : (
        <BookingShowPicker production={production} shows={bookable} />
      )}
    </Shell>
  )
}
