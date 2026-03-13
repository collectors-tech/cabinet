import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/')({
  beforeLoad: () => {
    throw redirect({
      to: '/dashboard',
      replace: true,
    })
  },
})
