import type {
  BankTransferDetails,
  CheckoutInput,
  Hold,
  HoldRequest,
  OrderSeat,
  PaymentMode,
  PublicOrder,
  PublicTicket,
  Seat,
  SeatMap,
  SelectionMode,
  ShowSummary,
  StagePosition,
  Ticket,
  TicketingClient,
  TierSummary,
} from './types'
import { MAX_SEATS_PER_ORDER, TicketingError } from './types'

const HOLD_MINUTES = 10
const CHECKOUT_HOLD_MINUTES = 15
const MANUAL_HOLD_HOURS = 24
const STORAGE_KEY = 'bp.ticketing.orders.v2'
const SHOWS_KEY = 'bp.ticketing.shows.v2'

export type MockShowMeta = {
  productionSlug: string
  title: string
  startsAt: string
  venue: string
  paymentMode?: PaymentMode
  selectionMode?: SelectionMode
  stagePosition?: StagePosition
  numbering?: boolean
  columns?: number
  tiers?: {
    id: string
    name: string
    price_cents: number
    count: number
  }[]
  bankTransfer?: BankTransferDetails | null
}

type MockTier = {
  id: string
  name: string
  price_cents: number
  count: number
}

const defaultPlan: MockTier[] = [
  { id: 'front', name: 'Front', price_cents: 300000, count: 20 },
  { id: 'standard', name: 'Standard', price_cents: 250000, count: 30 },
]

const defaultBank: BankTransferDetails = {
  account_name: 'Basement Plays',
  account_number: '0000 0000 0000',
  bank_name: 'Placeholder Bank',
  branch: 'Kandy',
  instructions:
    'Use your order reference as the transfer description, then send the receipt to the box office on WhatsApp.',
}

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

function scenarios(): string[] {
  if (typeof window === 'undefined') {
    return []
  }

  const raw = new URLSearchParams(window.location.search).get('scenario')

  return raw
    ? raw
        .toLowerCase()
        .split(/[,+]/)
        .map((value) => value.trim())
        .filter(Boolean)
    : []
}

