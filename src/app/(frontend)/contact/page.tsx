import { SocialLinks } from '@/components/SocialLinks'
import { Button } from '@/components/ui/Button'
import { Container } from '@/components/ui/Container'
import { getSiteSettings } from '@/lib/payload'

export const dynamic = 'force-dynamic'

export const metadata = {
  title: 'Contact',
}

export default async function ContactPage() {
  const settings = await getSiteSettings()
  const whatsappDigits = settings.whatsappNumber?.replace(/\D/g, '') ?? ''
  const whatsappHref = whatsappDigits
    ? `https://wa.me/${whatsappDigits}`
    : null
  const socialLinks = settings.socialLinks ?? []

  return (
    <Container className="py-16 sm:py-24">
      <p className="text-xs font-medium uppercase tracking-[0.22em] text-[var(--accent)]">
        Contact
      </p>
      <h1 className="font-display mt-4 text-5xl tracking-tight text-[var(--text)] sm:text-7xl">
        Get in touch
      </h1>
      {settings.tagline ? (
        <p className="mt-5 max-w-2xl text-base leading-7 text-[var(--muted)] sm:text-lg">
          {settings.tagline}
        </p>
      ) : null}

      <div className="mt-12 grid gap-6 md:grid-cols-2">
        <section className="rounded-sm border border-[var(--border)] bg-[var(--surface)]/40 p-6">
          <h2 className="font-display text-2xl text-[var(--text)]">Email</h2>
          {settings.email ? (
            <p className="mt-4 text-sm leading-7 text-[var(--muted)]">
              <a
                className="text-[var(--accent)] underline underline-offset-4"
                href={`mailto:${settings.email}`}
              >
                {settings.email}
              </a>
            </p>
          ) : (
            <p className="mt-4 text-sm leading-7 text-[var(--muted)]">
              Email details are coming soon.
            </p>
          )}
          {settings.phone ? (
            <p className="mt-2 text-sm leading-7 text-[var(--muted)]">
              Phone:{' '}
              <a
                className="text-[var(--accent)] underline underline-offset-4"
                href={`tel:${settings.phone.replace(/\s+/g, '')}`}
              >
                {settings.phone}
              </a>
            </p>
          ) : null}
        </section>

        <section className="rounded-sm border border-[var(--border)] bg-[var(--surface)]/40 p-6">
          <h2 className="font-display text-2xl text-[var(--text)]">WhatsApp</h2>
          <p className="mt-4 text-sm leading-7 text-[var(--muted)]">
            The quickest way to reach the company.
          </p>
          {whatsappHref ? (
            <Button
              className="mt-5"
              href={whatsappHref}
              rel="noreferrer"
              target="_blank"
            >
              Message us on WhatsApp
            </Button>
          ) : (
            <p className="mt-4 text-sm leading-7 text-[var(--muted)]">
              WhatsApp details are coming soon.
            </p>
          )}
        </section>

        <section className="rounded-sm border border-[var(--border)] bg-[var(--surface)]/40 p-6">
          <h2 className="font-display text-2xl text-[var(--text)]">Find us</h2>
          {settings.locationText ? (
            <p className="mt-4 text-sm leading-7 text-[var(--muted)]">
              {settings.locationText}
            </p>
          ) : (
            <p className="mt-4 text-sm leading-7 text-[var(--muted)]">
              Location details are coming soon.
            </p>
          )}
          {settings.mapsUrl ? (
            <Button
              className="mt-5"
              href={settings.mapsUrl}
              rel="noreferrer"
              target="_blank"
              variant="secondary"
            >
              Open in Google Maps
            </Button>
          ) : null}
        </section>

        <section className="rounded-sm border border-[var(--border)] bg-[var(--surface)]/40 p-6">
          <h2 className="font-display text-2xl text-[var(--text)]">Follow</h2>
          {socialLinks.length > 0 ? (
            <SocialLinks
              className="mt-4 flex flex-wrap gap-x-5 gap-y-2"
              linkClassName="text-sm text-[var(--accent)] underline decoration-[var(--accent)]/40 underline-offset-4 transition hover:text-[var(--text)]"
              links={socialLinks}
            />
          ) : (
            <p className="mt-4 text-sm leading-7 text-[var(--muted)]">
              Social links are coming soon.
            </p>
          )}
        </section>
      </div>
    </Container>
  )
}
