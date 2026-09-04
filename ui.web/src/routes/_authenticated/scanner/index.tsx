import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { Scanner } from '@/features/scanner'

const scannerSearchSchema = z.object({
  barcode: z.union([z.string(), z.number()]).optional().catch(''),
})

export const Route = createFileRoute('/_authenticated/scanner/')({
  validateSearch: scannerSearchSchema,
  component: Scanner,
})
