import { cn } from '@/lib/utils'

type MainProps = React.HTMLAttributes<HTMLElement> & {
  fixed?: boolean
  fluid?: boolean
  ref?: React.Ref<HTMLElement>
}

export function Main({ fixed, className, fluid, ...props }: MainProps) {
  return (
    <main
      data-layout={fixed ? 'fixed' : 'auto'}
      data-testid='app-main-content'
      className={cn(
        'w-full px-4 py-6',

        // If layout is fixed, make the main container flex and grow
        fixed && 'flex grow flex-col overflow-hidden',

        // Only apply a centered max-width when a screen explicitly opts out
        // of the default fluid workspace layout.
        fluid === false &&
          '@7xl/content:mx-auto @7xl/content:w-full @7xl/content:max-w-7xl',
        className
      )}
      {...props}
    />
  )
}
