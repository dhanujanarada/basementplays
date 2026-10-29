import Link from 'next/link'
import type { ComponentProps } from 'react'

type ButtonProps = ComponentProps<typeof Link> & {
  variant?: 'primary' | 'secondary' | 'ghost'
}

const variants = {
  primary:
    'bg-[var(--accent)] text-[var(--bg)] hover:opacity-90',
  secondary:
    'border border-[var(--border)] text-[var(--text)] hover:border-[var(--accent)] hover:text-[var(--accent)]',
  ghost: 'text-[var(--muted)] hover:text-[var(--accent)]',
}

export function Button({
  variant = 'primary',
  className = '',
  ...props
}: ButtonProps) {
  return (
    <Link
      className={`inline-flex items-center justify-center gap-2 rounded-full px-5 py-3 text-sm font-medium tracking-wide transition ${variants[variant]} ${className}`}
      {...props}
    />
  )
}
