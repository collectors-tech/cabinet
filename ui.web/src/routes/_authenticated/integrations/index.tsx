import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { Integrations } from '@/features/integrations'

const integrationsSearchSchema = z.object({
  type: z
    .enum(['all', 'connected', 'notConnected'])
    .optional()
    .catch(undefined),
  filter: z.string().optional().catch(''),
  sort: z.enum(['asc', 'desc']).optional().catch(undefined),
  view: z.enum(['rows', 'cards']).optional().catch(undefined),
  provider: z.string().optional().catch(undefined),
})

export const Route = createFileRoute('/_authenticated/integrations/')({
  validateSearch: integrationsSearchSchema,
  component: Integrations,
})
