import { z } from 'zod'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'

const searchSchema = z.object({
  redirect: z.string().optional(),
})

export const Route = createFileRoute('/(auth)/sign-out')({
  validateSearch: searchSchema,
  beforeLoad: ({ search }) => {
    useAuthStore.getState().auth.reset()
    throw redirect({
      to: '/sign-in',
      search: search.redirect ? { redirect: search.redirect } : {},
      replace: true,
    })
  },
})
