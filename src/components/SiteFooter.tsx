import Link from 'next/link'

import { Container } from '@/components/ui/Container'

export function SiteFooter() {
  return (
    <footer className="border-t border-[var(--border)] py-10">
      <Container className="flex flex-col gap-4 text-sm text-[var(--muted)] sm:flex-row sm:items-center sm:justify-between">
        <p className="font-display text-base text-[var(--text)]">Basement Plays</p>
        <p>Independent theatre, made in the basement.</p>
        <Link className="transition hover:text-[var(--accent)]" href="/admin">
          Admin
        </Link>
      </Container>
    </footer>
  )
}
