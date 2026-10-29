import 'dotenv/config'

import fs from 'node:fs'

import { getPayload } from 'payload'

import config from '../src/payload.config'

const email = process.env.SEED_ADMIN_EMAIL
const password = process.env.SEED_ADMIN_PASSWORD
const databaseUrl = process.env.DATABASE_URL || 'file:./payload.db'

if (!email || !password) {
  throw new Error(
    'Set SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD before running this script.',
  )
}

if (databaseUrl.startsWith('file:')) {
  const databasePath = databaseUrl.slice('file:'.length)

  if (!fs.existsSync(databasePath)) {
    throw new Error(
      'payload.db does not exist yet. Run `npm run dev` once to create the schema, then run `npm run seed:admin`.',
    )
  }
}

const payload = await getPayload({ config })

const existing = await payload.find({
  collection: 'users',
  depth: 0,
  limit: 1,
  overrideAccess: true,
  where: {
    email: {
      equals: email,
    },
  },
})

if (existing.docs.length > 0) {
  console.log(`Admin user already exists: ${email}`)
  process.exit(0)
}

await payload.create({
  collection: 'users',
  data: {
    email,
    password,
    role: 'admin',
  },
  overrideAccess: true,
})

console.log(`Created admin user: ${email}`)
process.exit(0)
