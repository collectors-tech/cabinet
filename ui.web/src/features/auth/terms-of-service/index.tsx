export function TermsOfService() {
  return (
    <main className='container mx-auto flex min-h-svh max-w-3xl flex-col gap-6 px-6 py-12'>
      <div className='space-y-2'>
        <p className='text-sm font-medium tracking-[0.2em] text-muted-foreground uppercase'>
          Cabinet 0.1 Private Beta
        </p>
        <h1 className='text-4xl font-semibold tracking-tight'>
          Terms of Service
        </h1>
        <p className='text-base text-muted-foreground'>
          These terms describe the operational use boundary of the Cabinet 0.1
          private beta. Any separate invitation or deployment agreement from
          your beta coordinator also applies.
        </p>
      </div>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Beta package and use</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Cabinet is supplied for invited evaluation as an unsigned Windows
          portable package, not an installer or automatic-update service. Use it
          only for lawful collection management and user-directed workflows. You
          are responsible for verifying the candidate checksum, keeping backups
          outside the extracted folder, and protecting the operating system
          account and local data directory.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Third-party providers</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Optional third-party providers, including configured identity,
          marketplace, AI, messaging, and diagnostics services, operate under
          their own provider terms and availability. You must have authority to
          use each account, credential, page, and item of provider data. Cabinet
          does not control provider changes, access challenges, or retention.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Browser Companion</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Browser Companion is a user-present, passive capture tool. You must
          grant only provider origins you recognise and remain responsible for
          the pages you open and observations you submit. You must not use it
          for unattended crawling, challenge bypass, cookie or credential
          extraction, hidden-page access, provider writes, cart actions, or
          checkout. Revoke a lost, replaced, or suspicious pairing in Cabinet.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Identity and diagnostics</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Local mode and deployment-configured ZITADEL mode have different
          identity boundaries. Do not treat a local profile as a hosted account
          or assume ZITADEL is active when the runtime is not configured for it.
          Remote diagnostics are sent only after you opt in and the runtime has
          a configured endpoint; review redacted exports before sharing them.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Data and stopping use</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Cabinet provides local JSON/CSV export and backup/restore surfaces,
          but you remain responsible for keeping usable copies. Close Cabinet
          before moving or deleting data. Removing the executable alone does not
          reliably remove the configured data directory, and removing local data
          cannot recall information already sent to a third party.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Support and changes</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Report beta problems to the beta coordinator who supplied the exact
          candidate and include only reviewed, redacted evidence. This beta has
          no support service-level commitment. Capabilities can change between
          candidates; the versioned release notes, checksums, and governed
          capability disclosure describe the candidate you received.
        </p>
      </section>
    </main>
  )
}
