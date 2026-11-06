import type {
  CheckoutInput,
  Hold,
  OrderSeat,
  PublicOrder,
  PublicTicket,
  Seat,
  SeatMap,
  ShowSummary,
  Ticket,
  TicketingClient,
  TierSummary,
} from './types'
import { MAX_SEATS_PER_ORDER, TicketingError } from './types'

const HOLD_MINUTES = 10
const CHECKOUT_HOLD_MINUTES = 15
const MANUAL_HOLD_HOURS = 24
const STORAGE_KEY = 'bp.ticketing.orders.v1'
const SHOWS_KEY = 'bp.ticketing.shows.v1'

export type MockShowMeta = {
  productionSlug: string
  title: string
  startsAt: string
  venue: string
}

type TierDefinition = {
  key: string
  name: string
  price_cents: number
  seat_count: number
}

const plan: TierDefinition[] = [
  { key: 'front', name: 'Front', price_cents: 300000, seat_count: 20 },
  { key: 'standard', name: 'Standard', price_cents: 250000, seat_count: 30 },
]

type MockRecord = {
  order: PublicOrder
  token: string
}

type MockTicketPayload = Omit<Ticket, 'qr_token' | 'qr_url'>

const memoryStore = new Map<string, MockRecord>()

function readStore(): Map<string, MockRecord> {
  if (typeof window === 'undefined') {
    return new Map(memoryStore)
  }

  try {
    const raw = window.sessionStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return new Map()
    }

    const parsed = JSON.parse(raw) as Record<string, MockRecord>
    return new Map(Object.entries(parsed))
  } catch {
    return new Map()
  }
}

function writeStore(records: Map<string, MockRecord>): void {
  if (typeof window === 'undefined') {
    memoryStore.clear()
    for (const [key, value] of records) {
      memoryStore.set(key, value)
    }
    return
  }

  window.sessionStorage.setItem(
    STORAGE_KEY,
    JSON.stringify(Object.fromEntries(records)),
  )
}

function randomID(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }

  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

function secretToken(): string {
  return `${randomID()}.${randomID()}`
}

function nowISO(): string {
  return new Date().toISOString()
}

function addMinutes(iso: string, minutes: number): string {
  return new Date(new Date(iso).getTime() + minutes * 60_000).toISOString()
}

function addHours(iso: string, hours: number): string {
  return new Date(new Date(iso).getTime() + hours * 3_600_000).toISOString()
}

function readShowMeta(showID: string): MockShowMeta | null {
  if (typeof window === 'undefined') {
    return null
  }

  try {
    const raw = window.sessionStorage.getItem(SHOWS_KEY)
    if (!raw) {
      return null
    }

    const parsed = JSON.parse(raw) as Record<string, MockShowMeta>
    return parsed[showID] ?? null
  } catch {
    return null
  }
}

// The mock needs the Payload-side show details to embed them in a ticket
// token. The booking page calls this before creating a hold. The real service
// already stores those details, so this has no HTTP equivalent.
export function rememberShow(showID: string, meta: MockShowMeta): void {
  if (typeof window === 'undefined') {
    return
  }

  try {
    const raw = window.sessionStorage.getItem(SHOWS_KEY)
    const parsed = raw
      ? (JSON.parse(raw) as Record<string, MockShowMeta>)
      : {}
    parsed[showID] = meta
    window.sessionStorage.setItem(SHOWS_KEY, JSON.stringify(parsed))
  } catch {
    // The mock still works without show details; the ticket page falls back to
    // whatever the token carries.
  }
}

function effectiveStatus(order: PublicOrder): PublicOrder['status'] {
  if (
    order.status === 'pending' &&
    new Date(order.expires_at).getTime() <= Date.now()
  ) {
    return 'expired'
  }

  return order.status
}

function buildBaseSeats(): Seat[] {
  const seats: Seat[] = []
  let index = 0

  for (const tier of plan) {
    for (let n = 0; n < tier.seat_count; n += 1) {
      const code = `S${String(index + 1).padStart(3, '0')}`
      const sold = index % 7 === 3
      const blocked = index === 49

      seats.push({
        code,
        label: String(index + 1),
        tier_key: tier.key,
        tier_name: tier.name,
        price_cents: tier.price_cents,
        status: blocked ? 'blocked' : sold ? 'sold' : 'available',
        sort_index: index,
      })

      index += 1
    }
  }

  return seats
}

