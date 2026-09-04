import { cn } from '@/lib/utils'

type DatabaseProfileIconProps = {
  label: string
  variant?: 'dark' | 'light' | 'theme'
  className?: string
  testId?: string
}

type DatabaseProfileIconVariantProps = {
  variant: 'dark' | 'light'
  className?: string
  testId?: string
}

function DatabaseProfileIconVariant({
  variant,
  className,
  testId,
}: DatabaseProfileIconVariantProps) {
  const isDark = variant === 'dark'

  return (
    <span
      className={cn(
        'inline-flex size-full items-center justify-center rounded-[6px]',
        isDark
          ? 'bg-[#0d1117] text-white'
          : 'border border-[#d0d7de] bg-white text-[#0d1117]',
        className
      )}
      data-db-icon-variant={variant}
      data-testid={testId}
    >
      <svg
        aria-hidden='true'
        className='size-[68%]'
        fill='none'
        viewBox='0 0 24 24'
        xmlns='http://www.w3.org/2000/svg'
      >
        <ellipse
          cx='12'
          cy='6.5'
          rx='6.5'
          ry='3'
          stroke='currentColor'
          strokeLinecap='round'
          strokeWidth='1.8'
        />
        <path
          d='M5.5 6.5v7.8c0 1.65 2.91 3 6.5 3s6.5-1.35 6.5-3V6.5'
          stroke='currentColor'
          strokeLinecap='round'
          strokeWidth='1.8'
        />
        <path
          d='M5.5 10.5c0 1.65 2.91 3 6.5 3s6.5-1.35 6.5-3'
          stroke='currentColor'
          strokeLinecap='round'
          strokeWidth='1.8'
        />
      </svg>
    </span>
  )
}

export function DatabaseProfileIcon({
  label,
  variant = 'theme',
  className,
  testId,
}: DatabaseProfileIconProps) {
  if (variant !== 'theme') {
    return (
      <span
        aria-label={label}
        className={cn('inline-flex items-center justify-center', className)}
        data-testid={testId}
        role='img'
      >
        <DatabaseProfileIconVariant
          variant={variant}
          testId={testId ? `${testId}-variant` : undefined}
        />
      </span>
    )
  }

  return (
    <span
      aria-label={label}
      className={cn('inline-flex items-center justify-center', className)}
      data-testid={testId}
      role='img'
    >
      <DatabaseProfileIconVariant
        className='dark:hidden'
        testId={testId ? `${testId}-light` : undefined}
        variant='light'
      />
      <DatabaseProfileIconVariant
        className='hidden dark:inline-flex'
        testId={testId ? `${testId}-dark` : undefined}
        variant='dark'
      />
    </span>
  )
}
