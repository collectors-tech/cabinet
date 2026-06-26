import { toast } from 'sonner'

type SubmittedDataHistoryOptions = {
  source_label?: string
  category?: string
}

function submittedDataSummary(data: unknown) {
  try {
    return JSON.stringify(data, null, 2)
  } catch {
    return 'Submitted data could not be serialized for display.'
  }
}

export function showSubmittedData(
  data: unknown,
  title: string = 'You submitted the following values:',
  history: SubmittedDataHistoryOptions = {}
) {
  const summary = submittedDataSummary(data)
  toast.message(title, {
    history: {
      title,
      summary,
      source_label: history.source_label ?? 'Submitted Data',
      category: history.category ?? 'system',
    },
    description: (
      // w-[340px]
      <pre className='mt-2 w-full overflow-x-auto rounded-md bg-slate-950 p-4'>
        <code className='text-white'>{summary}</code>
      </pre>
    ),
  } as never)
}
