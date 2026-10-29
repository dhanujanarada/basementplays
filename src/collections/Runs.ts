import type { CollectionConfig } from 'payload'

import { anyone, isStaff } from './access'

export const Runs: CollectionConfig = {
  slug: 'runs',
  admin: {
    group: 'Content',
    useAsTitle: 'label',
    defaultColumns: ['label', 'production', 'updatedAt'],
  },
  access: {
    read: anyone,
    create: isStaff,
    update: isStaff,
    delete: isStaff,
  },
  fields: [
    {
      name: 'production',
      type: 'relationship',
      relationTo: 'productions',
      required: true,
    },
    {
      name: 'label',
      type: 'text',
      required: true,
    },
    {
      name: 'venue',
      type: 'text',
    },
    {
      name: 'shows',
      type: 'array',
      fields: [
        {
          name: 'startsAt',
          type: 'date',
          required: true,
          admin: {
            date: {
              pickerAppearance: 'dayAndTime',
            },
          },
        },
        {
          name: 'status',
          type: 'select',
          required: true,
          defaultValue: 'open',
          options: [
            {
              label: 'Open',
              value: 'open',
            },
            {
              label: 'Few left',
              value: 'few-left',
            },
            {
              label: 'Sold out',
              value: 'sold-out',
            },
            {
              label: 'Cancelled',
              value: 'cancelled',
            },
          ],
        },
        {
          name: 'price',
          type: 'text',
          required: true,
          defaultValue: '',
          admin: {
            description: 'e.g. Rs. 2500',
          },
        },
        {
          name: 'seatsNote',
          type: 'text',
          admin: {
            description: 'Optional. e.g. Less than 20 seats left',
          },
        },
        {
          name: 'ticketUrl',
          type: 'text',
        },
      ],
    },
  ],
}
