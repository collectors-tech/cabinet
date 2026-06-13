import { createFileRoute } from '@tanstack/react-router'
import { useEffect } from 'react'
import { NotFoundError } from '@/features/errors/not-found-error'

export const Route = createFileRoute('/(errors)/404')({
  component: RouteComponent,
})

function RouteComponent() {
  useEffect(() => {
    document.title = 'Cabinet - Not Found'
  }, [])

  return <NotFoundError />
}