function summarize(seats: Seat[]): {
  tiers: TierSummary[]
  total: number
  remaining: number
} {
  const tiers: TierSummary[] = []
  const index = new Map<string, number>()
  let remaining = 0

  for (const seat of seats) {
    let position = index.get(seat.tier_key)

    if (position === undefined) {
      position = tiers.length
      index.set(seat.tier_key, position)
      tiers.push({
        key: seat.tier_key,
        name: seat.tier_name,
        price_cents: seat.price_cents,
        seats_total: 0,
        seats_remaining: 0,
      })
    }

    const tier = tiers[position]
    tier.seats_total += 1

    if (seat.status === 'available') {
      tier.seats_remaining += 1
      remaining += 1
    }
  }

  return { tiers, total: seats.length, remaining }
}

function buildSeatMap(showID: string, records: Map<string, MockRecord>): SeatMap {
  const seats = buildBaseSeats()

  for (const record of records.values()) {
    if (record.order.show_id !== showID) {
      continue
    }

    const status = effectiveStatus(record.order)
    if (status !== 'pending' && status !== 'paid') {
      continue
    }

    for (const seat of record.order.seats) {
      const match = seats.find((candidate) => candidate.code === seat.code)
      if (!match || match.status === 'blocked') {
        continue
      }

      match.status = status === 'paid' ? 'sold' : 'held'
    }
  }

  const { tiers, total, remaining } = summarize(seats)

  return {
    show_id: showID,
    status: 'on_sale',
    currency: 'LKR',
    columns: 10,
    stage_label: 'Stage',
    total_seats: total,
    seats_remaining: remaining,
    tiers,
    seats,
  }
}

function toShowSummary(map: SeatMap, showID: string): ShowSummary {
  const meta = readShowMeta(showID)

  return {
    id: showID,
    production_slug: meta?.productionSlug ?? '',
    title: meta?.title ?? '',
    starts_at: meta?.startsAt ?? '',
    venue: meta?.venue ?? '',
    status: map.status,
    currency: map.currency,
    columns: map.columns,
    stage_label: map.stage_label,
    total_seats: map.total_seats,
    seats_remaining: map.seats_remaining,
    price_from: map.tiers[0]?.price_cents ?? 0,
    tiers: map.tiers,
  }
}

function orderReference(orderID: string): string {
  return orderID.replace(/-/g, '').slice(-6).toUpperCase()
}

function encodePayload(payload: MockTicketPayload): string {
  const json = encodeURIComponent(JSON.stringify(payload))
  const base64 = btoa(json)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')

  return `mock.${base64}`
}

function decodePayload(token: string): MockTicketPayload {
  if (!token.startsWith('mock.')) {
    throw new TicketingError('ticket_not_found', 'Ticket not found', 404)
  }

  try {
    const base64 = token
      .slice('mock.'.length)
      .replace(/-/g, '+')
      .replace(/_/g, '/')
    const json = decodeURIComponent(atob(base64))
    return JSON.parse(json) as MockTicketPayload
  } catch {
    throw new TicketingError('ticket_not_found', 'Ticket not found', 404)
  }
}

function makeTicketToken(
  order: PublicOrder,
  ticket: PublicTicket,
  seat: OrderSeat,
): string {
  const label = seat.label ?? ''
  const description = label
    ? `Seat ${label} (${seat.tier_name})`
    : `${seat.tier_name} admission`
  const meta = readShowMeta(order.show_id)

  return encodePayload({
    ticket_id: ticket.ticket_id,
    status: ticket.status,
    production_slug: meta?.productionSlug ?? '',
    show_id: order.show_id,
    show_title: meta?.title ?? '',
    starts_at: meta?.startsAt ?? '',
    venue: meta?.venue ?? '',
    seat_code: seat.code,
    seat_label: seat.label,
    tier_key: seat.tier_key,
    tier_name: seat.tier_name,
    description,
    price_cents: seat.price_cents,
    order_reference: orderReference(order.order_id),
  })
}

export class MockTicketingClient implements TicketingClient {
  async getShowSummary(showID: string): Promise<ShowSummary> {
    const map = await this.getSeatMap(showID)
    return toShowSummary(map, showID)
  }

  async getSeatMap(showID: string): Promise<SeatMap> {
    return buildSeatMap(showID, readStore())
  }

