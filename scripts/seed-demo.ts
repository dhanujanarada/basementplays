import 'dotenv/config'

import fs from 'node:fs'

import type { Payload } from 'payload'
import { getPayload } from 'payload'

import config from '../src/payload.config'
import type { Production } from '../src/payload-types'
import {
  renderPlaceholder,
  type PlaceholderImage,
} from './seed-images'

type ShowStatus = 'open' | 'few-left' | 'sold-out' | 'cancelled'

type ShowSeed = {
  startsAt: string
  status: ShowStatus
  price: string
  seatsNote?: string
  ticketUrl?: string
}

type RunSeed = {
  label: string
  shows: ShowSeed[]
  venue: string
}

type ProductionSeed = {
  backstage: (PlaceholderImage & { caption: string })[]
  cast: { name: string; role: string }[]
  crew: { name: string; role: string }[]
  gallery: PlaceholderImage[]
  hero: PlaceholderImage
  poster: PlaceholderImage
  runs: RunSeed[]
  slug: string
  synopsis: string[]
  tagline: string
  title: string
}

type Synopsis = NonNullable<Production['synopsis']>

const databaseUrl = process.env.DATABASE_URL || 'file:./payload.db'

if (databaseUrl.startsWith('file:')) {
  const databasePath = databaseUrl.slice('file:'.length)

  if (!fs.existsSync(databasePath)) {
    throw new Error(
      'payload.db does not exist yet. Run `npm run dev` once to create the schema, then run `npm run seed:demo`.',
    )
  }
}

