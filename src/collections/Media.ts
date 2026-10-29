import type { CollectionConfig } from 'payload'

import { anyone, isStaff } from './access'

export const Media: CollectionConfig = {
  slug: 'media',
  admin: {
    group: 'Content',
    useAsTitle: 'alt',
  },
  access: {
    read: anyone,
    create: isStaff,
    update: isStaff,
    delete: isStaff,
  },
  fields: [
    {
      name: 'alt',
      type: 'text',
      required: true,
    },
  ],
  upload: {
    staticDir: 'media',
    adminThumbnail: 'thumbnail',
    focalPoint: true,
    imageSizes: [
      {
        name: 'thumbnail',
        width: 400,
      },
      {
        name: 'card',
        width: 800,
      },
      {
        name: 'full',
        width: 1920,
      },
    ],
    mimeTypes: ['image/*'],
  },
}
