import { createFileRoute } from '@tanstack/react-router'
import { Collections } from '@/features/collections'

export const Route = createFileRoute('/_authenticated/collections/')({
  component: Collections,
})
