import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { showSubmittedData } from '@/lib/show-submitted-data'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { type WishlistEntryDraft } from './tasks-mutate-drawer'

const formSchema = z.object({
  file: z
    .instanceof(FileList)
    .refine((files) => files.length > 0, {
      message: 'Please upload a file',
    })
    .refine(
      (files) => ['text/csv', 'application/vnd.ms-excel', ''].includes(files?.[0]?.type),
      'Please upload csv format.'
    ),
})

type TaskImportDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  routePath: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
  onWishlistImport?: (entries: WishlistEntryDraft[]) => Promise<void>
  isLoading?: boolean
}

function parseCsvLine(line: string) {
  const out: string[] = []
  let current = ''
  let inQuotes = false
  for (let i = 0; i < line.length; i += 1) {
    const char = line[i]
    if (char === '"') {
      if (inQuotes && line[i + 1] === '"') {
        current += '"'
        i += 1
      } else {
        inQuotes = !inQuotes
      }
      continue
    }
    if (char === ',' && !inQuotes) {
      out.push(current.trim())
      current = ''
      continue
    }
    current += char
  }
  out.push(current.trim())
  return out
}

function parseWishlistImportCsv(text: string): WishlistEntryDraft[] {
  const lines = text
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)

  if (lines.length < 2) {
    throw new Error('Import file must include a header row and at least one entry.')
  }

  const headers = parseCsvLine(lines[0]).map((header) =>
    header.trim().toLowerCase()
  )

  const columnIndex = (name: string) => headers.indexOf(name)
  const titleIndex = columnIndex('title')
  if (titleIndex < 0) {
    throw new Error('Import file must include a title column.')
  }

  return lines.slice(1).map((line, index) => {
    const cells = parseCsvLine(line)
    const title = cells[titleIndex]?.trim() ?? ''
    if (!title) {
      throw new Error(`Row ${index + 2} is missing a title.`)
    }
    return {
      title,
      partNumber: cells[columnIndex('part_number')]?.trim() ?? '',
      category: cells[columnIndex('category')]?.trim() ?? '',
      itemType: cells[columnIndex('item_type')]?.trim() ?? '',
      packagingGradeType:
        cells[columnIndex('packaging_grade_type')]?.trim() ?? '',
      condition: cells[columnIndex('condition')]?.trim() ?? '',
      priority: cells[columnIndex('priority')]?.trim() || 'medium',
      notes: cells[columnIndex('notes')]?.trim() ?? '',
      targetPrice: cells[columnIndex('target_price')]?.trim() ?? '',
      owned: false,
      delivered: false,
      pricePaid: '',
      purchaseUrl: '',
      purchaseDate: '',
      purchaseCondition: '',
      quantity: '0',
      neededQuantity: '1',
    }
  })
}

export function TasksImportDialog({
  open,
  onOpenChange,
  routePath,
  onWishlistImport,
  isLoading = false,
}: TaskImportDialogProps) {
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: { file: undefined },
  })

  const fileRef = form.register('file')

  const onSubmit = async () => {
    const file = form.getValues('file')

    if (file && file[0]) {
      if (isWishlistRoute && onWishlistImport) {
        const text = await file[0].text()
        const entries = parseWishlistImportCsv(text)
        await onWishlistImport(entries)
      } else {
        const fileDetails = {
          name: file[0].name,
          size: file[0].size,
          type: file[0].type,
        }
        showSubmittedData(fileDetails, 'You have imported the following file:')
      }
    }
    onOpenChange(false)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(val) => {
        onOpenChange(val)
        form.reset()
      }}
    >
      <DialogContent className='gap-2 sm:max-w-sm'>
        <DialogHeader className='text-start'>
          <DialogTitle>
            {isWishlistRoute ? 'Import Wishlist Entries' : 'Import Tasks'}
          </DialogTitle>
          <DialogDescription>
            {isWishlistRoute
              ? 'Import wishlist entries from CSV. Supported columns: title, part_number, category, priority, notes, target_price.'
              : 'Import tasks quickly from a CSV file.'}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form id='task-import-form' onSubmit={form.handleSubmit(onSubmit)}>
            <FormField
              control={form.control}
              name='file'
              render={() => (
                <FormItem className='my-2'>
                  <FormLabel>File</FormLabel>
                  <FormControl>
                    <Input type='file' {...fileRef} className='h-8 py-0' />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>
        <DialogFooter className='gap-2'>
          <DialogClose asChild>
            <Button variant='outline' disabled={isLoading}>
              Close
            </Button>
          </DialogClose>
          <Button type='submit' form='task-import-form' disabled={isLoading}>
            {isLoading ? 'Importing...' : 'Import'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
