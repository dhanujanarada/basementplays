import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { cache } from 'react'

import { OrderConfirmation } from '@/components/ticketing/OrderConfirmation'
import { designs, plainDesign } from '@/designs'
import { getEventView } from '@/lib/payload'

export const dynamic = 'force-dynamic'

type ConfirmationPageProps = {
  params: Promise<{ slug: string }>
  searchParams: Promise<Record<string, string | string[] | undefined>>
}

const getEventBySlug = cache(getEventView)

export async function generateMetadata({
  params,
}: ConfirmationPageProps): Promise<Metadata> {
  const { slug } = await params
  const event = await getEventBySlug(slug)

  return {
    title: event
      ? `Booking confirmation · ${event.title}`
      : 'Booking confirmation',
  }
}

export default async function ConfirmationPage({
  params,
  searchParams,
}: ConfirmationPageProps) {
  const { slug } = await params
  const query = await searchParams
  const orderID = typeof query.order === 'string' ? query.order : null
  const event = await getEventBySlug(slug)

  if (!event) {
    notFound()
  }
  if (!orderID) {
    notFound()
  }

  const design = designs[event.slug] ?? plainDesign
  const { Shell } = design

  return (
    <Shell event={event}>
      <OrderConfirmation
        orderID={orderID}
        productionSlug={event.slug}
        productionTitle={event.title}
      />
    </Shell>
  )
}
