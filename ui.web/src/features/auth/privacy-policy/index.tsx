export function PrivacyPolicy() {
  return (
    <main className='container mx-auto flex min-h-svh max-w-3xl flex-col gap-6 px-6 py-12'>
      <div className='space-y-2'>
        <p className='text-sm font-medium uppercase tracking-[0.2em] text-muted-foreground'>
          Cabinet Legal
        </p>
        <h1 className='text-4xl font-semibold tracking-tight'>Privacy Policy</h1>
        <p className='text-base text-muted-foreground'>
          Cabinet explains what profile data is stored locally, what optional integrations can
          transmit externally, and how users retain control over workspace data.
        </p>
      </div>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>What Cabinet stores</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Cabinet stores profile-scoped collection records, settings, activity history, and any
          optional media that you explicitly add to the workspace runtime.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Optional integrations</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          External providers run only when you enable them. Search, scanner, and AI integrations may
          send the query, item, or provider configuration required to complete the action you choose.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Your control</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          You can review, export, update, or remove your collection data from Cabinet, and you can
          return to sign-in or home without losing access to this policy page.
        </p>
      </section>
    </main>
  )
}