  async createHold(
    showID: string,
    seats: string[],
  ): Promise<Hold> {
    if (seats.length === 0) {
      throw new TicketingError('validation_error', 'Choose at least one seat.')
    }
    if (seats.length > MAX_SEATS_PER_ORDER) {
      throw new TicketingError(
        'validation_error',
        `You can reserve up to ${MAX_SEATS_PER_ORDER} seats at a time.`,
      )
    }

    const records = readStore()
    const map = buildSeatMap(showID, records)
    const selected: OrderSeat[] = []

    for (const code of seats) {
      const seat = map.seats.find((candidate) => candidate.code === code)

      if (!seat || seat.status !== 'available') {
        throw new TicketingError(
          'seat_taken',
          'One or more selected seats are no longer available.',
          409,
        )
      }

      selected.push({
        code: seat.code,
        label: seat.label,
        tier_key: seat.tier_key,
        tier_name: seat.tier_name,
        price_cents: seat.price_cents,
      })
    }

    const createdAt = nowISO()
    const order: PublicOrder = {
      order_id: randomID(),
      show_id: showID,
      status: 'pending',
      source: 'online',
      total_cents: selected.reduce((total, seat) => total + seat.price_cents, 0),
      currency: 'LKR',
      expires_at: addMinutes(createdAt, HOLD_MINUTES),
      paid_at: null,
      created_at: createdAt,
      seats: selected,
      tickets: [],
    }
    const token = secretToken()

    records.set(order.order_id, { order, token })
    writeStore(records)

    return {
      order_id: order.order_id,
      show_id: order.show_id,
      status: order.status,
      source: order.source,
      total_cents: order.total_cents,
      currency: order.currency,
      expires_at: order.expires_at,
      seats: order.seats,
      order_token: token,
    }
  }

  async releaseOrder(orderID: string, orderToken: string): Promise<void> {
    const records = readStore()
    const record = records.get(orderID)

    if (!record) {
      throw new TicketingError('not_found', 'Order not found', 404)
    }
    if (record.token !== orderToken) {
      throw new TicketingError(
        'invalid_order_token',
        'This order belongs to a different browser session.',
        403,
      )
    }
    if (effectiveStatus(record.order) !== 'pending') {
      throw new TicketingError(
        'order_not_pending',
        'This order can no longer be released.',
        409,
      )
    }

    record.order = { ...record.order, status: 'cancelled' }
    records.set(orderID, record)
    writeStore(records)
  }

  async checkout(
    orderID: string,
    orderToken: string,
    input: CheckoutInput,
  ): Promise<PublicOrder> {
    const records = readStore()
    const record = records.get(orderID)

    if (!record) {
      throw new TicketingError('not_found', 'Order not found', 404)
    }
    if (record.token !== orderToken) {
      throw new TicketingError(
        'invalid_order_token',
        'This order belongs to a different browser session.',
        403,
      )
    }
    if (effectiveStatus(record.order) !== 'pending') {
      throw new TicketingError(
        'order_not_checkoutable',
        'This hold has expired. Choose your seats again.',
        409,
      )
    }

    const name = input.buyer_name.trim()
    const phone = input.buyer_phone.trim()
    const email = input.buyer_email.trim()

    if (!name || !phone) {
      throw new TicketingError(
        'validation_error',
        'Name and phone number are required.',
      )
    }
    if (email && !/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email)) {
      throw new TicketingError(
        'validation_error',
        'Enter a valid email address.',
      )
    }

    const checkedOutAt = nowISO()
    let order = record.order

    if (input.payment_method === 'bank_transfer') {
      order = {
        ...order,
        source: 'manual',
        expires_at: addHours(checkedOutAt, MANUAL_HOLD_HOURS),
      }
    } else {
      // The mock pretends the sandbox payment succeeded immediately. The real
      // PayHere flow keeps the order pending until the signed notification
      // arrives from the ticketing service.
      const tickets: PublicTicket[] = order.seats.map((seat) => ({
        ticket_id: randomID(),
        seat_code: seat.code,
        status: 'issued',
        issued_at: checkedOutAt,
        used_at: null,
      }))

      order = {
        ...order,
        status: 'paid',
        paid_at: checkedOutAt,
        expires_at: addMinutes(checkedOutAt, CHECKOUT_HOLD_MINUTES),
        tickets,
      }
    }

    record.order = order
    records.set(orderID, record)
    writeStore(records)

    return order
  }

  async getOrder(orderID: string, orderToken: string): Promise<PublicOrder> {
    const records = readStore()
    const record = records.get(orderID)

    if (!record) {
      throw new TicketingError('not_found', 'Order not found', 404)
    }
    if (record.token !== orderToken) {
      throw new TicketingError(
        'invalid_order_token',
        'This order belongs to a different browser session.',
        403,
      )
    }

    const status = effectiveStatus(record.order)
    if (status !== record.order.status) {
      record.order = { ...record.order, status }
      records.set(orderID, record)
      writeStore(records)
    }

    return record.order
  }

  async getTicket(token: string): Promise<Ticket> {
    const payload = decodePayload(token)
    return {
      ...payload,
      qr_token: token,
      qr_url: `/t/${token}`,
    }
  }

  ticketHref(order: PublicOrder, ticket: PublicTicket): string {
    const seat = order.seats.find(
      (candidate) => candidate.code === ticket.seat_code,
    )

    if (!seat) {
      throw new TicketingError('ticket_not_found', 'Ticket not found', 404)
    }

    return `/t/${makeTicketToken(order, ticket, seat)}`
  }
}
