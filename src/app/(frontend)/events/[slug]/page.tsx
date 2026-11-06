import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { cache } from 'react'

import { designs, plainDesign } from '@/designs'
import { getEventView } from '@/lib/payload'

export const dynamic = 'force-dynamic'

type EventPageProps = {
  params: Promise<{ slug: string }>
}

const getEventBySlug = cache(getEventView)

export async function generateMetadata({
  params,
}: EventPageProps): Promise<Metadata> {
  const { slug } = await params
  const event = await getEventBySlug(slug)

  if (!event) {
    return {
      title: 'Event not found',
    }
  }

  return {
    description: event.tagline ?? undefined,
    title: event.title,
  }
}

export default async function EventPage({ params }: EventPageProps) {
  const { slug } = await params
  const event = await getEventBySlug(slug)

  if (!event) {
    notFound()
  }

  const design = designs[event.slug] ?? plainDesign
  const { Page, Shell } = design

  return (
    <Shell event={event}>
      <Page event={event} />
    </Shell>
  )
}
