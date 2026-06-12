import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { HelpCenter } from '@/features/help-center'

const helpCenterSearchSchema = z.object({
  article: z.string().optional().catch(''),
  category: z.string().optional().catch(''),
  q: z.string().optional().catch(''),
})

export const Route = createFileRoute('/_authenticated/help-center/')({
  validateSearch: helpCenterSearchSchema,
  component: HelpCenter,
})
