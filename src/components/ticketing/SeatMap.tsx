'use client'

import type { CSSProperties } from 'react'

import { formatPrice } from '@/lib/ticketing/format'
import type { SeatMap as SeatMapData, SeatStatus } from '@/lib/ticketing/types'

import { StageBar } from './StageBar'

import './ticketing.css'

type SeatMapProps = {
  disabled?: boolean
  map: SeatMapData
  onToggle: (code: string) => void
  selected: string[]
}

const statusLabels: Record<SeatStatus, string> = {
  available: 'available',
  held: 'on hold',
  sold: 'sold',
  blocked: 'unavailable',
}

export function SeatMap({
  disabled = false,
  map,
  onToggle,
  selected,
}: SeatMapProps) {
  const selectedSet = new Set(selected)
  const tiers =
    map.stage_position === 'back' ? [...map.tiers].reverse() : map.tiers

  return (
    <div>
      <StageBar label={map.stage_label} position={map.stage_position} />

      {tiers.map((tier) => {
        const seats = map.seats
          .filter((seat) => seat.tier_id === tier.tier_id)
          .sort((a, b) => a.sort_index - b.sort_index)

        return (
          <section className="tk-tier" key={tier.tier_id}>
            <div className="tk-tier-head">
              <h3>{tier.name}</h3>
              <p className="tk-tier-meta">
                {formatPrice(tier.price_cents, map.currency)} ·{' '}
                {tier.seats_remaining} of {tier.seats_total} left
              </p>
            </div>

            <div className="tk-seat-scroll">
              <div
                className="tk-seat-grid"
                style={
                  {
                    '--tk-columns': map.columns || 10,
                  } as CSSProperties
                }
              >
                {seats.map((seat) => {
                  const isSelected = selectedSet.has(seat.code)
                  const unavailable = seat.status !== 'available'

                  return (
                    <button
                      aria-label={`Seat ${seat.label ?? seat.code}, ${seat.tier_name}, ${formatPrice(
                        seat.price_cents,
                        map.currency,
                      )}, ${isSelected ? 'selected' : statusLabels[seat.status]}`}
                      aria-pressed={isSelected}
                      className={`tk-seat${
                        isSelected ? ' is-selected' : ''
                      }${unavailable ? ` is-${seat.status}` : ''}`}
                      disabled={disabled || (unavailable && !isSelected)}
                      key={seat.code}
                      onClick={() => onToggle(seat.code)}
                      type="button"
                    >
                      {seat.label ?? seat.code}
                    </button>
                  )
                })}
              </div>
            </div>
          </section>
        )
      })}

      {map.stage_position === 'back' ? (
        <StageBar label={map.stage_label} position="back" />
      ) : null}

      <ul className="tk-legend" aria-label="Seat status legend">
        <li>
          <span className="tk-swatch tk-swatch--available" />
          Available
        </li>
        <li>
          <span className="tk-swatch tk-swatch--selected" />
          Selected
        </li>
        <li>
          <span className="tk-swatch tk-swatch--held" />
          On hold
        </li>
        <li>
          <span className="tk-swatch tk-swatch--sold" />
          Sold
        </li>
      </ul>
    </div>
  )
}
