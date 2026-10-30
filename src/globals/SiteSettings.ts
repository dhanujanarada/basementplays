import type { GlobalConfig } from 'payload'

import { anyone, isStaff } from '../collections/access'

export const SiteSettings: GlobalConfig = {
  slug: 'site-settings',
  label: 'Site Settings',
  admin: {
    group: 'Content',
  },
  access: {
    read: anyone,
    update: isStaff,
  },
  fields: [
    {
      name: 'groupName',
      type: 'text',
      required: true,
    },
    {
      name: 'tagline',
      type: 'text',
    },
    {
      name: 'aboutText',
      type: 'richText',
    },
    {
      name: 'aboutImage',
      type: 'upload',
      relationTo: 'media',
    },
    {
      name: 'email',
      type: 'email',
    },
    {
      name: 'phone',
      type: 'text',
    },
    {
      name: 'whatsappNumber',
      type: 'text',
      admin: {
        description: 'International format, digits only, e.g. 94764508448',
      },
    },
    {
      name: 'locationText',
      type: 'text',
    },
    {
      name: 'mapsUrl',
      type: 'text',
    },
    {
      name: 'socialLinks',
      type: 'array',
      fields: [
        {
          name: 'platform',
          type: 'select',
          required: true,
          options: [
            {
              label: 'Instagram',
              value: 'instagram',
            },
            {
              label: 'Facebook',
              value: 'facebook',
            },
            {
              label: 'YouTube',
              value: 'youtube',
            },
            {
              label: 'TikTok',
              value: 'tiktok',
            },
            {
              label: 'X',
              value: 'x',
            },
            {
              label: 'Other',
              value: 'other',
            },
          ],
        },
        {
          name: 'url',
          type: 'text',
          required: true,
        },
      ],
    },
  ],
}
