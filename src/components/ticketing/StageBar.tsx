import type { StagePosition } from '@/lib/ticketing/types'

type StageBarProps = {
  label?: string
  position: StagePosition
}

export function StageBar({ label = 'Stage', position }: StageBarProps) {
  if (position === 'none') {
    return null
  }

  return (
    <p className={`tk-stage${position === 'back' ? ' tk-stage--back' : ''}`}>
      {label || 'Stage'}
    </p>
  )
}
