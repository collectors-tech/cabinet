import { Collection } from '@/features/collection'

export function Inventory() {
  return (
    <Collection
      title='Inventory'
      description='Manage collection items, grading status, packaging, and media.'
      routePath='/_authenticated/inventory/'
    />
  )
}
