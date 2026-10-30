import type { CollectionConfig } from 'payload'

import { canCreateUser, isAdmin } from './access'

export const Users: CollectionConfig = {
  slug: 'users',
  admin: {
    group: 'System',
    useAsTitle: 'email',
  },
  auth: {
    maxLoginAttempts: 5,
    lockTime: 10 * 60 * 1000,
  },
  access: {
    create: canCreateUser,
    read: isAdmin,
    update: isAdmin,
    delete: isAdmin,
    unlock: isAdmin,
  },
  fields: [
    {
      name: 'role',
      type: 'select',
      required: true,
      defaultValue: 'member',
      options: [
        {
          label: 'Admin',
          value: 'admin',
        },
        {
          label: 'Member',
          value: 'member',
        },
      ],
      saveToJWT: true,
    },
  ],
}
