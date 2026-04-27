import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { showSubmittedData } from '@/lib/show-submitted-data'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { SelectDropdown } from '@/components/select-dropdown'
import { priorities } from '../data/data'
import { type Task } from '../data/schema'

export type WishlistEntryDraft = {
  title: string
  partNumber: string
  category: string
  priority: string
  notes: string
  targetPrice: string
}

type TaskMutateDrawerProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: Task
  routePath: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
  onWishlistSubmit?: (
    draft: WishlistEntryDraft,
    currentRow?: Task
  ) => Promise<void>
  isLoading?: boolean
  canNavigatePrevious?: boolean
  canNavigateNext?: boolean
  onNavigatePrevious?: () => void
  onNavigateNext?: () => void
}

const taskFormSchema = z.object({
  title: z.string().min(1, 'Title is required.'),
  status: z.string().min(1, 'Please select a status.'),
  label: z.string().min(1, 'Please select a label.'),
  priority: z.string().min(1, 'Please choose a priority.'),
})

const wishlistFormSchema = z.object({
  title: z.string().trim().min(1, 'Title is required.'),
  partNumber: z.string(),
  category: z.string(),
  priority: z.string().trim().min(1, 'Please choose a priority.'),
  notes: z.string(),
  targetPrice: z.string().refine((value) => {
      if (value.trim() === '') {
        return true
      }
      return !Number.isNaN(Number(value)) && Number(value) >= 0
    }, 'Target price must be a positive number.'),
})

type TaskForm = z.infer<typeof taskFormSchema>
type WishlistForm = z.infer<typeof wishlistFormSchema>

function wishlistDefaults(currentRow?: Task): WishlistForm {
  return {
    title: currentRow?.title ?? '',
    partNumber: currentRow?.partNumber ?? '',
    category: currentRow?.label ?? '',
    priority: currentRow?.priority ?? 'medium',
    notes: currentRow?.notes ?? '',
    targetPrice:
      typeof currentRow?.targetPrice === 'number' && currentRow.targetPrice > 0
        ? String(currentRow.targetPrice)
        : '',
  }
}