function hasScenario(...names: string[]): boolean {
  const active = scenarios()

  return names.some((name) => active.includes(name))
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

function planFor(meta: MockShowMeta | null): MockTier[] {
  return meta?.tiers?.length ? meta.tiers : defaultPlan
}

function selectionModeFor(meta: MockShowMeta | null): SelectionMode {
  if (hasScenario('seats', 'seat-map')) {
    return 'seats'
  }
  if (hasScenario('quantity', 'tiers')) {
    return 'quantity'
  }

  return meta?.selectionMode ?? 'seats'
}

function paymentModeFor(meta: MockShowMeta | null): PaymentMode {
  if (hasScenario('online', 'payhere')) {
    return 'online'
  }
  if (hasScenario('bank-transfer', 'bank_transfer')) {
    return 'bank_transfer'
  }
  if (hasScenario('off')) {
    return 'off'
  }

  return meta?.paymentMode ?? 'online'
}

function stagePositionFor(meta: MockShowMeta | null): StagePosition {
  if (hasScenario('stage-back')) {
    return 'back'
  }
  if (hasScenario('stage-none')) {
    return 'none'
  }
  if (hasScenario('stage-front')) {
    return 'front'
  }

  return meta?.stagePosition ?? 'front'
}

function buildBaseSeats(
  meta: MockShowMeta | null,
  selectionMode: SelectionMode,
): Seat[] {
  const seats: Seat[] = []
  const soldOut = hasScenario('sold-out', 'soldout', 'empty')
  const fewLeft = hasScenario('few-left', 'fewleft')
  const numbered = selectionMode === 'seats' && (meta?.numbering ?? true)
  let index = 0

  for (const tier of planFor(meta)) {
    for (let n = 0; n < tier.count; n += 1) {
      const code = `S${String(index + 1).padStart(3, '0')}`
      const status = soldOut || (fewLeft && n >= 3) ? 'sold' : 'available'

      seats.push({
        code,
        label: numbered ? String(index + 1) : null,
        tier_id: tier.id,
        tier_name: tier.name,
        price_cents: tier.price_cents,
        status,
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
    let position = index.get(seat.tier_id)

    if (position === undefined) {
      position = tiers.length
      index.set(seat.tier_id, position)
      tiers.push({
        tier_id: seat.tier_id,
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
  if (hasScenario('error', 'unavailable')) {
    throw new TicketingError(
      'service_unavailable',
      'The booking system is unavailable. Try again in a moment.',
      503,
    )
  }

  const meta = readShowMeta(showID)
  const selectionMode = selectionModeFor(meta)
  const seats = buildBaseSeats(meta, selectionMode)

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
    payment_mode: paymentModeFor(meta),
    selection_mode: selectionMode,
    stage_position: stagePositionFor(meta),
    numbering: meta?.numbering ?? true,
    currency: 'LKR',
    columns: meta?.columns ?? 10,
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
    payment_mode: map.payment_mode,
    selection_mode: map.selection_mode,
    stage_position: map.stage_position,
    numbering: map.numbering,
    currency: map.currency,
    columns: map.columns,
    stage_label: map.stage_label,
    total_seats: map.total_seats,
    seats_remaining: map.seats_remaining,
    price_from: map.tiers.length
      ? Math.min(...map.tiers.map((tier) => tier.price_cents))
      : 0,
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
    tier_id: seat.tier_id,
    tier_name: seat.tier_name,
    description,
    price_cents: seat.price_cents,
    order_reference: orderReference(order.order_id),
  })
}

function createOrder(
  showID: string,
  selected: OrderSeat[],
  createdAt: string,
  meta: MockShowMeta | null,
): PublicOrder {
  return {
    order_id: randomID(),
    show_id: showID,
    status: 'pending',
    source: 'online',
    payment_method:
      meta?.paymentMode === 'bank_transfer' ? 'bank_transfer' : 'payhere',
    total_cents: selected.reduce((total, seat) => total + seat.price_cents, 0),
    currency: 'LKR',
    expires_at: hasScenario('hold-expired')
      ? addMinutes(createdAt, -1)
      : addMinutes(createdAt, HOLD_MINUTES),
    paid_at: null,
    created_at: createdAt,
    bank_transfer: null,
    seats: selected,
    tickets: [],
  }
}

function issueTickets(
  order: PublicOrder,
  issuedAt: string,
): PublicTicket[] {
  return order.seats.map((seat) => {
    const ticket: PublicTicket = {
      ticket_id: randomID(),
      seat_code: seat.code,
      status: 'issued',
      issued_at: issuedAt,
      used_at: null,
    }
    const token = makeTicketToken(order, ticket, seat)

    return {
      ...ticket,
      qr_token: token,
      qr_url: `/t/${token}`,
    }
  })
}

function orderSeat(seat: Seat): OrderSeat {
  return {
    code: seat.code,
    label: seat.label,
    tier_id: seat.tier_id,
    tier_name: seat.tier_name,
    price_cents: seat.price_cents,
  }
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
    request: HoldRequest,
    options?: { idempotencyKey?: string },
  ): Promise<Hold> {
    void options
    const records = readStore()
    const map = buildSeatMap(showID, records)
    const meta = readShowMeta(showID)
    const selectionMode = map.selection_mode

    if (selectionMode === 'quantity') {
      if (!('tiers' in request)) {
        throw new TicketingError(
          'validation_error',
          'This performance uses tier quantities.',
        )
      }
      if (hasScenario('insufficient', 'insufficient-inventory')) {
        throw new TicketingError(
          'insufficient_inventory',
          'There are not enough seats left in those tiers.',
          409,
        )
      }

      const requested = request.tiers
      const tierIDs = new Set<string>()
      let total = 0
      const selected: OrderSeat[] = []

      for (const item of requested) {
        if (
          !item.tier_id ||
          !Number.isInteger(item.quantity) ||
          item.quantity < 1 ||
          tierIDs.has(item.tier_id)
        ) {
          throw new TicketingError(
            'validation_error',
            'Choose a valid quantity for each tier.',
          )
        }

        tierIDs.add(item.tier_id)
        total += item.quantity
      }

      if (total < 1) {
        throw new TicketingError(
          'validation_error',
          'Choose at least one ticket.',
        )
      }
      if (total > MAX_SEATS_PER_ORDER) {
        throw new TicketingError(
          'validation_error',
          `You can reserve up to ${MAX_SEATS_PER_ORDER} tickets at a time.`,
        )
      }

      for (const item of requested) {
        const available = map.seats
          .filter(
            (seat) =>
              seat.tier_id === item.tier_id && seat.status === 'available',
          )
          .sort((a, b) => a.sort_index - b.sort_index)

        if (available.length < item.quantity) {
          throw new TicketingError(
            'insufficient_inventory',
            'There are not enough seats left in those tiers.',
            409,
          )
        }

        for (const seat of available.slice(0, item.quantity)) {
          seat.status = 'held'
          selected.push(orderSeat(seat))
        }
      }

      const createdAt = nowISO()
      const order = createOrder(showID, selected, createdAt, meta)
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

    if (!('seats' in request)) {
      throw new TicketingError(
        'validation_error',
        'This performance uses seat selection.',
      )
    }
    if (hasScenario('seat-taken', 'seat_taken')) {
      throw new TicketingError(
        'seat_taken',
        'One or more selected seats are no longer available.',
        409,
      )
    }

    const seats = request.seats
    if (seats.length === 0) {
      throw new TicketingError('validation_error', 'Choose at least one seat.')
    }
    if (seats.length > MAX_SEATS_PER_ORDER) {
      throw new TicketingError(
        'validation_error',
        `You can reserve up to ${MAX_SEATS_PER_ORDER} seats at a time.`,
      )
    }

    const selected: OrderSeat[] = []

    for (const code of new Set(seats)) {
      const seat = map.seats.find((candidate) => candidate.code === code)

      if (!seat || seat.status !== 'available') {
        throw new TicketingError(
          'seat_taken',
          'One or more selected seats are no longer available.',
          409,
        )
      }

      selected.push(orderSeat(seat))
    }

    const createdAt = nowISO()
    const order = createOrder(showID, selected, createdAt, meta)
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
    const meta = readShowMeta(record.order.show_id)
    let order = record.order

    if (input.payment_method === 'bank_transfer') {
      order = {
        ...order,
        bank_transfer: meta?.bankTransfer ?? defaultBank,
        expires_at: addHours(checkedOutAt, MANUAL_HOLD_HOURS),
        payment_method: 'bank_transfer',
        source: 'manual',
      }
    } else {
      // The mock pretends the sandbox payment succeeded immediately. The real
      // PayHere flow keeps the order pending until the signed notification
      // arrives from the ticketing service.
      order = {
        ...order,
        bank_transfer: null,
        status: 'paid',
        paid_at: checkedOutAt,
        expires_at: addMinutes(checkedOutAt, CHECKOUT_HOLD_MINUTES),
        payment_method: 'payhere',
        tickets: issueTickets(order, checkedOutAt),
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

  async devMarkPaid(
    orderID: string,
    orderToken: string,
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

    const paidAt = nowISO()
    const tickets = record.order.tickets.length
      ? record.order.tickets
      : issueTickets(record.order, paidAt)

    record.order = {
      ...record.order,
      status: 'paid',
      paid_at: paidAt,
      tickets,
    }
    records.set(orderID, record)
    writeStore(records)

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
