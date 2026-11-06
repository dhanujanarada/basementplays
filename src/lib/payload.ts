import config from '@payload-config'
import { getPayload } from 'payload'

import type { Production, SiteSetting } from '@/payload-types'

import {
  deriveShowStatus,
  deriveTicketingActive,
  sortShowsAscending,
  type EventTier,
  type EventView,
} from './event-view'

async function loadEventViews(
  productionDocs: Production[],
  settings: SiteSetting,
): Promise<EventView[]> {
  return productionDocs.map((production) => toEventView(production, settings))
}

function toEventView(production: Production, settings: SiteSetting): EventView {
  const tiers: EventTier[] = (production.ticketing?.tiers ?? []).map(
    (tier) => ({
      count: tier.count,
      id: tier.id ?? '',
      name: tier.name,
      price: tier.price,
    }),
  )
  const ticketingActive = deriveTicketingActive(
    production.ticketing.paymentMode,
    tiers,
  )
  const priceFrom = tiers.length
    ? Math.min(...tiers.map((tier) => tier.price))
    : null

  const shows = sortShowsAscending(
    (production.shows ?? []).map((show) => {
      const venue = show.venue || production.venue || ''
      const eventShow = {
        cancelled: show.cancelled === true,
        id: show.id ?? '',
        priceFrom,
        seatsRemaining: null,
        startsAt: show.startsAt,
        status: deriveShowStatus(
          {
            cancelled: show.cancelled === true,
            seatsRemaining: null,
            startsAt: show.startsAt,
          },
          ticketingActive,
        ),
        tiers,
        venue,
      }

      return eventShow
    }),
  )

  return {
    bankTransfer: settings.bankTransfer ?? null,
    cast: production.cast ?? [],
    crew: production.crew ?? [],
    directedBy: production.directedBy ?? null,
    gallery: production.gallery ?? [],
    heroImage: production.heroImage,
    id: production.id,
    partners: production.partners ?? [],
    poster: production.poster,
    presentedBy: production.presentedBy ?? null,
    shows,
    slug: production.slug,
    synopsis: production.synopsis,
    tagline: production.tagline ?? null,
    ticketing: {
      active: ticketingActive,
      columns: production.ticketing.columns ?? 10,
      numbering: production.ticketing.numbering === true,
      paymentMode: production.ticketing.paymentMode,
      selectionMode: production.ticketing.selectionMode,
      stagePosition: production.ticketing.stagePosition,
      tiers,
    },
    title: production.title,
    venue: production.venue ?? null,
    whatsappNumber: settings.whatsappNumber ?? null,
    writtenBy: production.writtenBy ?? null,
  }
}

export async function getAllEventViews(): Promise<EventView[]> {
  const payload = await getPayload({ config })

  const [productions, settings] = await Promise.all([
    payload.find({
      collection: 'productions',
      depth: 2,
      limit: 100,
      overrideAccess: false,
      sort: 'title',
    }),
    payload.findGlobal({
      slug: 'site-settings',
      depth: 2,
      overrideAccess: false,
    }),
  ])

  return loadEventViews(productions.docs, settings)
}

export async function getEventView(slug: string): Promise<EventView | null> {
  const payload = await getPayload({ config })

  const [productions, settings] = await Promise.all([
    payload.find({
      collection: 'productions',
      depth: 2,
      limit: 1,
      overrideAccess: false,
      where: {
        slug: {
          equals: slug,
        },
      },
    }),
    payload.findGlobal({
      slug: 'site-settings',
      depth: 2,
      overrideAccess: false,
    }),
  ])

  const production = productions.docs[0]

  return production ? toEventView(production, settings) : null
}

export async function getSiteSettings(): Promise<SiteSetting> {
  const payload = await getPayload({ config })

  return payload.findGlobal({
    slug: 'site-settings',
    depth: 2,
    overrideAccess: false,
  })
}