export function TasksMutateDrawer({
  open,
  onOpenChange,
  currentRow,
  routePath,
  onWishlistSubmit,
  isLoading = false,
  canNavigatePrevious = false,
  canNavigateNext = false,
  onNavigatePrevious,
  onNavigateNext,
}: TaskMutateDrawerProps) {
  const isUpdate = !!currentRow
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'

  const taskForm = useForm<TaskForm>({
    resolver: zodResolver(taskFormSchema),
    defaultValues: currentRow ?? {
      title: '',
      status: '',
      label: '',
      priority: '',
    },
  })

  const wishlistForm = useForm<WishlistForm>({
    resolver: zodResolver(wishlistFormSchema),
    defaultValues: wishlistDefaults(currentRow),
  })

  useEffect(() => {
    if (!isWishlistRoute) {
      return
    }
    wishlistForm.reset(wishlistDefaults(currentRow))
  }, [currentRow, isWishlistRoute, wishlistForm])

  const handleClose = (nextOpen: boolean) => {
    onOpenChange(nextOpen)
    taskForm.reset()
    wishlistForm.reset(wishlistDefaults(currentRow))
  }

  const onTaskSubmit = (data: TaskForm) => {
    onOpenChange(false)
    taskForm.reset()
    showSubmittedData(data)
  }

  const onWishlistFormSubmit = async (data: WishlistForm) => {
    if (!onWishlistSubmit) {
      return
    }
    await onWishlistSubmit(
      {
        title: data.title.trim(),
        partNumber: data.partNumber.trim(),
        category: data.category.trim(),
        priority: data.priority.trim(),
        notes: data.notes.trim(),
        targetPrice: data.targetPrice.trim(),
      },
      currentRow
    )
    onOpenChange(false)
    wishlistForm.reset(wishlistDefaults(undefined))
  }

  if (isWishlistRoute) {
    return (
      <Sheet open={open} onOpenChange={handleClose}>
        <SheetContent
          className='flex flex-col'
          data-testid={
            isUpdate ? 'wishlist-edit-panel' : 'wishlist-create-panel'
          }
          data-side='right'
          side='right'
        >
          <SheetHeader className='text-start'>
            <SheetTitle>
              {isUpdate ? 'Edit Wishlist Entry' : 'Create Wishlist Entry'}
            </SheetTitle>
            <SheetDescription>
              {isUpdate
                ? 'Update the selected wishlist entry and keep planning details in sync.'
                : 'Create a new wishlist entry with the details you want to track.'}
            </SheetDescription>
            {isUpdate ? (
              <div className='flex flex-wrap items-center gap-2 pt-2'>
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  data-testid='wishlist-edit-previous'
                  disabled={!canNavigatePrevious || isLoading}
                  onClick={onNavigatePrevious}
                >
                  Previous
                </Button>
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  data-testid='wishlist-edit-next'
                  disabled={!canNavigateNext || isLoading}
                  onClick={onNavigateNext}
                >
                  Next
                </Button>
              </div>
            ) : null}
          </SheetHeader>
          <Form {...wishlistForm}>
            <form
              id='wishlist-form'
              onSubmit={wishlistForm.handleSubmit(onWishlistFormSubmit)}
              className='flex-1 space-y-6 overflow-y-auto px-4'
            >
              <FormField
                control={wishlistForm.control}
                name='title'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Title</FormLabel>
                    <FormControl>
                      <Input {...field} placeholder='Enter a title' />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={wishlistForm.control}
                name='partNumber'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Part Number</FormLabel>
                    <FormControl>
                      <Input {...field} placeholder='Optional part number' />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={wishlistForm.control}
                name='category'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Category</FormLabel>
                    <FormControl>
                      <Input {...field} placeholder='Optional category' />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={wishlistForm.control}
                name='priority'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Priority</FormLabel>
                    <SelectDropdown
                      defaultValue={field.value}
                      onValueChange={field.onChange}
                      placeholder='Select priority'
                      items={priorities.map((priority) => ({
                        label: priority.label,
                        value: priority.value,
                      }))}
                    />
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={wishlistForm.control}
                name='targetPrice'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Target Price</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        type='number'
                        min='0'
                        step='0.01'
                        placeholder='Optional target price'
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={wishlistForm.control}
                name='notes'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Notes</FormLabel>
                    <FormControl>
                      <Textarea
                        {...field}
                        placeholder='Add planning notes'
                        className='min-h-24'
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </form>
          </Form>
          <SheetFooter className='gap-2'>
            <SheetClose asChild>
              <Button variant='outline' disabled={isLoading}>
                Close
              </Button>
            </SheetClose>
            <Button form='wishlist-form' type='submit' disabled={isLoading}>
              {isLoading
                ? isUpdate
                  ? 'Saving...'
                  : 'Creating...'
                : 'Save changes'}
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    )
  }

  return (
    <Sheet open={open} onOpenChange={handleClose}>
      <SheetContent className='flex flex-col'>
        <SheetHeader className='text-start'>
          <SheetTitle>{isUpdate ? 'Update' : 'Create'} Task</SheetTitle>
          <SheetDescription>
            {isUpdate
              ? 'Update the task by providing necessary info.'
              : 'Add a new task by providing necessary info.'}
            Click save when you&apos;re done.
          </SheetDescription>
        </SheetHeader>
        <Form {...taskForm}>
          <form
            id='tasks-form'
            onSubmit={taskForm.handleSubmit(onTaskSubmit)}
            className='flex-1 space-y-6 overflow-y-auto px-4'
          >
            <FormField
              control={taskForm.control}
              name='title'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Title</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder='Enter a title' />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={taskForm.control}
              name='status'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Status</FormLabel>
                  <SelectDropdown
                    defaultValue={field.value}
                    onValueChange={field.onChange}
                    placeholder='Select dropdown'
                    items={[
                      { label: 'In Progress', value: 'in progress' },
                      { label: 'Backlog', value: 'backlog' },
                      { label: 'Todo', value: 'todo' },
                      { label: 'Canceled', value: 'canceled' },
                      { label: 'Done', value: 'done' },
                    ]}
                  />
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={taskForm.control}
              name='label'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Label</FormLabel>
                  <FormControl>
                    <Input {...field} placeholder='Enter a label' />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={taskForm.control}
              name='priority'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Priority</FormLabel>
                  <SelectDropdown
                    defaultValue={field.value}
                    onValueChange={field.onChange}
                    placeholder='Select priority'
                    items={priorities.map((priority) => ({
                      label: priority.label,
                      value: priority.value,
                    }))}
                  />
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>
        <SheetFooter className='gap-2'>
          <SheetClose asChild>
            <Button variant='outline'>Close</Button>
          </SheetClose>
          <Button form='tasks-form' type='submit'>
            Save changes
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
