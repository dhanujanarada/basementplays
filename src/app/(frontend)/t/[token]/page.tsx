import type { Metadata } from 'next'
import { headers } from 'next/headers'
import { notFound } from 'next/navigation'
import { cache } from 'react'
import QRCode from 'qrcode'

import { TicketView } from '@/components/ticketing/TicketView'
import { designs, plainDesign } from '@/designs'
import { getProduction, getRunsForProduction } from '@/lib/payload'
import { getTicketingClient } from '@/lib/ticketing/client'

export const dynamic = 'force-dynamic'

type TicketPageProps = {
  params: Promise<{ token: string }>
}

const getProductionBySlug = cache(getProduction)

async function loadTicket(token: string) {
  try {
    return await getTicketingClient().getTicket(token)
  } catch {
    return null
  }
}

const getTicketByToken = cache(loadTicket)

export async function generateMetadata({
  params,
}: TicketPageProps): Promise<Metadata> {
  const { token } = await params
  const ticket = await getTicketByToken(token)

  return {
    title: ticket?.show_title ? `Ticket · ${ticket.show_title}` : 'Ticket',
    robots: {
      follow: false,
      index: false,
    },
  }
}

export default async function TicketPage({ params }: TicketPageProps) {
  const { token } = await params
  const ticket = await getTicketByToken(token)

  if (!ticket) {
    notFound()
  }

  const production = ticket.production_slug
    ? await getProductionBySlug(ticket.production_slug)
    : null

  if (!production) {
    notFound()
  }

  const runs = await getRunsForProduction(production.id)
  const design = designs[production.slug] ?? plainDesign
  const { Shell } = design

  const requestHeaders = await headers()
  const host =
    process.env.NEXT_PUBLIC_SITE_URL ??
    requestHeaders.get('x-forwarded-host') ??
    requestHeaders.get('host') ??
    'localhost:3000'
  const protocol =
    requestHeaders.get('x-forwarded-proto') ??
    (process.env.NODE_ENV === 'production' ? 'https' : 'http')
  const qrUrl = host.startsWith('http')
    ? `${host.replace(/\/$/, '')}/t/${token}`
    : `${protocol}://${host}/t/${token}`

  let qrSvg: string | null = null
  try {
    qrSvg = await QRCode.toString(qrUrl, {
      color: {
        dark: '#241f19',
        light: '#ffffff',
      },
      errorCorrectionLevel: 'M',
      margin: 1,
      type: 'svg',
      width: 240,
    })
  } catch {
    // The page still shows the seat and order details without the image.
  }

  return (
    <Shell production={production} runs={runs}>
      <TicketView qrSvg={qrSvg} ticket={ticket} />
    </Shell>
  )
}
