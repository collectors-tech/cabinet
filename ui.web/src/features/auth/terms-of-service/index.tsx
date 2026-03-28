export function TermsOfService() {
  return (
    <main className='container mx-auto flex min-h-svh max-w-3xl flex-col gap-6 px-6 py-12'>
      <div className='space-y-2'>
        <p className='text-sm font-medium uppercase tracking-[0.2em] text-muted-foreground'>
          Cabinet Legal
        </p>
        <h1 className='text-4xl font-semibold tracking-tight'>Terms of Service</h1>
        <p className='text-base text-muted-foreground'>
          Cabinet sets clear expectations for acceptable workspace use, account responsibility, and
          how optional integrations and local runtime features should be used.
        </p>
      </div>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Acceptable use</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Cabinet is intended for lawful collection management, inventory workflows, and explicitly
          user-directed automation. You remain responsible for the actions initiated from your
          workspace and connected integrations.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Account and data responsibility</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          You are responsible for protecting access to your local runtime, profile data, and any
          credentials configured for providers, messaging, or external services.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Service boundaries</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Optional integrations depend on third-party availability and policies. Cabinet can expose
          legal and operational guidance, but external provider behavior may change independently of
          the local application runtime.
        </p>
      </section>
    </main>
  )
}
