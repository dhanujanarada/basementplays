# 0007 — Site Settings feeds public chrome and fallbacks

- **Status:** accepted
- **Covers:** `src/globals/SiteSettings.ts`, `src/components/SiteHeader.tsx`, `src/components/SiteFooter.tsx`, `src/app/(frontend)/about/`, `src/app/(frontend)/contact/`
- **Last checked:** 8de3320
- **Implementation:** partial
- **Date:** 2026-11
- **Related:** [0023](0023-payment-modes.md), [0030](0030-access-control.md), `src/globals/SiteSettings.ts`

## Context

Contact details, navigation text, social links, the WhatsApp fallback, and bank details should be editable once, not per event. A contact form would need email delivery and spam protection.

## Decision

Use the Site Settings global for public identity and contact content. Use direct email, phone, WhatsApp, and maps links instead of a contact form. Keep bank transfer details there too. A Privacy Policy page is required before public booking because buyer names, phones, and emails are collected.

## Alternatives considered

- **A contact form** — needs email delivery and spam protection before it is trustworthy.

## Consequences and trade-offs accepted

One edit updates the header, footer, About, Contact, fallback links, and bank instructions. Privacy and Terms links are deferred, but the privacy page is a launch blocker for real bookings.

## Revisit when

Revisit when the site starts collecting real bookings, which makes the Privacy Policy page mandatory. Revisit the contact form only with a delivery and anti-spam plan.

## Related

[0023](0023-payment-modes.md), [0030](0030-access-control.md), `src/globals/SiteSettings.ts`
