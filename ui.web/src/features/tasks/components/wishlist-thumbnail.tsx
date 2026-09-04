import { ImageIcon } from 'lucide-react'
import { type Task } from '../data/schema'

function hashWishlistThumbnailKey(value: string) {
  return [...value].reduce((hash, char) => {
    return (hash * 31 + char.charCodeAt(0)) % 360
  }, 23)
}

export function WishlistThumbnail({
  task,
  variant = 'row',
}: {
  task: Task
  variant?: 'row' | 'card'
}) {
  const key = task.itemID?.trim() || task.id.trim() || task.title.trim()
  const hue = hashWishlistThumbnailKey(key)
  const accentHue = (hue + 46) % 360
  const initials = task.title
    .split(/\s+/)
    .map((word) => word[0])
    .filter(Boolean)
    .slice(0, 2)
    .join('')
    .toUpperCase()

  if (task.thumbnailUrl) {
    return (
      <img
        src={task.thumbnailUrl}
        alt=''
        aria-hidden='true'
        data-testid={
          variant === 'card'
            ? `wishlist-card-thumbnail-${task.id}`
            : `wishlist-thumbnail-${task.id}`
        }
        data-thumbnail-key={key}
        className={
          variant === 'card'
            ? 'h-20 w-full rounded-md border object-cover'
            : 'h-8 w-8 shrink-0 rounded-md border object-cover'
        }
      />
    )
  }

  if (variant === 'card') {
    return (
      <div
        aria-hidden='true'
        data-testid={`wishlist-card-thumbnail-placeholder-${task.id}`}
        data-thumbnail-key={key}
        className='flex h-20 w-full flex-col items-center justify-center gap-1 rounded-md border border-dashed bg-muted/40 text-muted-foreground'
      >
        <ImageIcon className='size-5' />
        <span className='text-xs font-medium'>No asset</span>
      </div>
    )
  }

  return (
    <span
      aria-hidden='true'
      data-testid={`wishlist-thumbnail-${task.id}`}
      data-thumbnail-key={key}
      className='relative flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-md border text-[0.625rem] font-semibold text-white shadow-sm'
      style={{
        background: `linear-gradient(135deg, hsl(${hue} 68% 42%), hsl(${accentHue} 72% 36%))`,
      }}
    >
      <span className='absolute inset-x-0 top-0 h-1/3 bg-white/18' />
      <span className='absolute right-0 bottom-0 h-4 w-4 rounded-tl-full bg-black/20' />
      <span className='relative'>{initials || 'W'}</span>
    </span>
  )
}
