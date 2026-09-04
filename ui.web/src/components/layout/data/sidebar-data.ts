import { Database } from 'lucide-react'
import { buildSidebarNavigationGroups } from '@/lib/route-navigation'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  user: {
    name: 'Local Admin',
    email: 'admin@local',
    avatar: '/avatars/01.png',
  },
  teams: [
    {
      name: 'Local Workspace',
      logo: Database,
      plan: 'Primary DB',
    },
  ],
  navGroups: buildSidebarNavigationGroups(),
}
