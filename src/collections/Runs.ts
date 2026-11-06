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
          admin: {
            description:
              'Manual status. While ticketing is enabled, availability comes from the ticketing service; use this only to force sold out or cancelled.',
          },
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
          name: 'ticketingEnabled',
          type: 'checkbox',
          defaultValue: false,
          admin: {
            description:
              'Use the ticketing service for this show. Prices, status, and seats remaining then come from its seat plan; the price and seats note below are ignored.',
          },
        },
        {
          name: 'price',
          type: 'text',
          required: true,
          defaultValue: '',
          admin: {
            description:
              'e.g. Rs. 2500. Ignored while ticketing is enabled.',
          },
        },
        {
          name: 'seatsNote',
          type: 'text',
          admin: {
            description:
              'Optional. e.g. Less than 20 seats left. Ignored while ticketing is enabled.',
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
