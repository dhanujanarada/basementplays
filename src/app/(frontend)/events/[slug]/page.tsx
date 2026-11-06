import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { cache } from 'react'

import { designs, plainDesign } from '@/designs'
import { getProduction, getRunsForProduction } from '@/lib/payload'

export const dynamic = 'force-dynamic'

type EventPageProps = {
  params: Promise<{ slug: string }>
}

const getProductionBySlug = cache(getProduction)

export async function generateMetadata({
  params,
}: EventPageProps): Promise<Metadata> {
  const { slug } = await params
  const production = await getProductionBySlug(slug)

  if (!production) {
    return {
      title: 'Event not found',
    }
  }

  return {
    description: production.tagline ?? undefined,
    title: production.title,
  }
}

export default async function EventPage({ params }: EventPageProps) {
  const { slug } = await params
  const production = await getProductionBySlug(slug)

  if (!production) {
    notFound()
  }

  const runs = await getRunsForProduction(production.id)
  const design = designs[production.slug] ?? plainDesign
  const { Page, Shell } = design

  return (
    <Shell production={production} runs={runs}>
      <Page production={production} runs={runs} />
    </Shell>
  )
}
