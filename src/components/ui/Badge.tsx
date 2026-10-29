import type { ReactNode } from 'react'

type BadgeProps = {
  children: ReactNode
  tone?: 'default' | 'accent' | 'muted'
}

const tones = {
  default: 'border-[var(--border)] text-[var(--text)]',
  accent: 'border-[var(--accent)] text-[var(--accent)]',
  muted: 'border-[var(--border)] text-[var(--muted)]',
}

export function Badge({ children, tone = 'default' }: BadgeProps) {
  return (
    <span
      className={`inline-flex items-center rounded-full border px-3 py-1 text-[11px] font-medium uppercase tracking-[0.18em] ${tones[tone]}`}
    >
      {children}
    </span>
  )
}
