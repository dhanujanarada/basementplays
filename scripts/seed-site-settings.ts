import 'dotenv/config'

import fs from 'node:fs'

import { getPayload } from 'payload'

import config from '../src/payload.config'
import type { SiteSetting } from '../src/payload-types'

type AboutText = NonNullable<SiteSetting['aboutText']>

const databaseUrl = process.env.DATABASE_URL || 'file:./payload.db'

if (databaseUrl.startsWith('file:')) {
  const databasePath = databaseUrl.slice('file:'.length)

  if (!fs.existsSync(databasePath)) {
    throw new Error(
      'payload.db does not exist yet. Run `npm run dev` once to create the schema, then run `npm run seed:site-settings`.',
    )
  }
}

const mapsQuery = encodeURIComponent('Kandy, Sri Lanka')
const socialLinks = [
  { platform: 'instagram' as const, url: 'https://instagram.com/example' },
  { platform: 'facebook' as const, url: 'https://facebook.com/example' },
]

const payload = await getPayload({ config })

const existing = await payload.findGlobal({
  slug: 'site-settings',
  depth: 0,
  overrideAccess: true,
})

if (existing.tagline && existing.aboutText) {
  console.log('Site Settings already seeded. Skipping content creation.')
} else {
  console.log('Seeding Site Settings')

  await payload.updateGlobal({
    slug: 'site-settings',
    data: {
      aboutText: richText([
        'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.',
        'Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.',
        'Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.',
      ]),
      email: 'hello@example.com',
      groupName: 'Basement',
      locationText: 'Kandy, Sri Lanka',
      mapsUrl: `https://www.google.com/maps/search/?api=1&query=${mapsQuery}`,
      socialLinks,
      tagline: 'Lorem ipsum dolor sit amet, consectetur adipiscing elit.',
      whatsappNumber: '94764508448',
    },
    overrideAccess: true,
  })

  console.log('Seeded Site Settings')
}

console.log('')
console.log('!'.repeat(78))
console.log('!!  SITE SETTINGS PLACEHOLDERS — REPLACE THESE IN BACKSTAGE             !!')
console.log('!!  groupName:      Basement')
console.log('!!  tagline:        Lorem ipsum dolor sit amet, consectetur adipiscing elit.')
console.log('!!  aboutText:      3 lorem ipsum paragraphs')
console.log('!!  email:          hello@example.com')
console.log('!!  whatsappNumber: 94764508448')
console.log('!!  locationText:   Kandy, Sri Lanka')
console.log('!!  mapsUrl:        Google Maps search for "Kandy, Sri Lanka"')
console.log('!!  socialLinks:    https://instagram.com/example, https://facebook.com/example')
console.log('!!  aboutImage:     not set — add a real image')
console.log(
  '!!  Backstage:      http://localhost:3000/backstage/globals/site-settings',
)
console.log('!'.repeat(78))
console.log('')

process.exit(0)

function richText(paragraphs: string[]): AboutText {
  return {
    root: {
      children: paragraphs.map((text) => ({
        children: [
          {
            detail: 0,
            format: 0,
            mode: 'normal',
            style: '',
            text,
            type: 'text',
            version: 1,
          },
        ],
        direction: 'ltr',
        format: '',
        indent: 0,
        type: 'paragraph',
        version: 1,
      })),
      direction: 'ltr',
      format: '',
      indent: 0,
      type: 'root',
      version: 1,
    },
  }
}
