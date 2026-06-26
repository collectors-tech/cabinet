import { AxiosError } from 'axios'
import { toast } from 'sonner'

function serverErrorToastHistory(title: string, summary: string) {
  return {
    history: {
      title,
      summary,
      source_label: 'Global server error',
      category: 'system',
    },
  } as never
}

export function handleServerError(error: unknown) {
  // eslint-disable-next-line no-console
  console.log(error)

  let errMsg = 'Something went wrong!'

  if (
    error &&
    typeof error === 'object' &&
    'status' in error &&
    Number(error.status) === 204
  ) {
    errMsg = 'Content not found.'
  }

  if (error instanceof AxiosError) {
    errMsg = error.response?.data.title
  }

  toast.error(
    errMsg,
    serverErrorToastHistory('Server error feedback', errMsg)
  )
}