const productions: ProductionSeed[] = [
  {
    title: 'The Hollow Crown',
    slug: 'the-hollow-crown',
    tagline: 'A gothic tragedy of inheritance and rot.',
    synopsis: [
      'The old king is dead, and the crown will not stay quiet. When a fractured family gathers to divide a rotting estate, a stranger arrives with a claim no one can prove—and a secret everyone already suspects.',
      'Part ghost story, part inheritance drama, The Hollow Crown is a slow-burning gothic tragedy about the stories we inherit and the ones we bury.',
    ],
    poster: {
      accent: '#b5453c',
      alt: 'The Hollow Crown poster with a crown motif and dark red lettering',
      background: '#08070b',
      eyebrow: 'A Basement Plays production',
      foreground: '#e9dfd2',
      key: 'hollow-crown-poster',
      kind: 'poster',
      motif: 'crown',
      subtitle: 'A gothic tragedy of inheritance and rot',
      title: 'THE HOLLOW|CROWN',
    },
    hero: {
      accent: '#b5453c',
      alt: 'The Hollow Crown hero image with a crown and stage light',
      background: '#08070b',
      eyebrow: 'Basement Plays presents',
      foreground: '#e9dfd2',
      key: 'hollow-crown-hero',
      kind: 'hero',
      motif: 'crown',
      subtitle: 'A gothic tragedy of inheritance and rot',
      title: 'THE HOLLOW|CROWN',
    },
    cast: [
      { name: 'Anushka Perera', role: 'Queen Isolde' },
      { name: 'Ravi Fernando', role: 'The Heir' },
      { name: 'Maya Jayasinghe', role: 'The Fool' },
      { name: 'Dilan Silva', role: 'The Regent' },
    ],
    crew: [
      { name: 'Nadeesha Wickramasinghe', role: 'Director' },
      { name: 'Kasun Mendis', role: 'Set & Lighting' },
      { name: 'Ishara de Silva', role: 'Costumes' },
      { name: 'Tharindu Gunasekara', role: 'Stage Manager' },
    ],
    gallery: [
      {
        accent: '#b5453c',
        alt: 'The cast of The Hollow Crown gathered around a long banquet table',
        background: '#0d0a0e',
        eyebrow: 'Production still',
        foreground: '#efe6d8',
        key: 'hollow-crown-gallery-1',
        kind: 'gallery',
        motif: 'stage',
        subtitle: 'Act I, scene ii',
        title: 'THE BANQUET|SCENE',
      },
      {
        accent: '#b5453c',
        alt: 'The Heir confronts the Regent under a cold spotlight',
        background: '#100c11',
        eyebrow: 'Production still',
        foreground: '#efe6d8',
        key: 'hollow-crown-gallery-2',
        kind: 'gallery',
        motif: 'stage',
        subtitle: 'Act II, scene i',
        title: 'THE HEIR|RETURNS',
      },
      {
        accent: '#b5453c',
        alt: 'Queen Isolde alone in the crown room with a single candle',
        background: '#0a080c',
        eyebrow: 'Production still',
        foreground: '#efe6d8',
        key: 'hollow-crown-gallery-3',
        kind: 'gallery',
        motif: 'stage',
        subtitle: 'Act II, scene iii',
        title: 'THE CROWN|ROOM',
      },
      {
        accent: '#b5453c',
        alt: 'The family raises a final toast beneath a cracked portrait',
        background: '#120d10',
        eyebrow: 'Production still',
        foreground: '#efe6d8',
        key: 'hollow-crown-gallery-4',
        kind: 'gallery',
        motif: 'stage',
        subtitle: 'Act III, scene ii',
        title: 'THE LAST|TOAST',
      },
    ],
    backstage: [
      {
        accent: '#c9a227',
        alt: 'Backstage at The Hollow Crown: a crown prop and costume rack',
        background: '#1a1110',
        caption: 'The crown prop gets its final fitting before opening night.',
        eyebrow: 'Behind the scenes',
        foreground: '#f1e3d2',
        key: 'hollow-crown-backstage-1',
        kind: 'backstage',
        motif: 'candid',
        subtitle: 'Backstage — costume fitting',
        title: 'CROWN|PROP',
      },
      {
        accent: '#c9a227',
        alt: 'Backstage at The Hollow Crown: the banquet table under construction',
        background: '#17100e',
        caption: 'Building the banquet table out of reclaimed timber.',
        eyebrow: 'Behind the scenes',
        foreground: '#f1e3d2',
        key: 'hollow-crown-backstage-2',
        kind: 'backstage',
        motif: 'candid',
        subtitle: 'Backstage — set build',
        title: 'SET|BUILD',
      },
      {
        accent: '#c9a227',
        alt: 'Backstage at The Hollow Crown: a stage manager calling cues',
        background: '#1c1210',
        caption: 'The stage manager calls the final cue before the dress rehearsal.',
        eyebrow: 'Behind the scenes',
        foreground: '#f1e3d2',
        key: 'hollow-crown-backstage-3',
        kind: 'backstage',
        motif: 'candid',
        subtitle: 'Backstage — cue call',
        title: 'CUE|CALL',
      },
    ],
    runs: [
      {
        label: '2026 Season',
        venue: 'The Basement Theatre, Colombo',
        shows: [
          show(
            '2026-11-06T19:30:00+05:30',
            'open',
            'https://example.com/tickets/the-hollow-crown',
          ),
          show('2026-11-07T19:30:00+05:30', 'sold-out'),
          show(
            '2026-11-08T15:00:00+05:30',
            'open',
            'https://example.com/tickets/the-hollow-crown',
          ),
          show(
            '2026-11-12T19:30:00+05:30',
            'open',
            'https://example.com/tickets/the-hollow-crown',
          ),
          show('2026-11-13T19:30:00+05:30', 'cancelled'),
        ],
      },
      {
        label: '2025 Season',
        venue: 'The Basement Theatre, Colombo',
        shows: [
          show('2025-11-14T19:30:00+05:30', 'sold-out'),
          show('2025-11-15T19:30:00+05:30', 'open'),
          show('2025-11-16T15:00:00+05:30', 'open'),
        ],
      },
    ],
  },
  {
    title: 'The Sunshine Committee',
    slug: 'the-sunshine-committee',
    tagline: 'A small-town comedy about terrible ideas and excellent snacks.',
    synopsis: [
      'The Sunshine Committee has one job: make the town square feel cheerful for the annual harvest fair. Unfortunately, the committee has no budget, one broken gazebo, and a deeply personal feud about lemon bars.',
      'A warm, chaotic comedy about small-town ambition, terrible craft projects, and the people who show up when everything goes sideways.',
    ],
    poster: {
      accent: '#e46f3d',
      alt: 'The Sunshine Committee poster with a bright sun motif',
      background: '#f7e6ae',
      eyebrow: 'A Basement Plays comedy',
      foreground: '#1e2a24',
      key: 'sunshine-committee-poster',
      kind: 'poster',
      motif: 'sun',
      subtitle: 'Terrible ideas. Excellent snacks.',
      title: 'THE SUNSHINE|COMMITTEE',
    },
    hero: {
      accent: '#e46f3d',
      alt: 'The Sunshine Committee hero image with a bright sun motif',
      background: '#f7e6ae',
      eyebrow: 'Basement Plays presents',
      foreground: '#1e2a24',
      key: 'sunshine-committee-hero',
      kind: 'hero',
      motif: 'sun',
      subtitle: 'Terrible ideas. Excellent snacks.',
      title: 'THE SUNSHINE|COMMITTEE',
    },
    cast: [
      { name: 'Priya Nair', role: 'Sunita, the mayor' },
      { name: 'Joshua de Alwis', role: 'Miles, the intern' },
      { name: 'Hannah Rodrigo', role: 'Bea, the baker' },
      { name: 'Sam Fonseka', role: 'Uncle Percy' },
    ],
    crew: [
      { name: 'Ayesha Fernando', role: 'Director' },
      { name: 'Rohan Perera', role: 'Musical Director' },
      { name: 'Nimali Costa', role: 'Choreography' },
      { name: 'Dinesh Weerasinghe', role: 'Stage Manager' },
    ],
    gallery: [
      {
        accent: '#e46f3d',
        alt: 'The Sunshine Committee cast gathered around a collapsing gazebo',
        background: '#f9edc4',
        eyebrow: 'Production still',
        foreground: '#1e2a24',
        key: 'sunshine-committee-gallery-1',
        kind: 'gallery',
        motif: 'stage',
        subtitle: 'Act I, scene i',
        title: 'THE GAZEBO|DISASTER',
      },
      {
        accent: '#e46f3d',
        alt: 'The committee holds an emergency meeting over lemon bars',
        background: '#f5e3a6',
        eyebrow: 'Production still',
        foreground: '#1e2a24',
        key: 'sunshine-committee-gallery-2',
        kind: 'gallery',
        motif: 'stage',
        subtitle: 'Act I, scene iii',
        title: 'LEMON BAR|SUMMIT',
      },
      {
        accent: '#e46f3d',
        alt: 'A craft table covered in ribbon, glue, and questionable decorations',
        background: '#fbf0cf',
        eyebrow: 'Production still',
        foreground: '#1e2a24',
        key: 'sunshine-committee-gallery-3',
        kind: 'gallery',
        motif: 'stage',
        subtitle: 'Act II, scene i',
        title: 'THE CRAFT|TABLE',
      },
      {
        accent: '#e46f3d',
        alt: 'The harvest fair finally opens under a string of paper lanterns',
        background: '#f4df9e',
        eyebrow: 'Production still',
        foreground: '#1e2a24',
        key: 'sunshine-committee-gallery-4',
        kind: 'gallery',
        motif: 'stage',
        subtitle: 'Act II, scene iii',
        title: 'HARVEST|FAIR',
      },
    ],
    backstage: [
      {
        accent: '#b5522c',
        alt: 'Backstage at The Sunshine Committee: the prop table with lemon bars',
        background: '#3a2417',
        caption:
          'Lemon bars, fake prize ribbons, and one very real tray of snacks.',
        eyebrow: 'Behind the scenes',
        foreground: '#ffe9bd',
        key: 'sunshine-committee-backstage-1',
        kind: 'backstage',
        motif: 'candid',
        subtitle: 'Backstage — prop table',
        title: 'PROP|TABLE',
      },
      {
        accent: '#b5522c',
        alt: 'Backstage at The Sunshine Committee: the quick-change corner',
        background: '#42291a',
        caption:
          'The quick-change corner, held together with velcro and hope.',
        eyebrow: 'Behind the scenes',
        foreground: '#ffe9bd',
        key: 'sunshine-committee-backstage-2',
        kind: 'backstage',
        motif: 'candid',
        subtitle: 'Backstage — quick change',
        title: 'COSTUME|SWAP',
      },
      {
        accent: '#b5522c',
        alt: 'Backstage at The Sunshine Committee: painting the gazebo set',
        background: '#372216',
        caption:
          'Painting the gazebo the exact shade of municipal optimism.',
        eyebrow: 'Behind the scenes',
        foreground: '#ffe9bd',
        key: 'sunshine-committee-backstage-3',
        kind: 'backstage',
        motif: 'candid',
        subtitle: 'Backstage — set paint',
        title: 'SET|PAINT',
      },
    ],
    runs: [
      {
        label: '2026 Season',
        venue: 'The Basement Theatre, Colombo',
        shows: [
          show(
            '2026-11-20T19:30:00+05:30',
            'open',
            'https://example.com/tickets/the-sunshine-committee',
          ),
          show(
            '2026-11-21T19:30:00+05:30',
            'open',
            'https://example.com/tickets/the-sunshine-committee',
          ),
          show('2026-11-22T15:00:00+05:30', 'sold-out'),
          show(
            '2026-11-27T19:30:00+05:30',
            'open',
            'https://example.com/tickets/the-sunshine-committee',
          ),
          show(
            '2026-11-28T19:30:00+05:30',
            'open',
            'https://example.com/tickets/the-sunshine-committee',
          ),
        ],
      },
    ],
  },
]

