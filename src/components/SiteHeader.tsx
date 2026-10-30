import Link from 'next/link'

import { Container } from '@/components/ui/Container'
import { getSiteSettings } from '@/lib/payload'

const navigation = [
  { href: '/', label: 'Home' },
  { href: '/events', label: 'Events' },
  { href: '/about', label: 'About' },
  { href: '/contact', label: 'Contact' },
]

export async function SiteHeader() {
  const settings = await getSiteSettings()
  const groupName = settings.groupName || 'Basement Plays'

  return (
    <header className="sticky top-0 z-50 border-b border-[var(--border)] bg-[var(--bg)]/90 backdrop-blur">
      <Container className="flex h-20 items-center justify-between gap-4 sm:gap-6">
        <Link
          className="font-display text-lg tracking-wide text-[var(--text)] transition hover:text-[var(--accent)] sm:text-xl"
          href="/"
        >
          {groupName}
        </Link>

        <nav
          aria-label="Primary navigation"
          className="flex items-center gap-4 sm:gap-6"
        >
          {navigation.map((item) => (
            <Link
              className="text-[10px] font-medium uppercase tracking-[0.18em] text-[var(--muted)] transition hover:text-[var(--accent)] sm:text-xs"
              href={item.href}
              key={item.href}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </Container>
    </header>
  )
}
