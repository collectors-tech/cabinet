import { z } from 'zod'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { lockLocalServerSession } from '@/lib/cabinet-session'

const searchSchema = z.object({
  redirect: z.string().optional(),
})

export const Route = createFileRoute('/(auth)/sign-out')({
  validateSearch: searchSchema,
  beforeLoad: async ({ search }) => {
    let providerLogoutURL = ''
    await lockLocalServerSession()
    try {
      const response = await fetch('/api/auth/zitadel/logout', {
        method: 'POST',
        credentials: 'same-origin',
        headers: { Accept: 'application/json' },
      })
      if (response.ok) {
        const payload = (await response.json()) as {
          provider_logout_url?: string
        }
        providerLogoutURL = payload.provider_logout_url?.trim() || ''
      }
    } catch {
      // Local reset below still runs when provider logout is unavailable.
    }
    useAuthStore.getState().auth.reset()
    if (providerLogoutURL) {
      window.location.replace(providerLogoutURL)
      return
    }
    throw redirect({
      to: '/sign-in',
      search: search.redirect ? { redirect: search.redirect } : {},
      replace: true,
    })
  },
})