const payload = await getPayload({ config })

for (const productionSeed of productions) {
  await seedProduction(payload, productionSeed)
}

const productionCount = await payload.count({
  collection: 'productions',
  overrideAccess: true,
})
const runCount = await payload.count({
  collection: 'runs',
  overrideAccess: true,
})

console.log(
  `Demo seed complete: ${productionCount.totalDocs} productions, ${runCount.totalDocs} runs.`,
)

await payload.destroy()
process.exit(0)

async function seedProduction(
  payload: Payload,
  seed: ProductionSeed,
): Promise<void> {
  const existing = await payload.find({
    collection: 'productions',
    depth: 0,
    limit: 1,
    overrideAccess: true,
    where: {
      slug: {
        equals: seed.slug,
      },
    },
  })

  let production = existing.docs[0]

  if (production) {
    console.log(`Production already exists: ${seed.title}`)
  } else {
    console.log(`Creating production: ${seed.title}`)

    const poster = await uploadImage(payload, seed.poster)
    const hero = await uploadImage(payload, seed.hero)
    const gallery: number[] = []

    for (const image of seed.gallery) {
      const media = await uploadImage(payload, image)
      gallery.push(media.id)
    }

    const backstage: { caption: string; image: number }[] = []

    for (const image of seed.backstage) {
      const media = await uploadImage(payload, image)
      backstage.push({
        caption: image.caption,
        image: media.id,
      })
    }

    production = await payload.create({
      collection: 'productions',
      data: {
        backstage,
        cast: seed.cast,
        crew: seed.crew,
        gallery,
        heroImage: hero.id,
        poster: poster.id,
        slug: seed.slug,
        synopsis: richText(seed.synopsis),
        tagline: seed.tagline,
        title: seed.title,
      },
      overrideAccess: true,
    })
  }

  for (const run of seed.runs) {
    await seedRun(payload, production.id, run)
  }
}

