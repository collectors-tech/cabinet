import { type SVGProps } from 'react'
import { cn } from '@/lib/utils'

export function IconApple({ className, ...props }: SVGProps<SVGSVGElement>) {
  return (
    <svg
      role='img'
      viewBox='0 0 24 24'
      xmlns='http://www.w3.org/2000/svg'
      width='24'
      height='24'
      className={cn('fill-current', className)}
      {...props}
    >
      <title>Apple</title>
      <path d='M17.05 12.54c-.02-2.2 1.8-3.25 1.88-3.3-1.03-1.5-2.62-1.7-3.18-1.73-1.35-.14-2.64.8-3.33.8-.7 0-1.77-.78-2.9-.76-1.5.02-2.88.87-3.65 2.2-1.55 2.7-.4 6.7 1.12 8.9.74 1.07 1.62 2.28 2.78 2.23 1.12-.04 1.54-.72 2.9-.72 1.35 0 1.74.72 2.93.7 1.2-.02 1.98-1.1 2.72-2.18.85-1.24 1.2-2.45 1.22-2.51-.03-.02-2.35-.9-2.49-3.63zM14.86 6.08c.62-.75 1.04-1.8.93-2.84-.9.04-2 .6-2.65 1.35-.58.67-1.1 1.74-.96 2.77 1 .08 2.04-.51 2.68-1.28z' />
    </svg>
  )
}
