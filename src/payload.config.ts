import { sqliteAdapter } from '@payloadcms/db-sqlite'
import { lexicalEditor } from '@payloadcms/richtext-lexical'
import path from 'path'
import { buildConfig } from 'payload'
import { fileURLToPath } from 'url'
import sharp from 'sharp'

import { Media } from './collections/Media'
import { Productions } from './collections/Productions'
import { Props } from './collections/Props'
import { Runs } from './collections/Runs'
import { Users } from './collections/Users'
import { SiteSettings } from './globals/SiteSettings'

const filename = fileURLToPath(import.meta.url)
const dirname = path.dirname(filename)

export default buildConfig({
  admin: {
    user: Users.slug,
    importMap: {
      baseDir: path.resolve(dirname),
    },
    meta: {
      titleSuffix: ' - Backstage',
    },
  },
  routes: {
    admin: '/backstage',
  },
  collections: [Users, Media, Productions, Runs, Props],
  globals: [SiteSettings],
  editor: lexicalEditor(),
  secret: process.env.PAYLOAD_SECRET || '',
  typescript: {
    outputFile: path.resolve(dirname, 'payload-types.ts'),
  },
  db: sqliteAdapter({
    client: {
      // Keep the database connection configurable per environment.
      url: process.env.DATABASE_URL || 'file:./payload.db',
    },
    push: process.env.PAYLOAD_SKIP_DB_PUSH !== 'true',
  }),
  sharp,
  plugins: [],
})