async function seedRun(
  payload: Payload,
  productionId: number,
  run: RunSeed,
): Promise<void> {
  const existing = await payload.find({
    collection: 'runs',
    depth: 0,
    limit: 1,
    overrideAccess: true,
    where: {
      and: [
        {
          production: {
            equals: productionId,
          },
        },
        {
          label: {
            equals: run.label,
          },
        },
      ],
    },
  })

  if (existing.docs.length > 0) {
    console.log(`Run already exists: ${run.label}`)
    return
  }

  await payload.create({
    collection: 'runs',
    data: {
      label: run.label,
      production: productionId,
      shows: run.shows,
      venue: run.venue,
    },
    overrideAccess: true,
  })

  console.log(`Created run: ${run.label}`)
}

async function uploadImage(payload: Payload, image: PlaceholderImage) {
  const filePath = await renderPlaceholder(image)

  return payload.create({
    collection: 'media',
    data: {
      alt: image.alt,
    },
    filePath,
    overrideAccess: true,
  })
}

function show(
  startsAt: string,
  status: ShowStatus,
  ticketUrl?: string,
  price = '',
  seatsNote?: string,
): ShowSeed {
  return {
    startsAt: new Date(startsAt).toISOString(),
    status,
    price,
    ...(seatsNote ? { seatsNote } : {}),
    ...(ticketUrl ? { ticketUrl } : {}),
  }
}

function richText(paragraphs: string[]): Synopsis {
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
