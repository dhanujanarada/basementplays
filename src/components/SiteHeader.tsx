import Link from 'next/link'

import { Container } from '@/components/ui/Container'

const navigation = [
  { href: '/', label: 'Home' },
  { href: '/events', label: 'Events' },
]

export function SiteHeader() {
  return (
    <header className="sticky top-0 z-50 border-b border-[var(--border)] bg-[var(--bg)]/90 backdrop-blur">
      <Container className="flex h-20 items-center justify-between gap-6">
        <Link
          className="font-display text-xl tracking-wide text-[var(--text)] transition hover:text-[var(--accent)]"
          href="/"
        >
          Basement Plays
        </Link>

        <nav aria-label="Primary navigation" className="flex items-center gap-6">
          {navigation.map((item) => (
            <Link
              className="text-xs font-medium uppercase tracking-[0.18em] text-[var(--muted)] transition hover:text-[var(--accent)]"
              href={item.href}
              key={item.href}
            >
              {item.label}
            </Link>
          ))}
          <Link
            className="hidden text-xs font-medium uppercase tracking-[0.18em] text-[var(--muted)] transition hover:text-[var(--accent)] sm:inline"
            href="/admin"
          >
            Admin
          </Link>
        </nav>
      </Container>
    </header>
  )
}
