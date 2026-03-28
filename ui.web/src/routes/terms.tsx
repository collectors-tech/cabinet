import { createFileRoute } from '@tanstack/react-router'
import { TermsOfService } from '@/features/auth/terms-of-service'

export const Route = createFileRoute('/terms')({
  component: TermsOfService,
})
