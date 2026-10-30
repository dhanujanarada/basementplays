import Link from 'next/link'

import { Container } from '@/components/ui/Container'
import { SocialLinks } from '@/components/SocialLinks'
import { getSiteSettings } from '@/lib/payload'

const quickLinks = [
  { href: '/', label: 'Home' },
  { href: '/events', label: 'Events' },
  { href: '/about', label: 'About' },
  { href: '/contact', label: 'Contact' },
]

export async function SiteFooter() {
  const settings = await getSiteSettings()
  const groupName = settings.groupName || 'Basement Plays'
  const year = new Date().getFullYear()
  const whatsappDigits = settings.whatsappNumber?.replace(/\D/g, '') ?? ''
  const whatsappHref = whatsappDigits
    ? `https://wa.me/${whatsappDigits}`
    : null
  const phoneHref = settings.phone
    ? `tel:${settings.phone.replace(/\s+/g, '')}`
    : null
  const socialLinks = settings.socialLinks ?? []
  const linkClass =
    'text-sm text-[var(--footer-text,#eee)] opacity-85 transition hover:text-[var(--footer-accent,#c9a227)] hover:opacity-100'

  return (
    <footer
      className="border-t bg-[var(--footer-bg,#111)] text-[var(--footer-text,#eee)]"
      style={{
        borderColor:
          'color-mix(in srgb, var(--footer-accent, #c9a227) 28%, transparent)',
      }}
    >
      <Container className="py-12 sm:py-16">
        <div className="grid gap-10 sm:grid-cols-2 lg:grid-cols-[1.5fr_0.8fr_1fr_0.8fr]">
          <div>
            <p className="font-display text-2xl leading-tight">{groupName}</p>
            {settings.tagline ? (
              <p className="mt-3 max-w-sm text-sm leading-6 opacity-80">
                {settings.tagline}
              </p>
            ) : null}
          </div>

          <nav aria-label="Footer navigation">
            <p className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--footer-accent,#c9a227)]">
              Explore
            </p>
            <ul className="mt-5 space-y-3">
              {quickLinks.map((link) => (
                <li key={link.href}>
                  <Link className={linkClass} href={link.href}>
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>

          <div>
            <p className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--footer-accent,#c9a227)]">
              Contact
            </p>
            <ul className="mt-5 space-y-3">
              {settings.email ? (
                <li>
                  <a className={linkClass} href={`mailto:${settings.email}`}>
                    {settings.email}
                  </a>
                </li>
              ) : null}
              {whatsappHref ? (
                <li>
                  <a
                    className={linkClass}
                    href={whatsappHref}
                    rel="noreferrer"
                    target="_blank"
                  >
                    WhatsApp
                  </a>
                </li>
              ) : null}
              {phoneHref ? (
                <li>
                  <a className={linkClass} href={phoneHref}>
                    {settings.phone}
                  </a>
                </li>
              ) : null}
              {settings.locationText ? (
                <li>
                  {settings.mapsUrl ? (
                    <a
                      className={linkClass}
                      href={settings.mapsUrl}
                      rel="noreferrer"
                      target="_blank"
                    >
                      {settings.locationText}
                    </a>
                  ) : (
                    <span className="text-sm opacity-85">
                      {settings.locationText}
                    </span>
                  )}
                </li>
              ) : null}
            </ul>
          </div>

          {socialLinks.length > 0 ? (
            <div>
              <p className="text-[10px] font-medium uppercase tracking-[0.22em] text-[var(--footer-accent,#c9a227)]">
                Follow
              </p>
              <SocialLinks
                className="mt-5 flex flex-wrap gap-x-5 gap-y-2"
                linkClassName={linkClass}
                links={socialLinks}
              />
            </div>
          ) : null}
        </div>

        <div
          className="mt-12 border-t pt-6 text-xs opacity-70"
          style={{
            borderColor:
              'color-mix(in srgb, var(--footer-accent, #c9a227) 22%, transparent)',
          }}
        >
          © {year} {groupName}. All rights reserved.
        </div>
      </Container>
    </footer>
  )
}
