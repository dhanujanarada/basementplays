import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { cache } from 'react'

import { OrderConfirmation } from '@/components/ticketing/OrderConfirmation'
import { designs, plainDesign } from '@/designs'
import { getProduction, getRunsForProduction } from '@/lib/payload'

export const dynamic = 'force-dynamic'

type ConfirmationPageProps = {
  params: Promise<{ slug: string }>
  searchParams: Promise<Record<string, string | string[] | undefined>>
}

const getProductionBySlug = cache(getProduction)

export async function generateMetadata({
  params,
}: ConfirmationPageProps): Promise<Metadata> {
  const { slug } = await params
  const production = await getProductionBySlug(slug)

  return {
    title: production
      ? `Booking confirmation · ${production.title}`
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
  const production = await getProductionBySlug(slug)

  if (!production) {
    notFound()
  }
  if (!orderID) {
    notFound()
  }

  const runs = await getRunsForProduction(production.id)
  const design = designs[production.slug] ?? plainDesign
  const { Shell } = design

  return (
    <Shell production={production} runs={runs}>
      <OrderConfirmation
        orderID={orderID}
        productionSlug={production.slug}
        productionTitle={production.title}
      />
    </Shell>
  )
}
