import { Tasks } from '@/features/tasks'

export function Inventory() {
  return (
    <Tasks
      title='Inventory'
      description='Manage collection items, grading status, packaging, and media.'
      routePath='/_authenticated/inventory/'
    />
  )
}
