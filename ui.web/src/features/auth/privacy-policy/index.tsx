export function PrivacyPolicy() {
  return (
    <main className='container mx-auto flex min-h-svh max-w-3xl flex-col gap-6 px-6 py-12'>
      <div className='space-y-2'>
        <p className='text-sm font-medium tracking-[0.2em] text-muted-foreground uppercase'>
          Cabinet 0.1 Private Beta
        </p>
        <h1 className='text-4xl font-semibold tracking-tight'>
          Privacy Policy
        </h1>
        <p className='text-base text-muted-foreground'>
          This notice describes the data paths and optional external processing
          implemented by the Cabinet 0.1 private beta. A deployment operator or
          beta invitation may supply additional terms and contact details.
        </p>
      </div>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Local storage and data paths</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Cabinet stores collection records, profile settings, media, provider
          configuration, activity records, backups, and Browser Companion
          capture records in the configured local runtime. A portable launch
          normally uses the data folder beside cabinet.exe; an environment or
          first-run storage override can select another path. The running
          runtime&apos;s /api/runtime response is the authoritative data path.
        </p>
        <p className='text-sm leading-7 text-muted-foreground'>
          Local mode does not require an external identity provider. If Cabinet
          is configured for ZITADEL, the browser and configured ZITADEL
          authority process the sign-in request, and Cabinet stores the local
          identity and session metadata needed to resolve the workspace. The
          deployment operator controls that ZITADEL configuration.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Provider processing</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          A direct provider integration sends the query, identifiers, and
          configuration needed for the action you choose from your Cabinet
          runtime to that third-party provider. The provider also receives
          normal network metadata and applies its own terms and retention. Do
          not configure a provider unless you are authorised to use its account
          and data.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Browser Companion capture</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          The optional Browser Companion pairs to Cabinet over loopback and is
          limited to the enabled provider origins you grant. It submits
          supported page data from a user-present Chrome or Edge tab, including
          bounded listing, price, stock, URL, image, and capture evidence used
          for review in Cabinet. It is prohibited from exporting cookies,
          passwords, tokens, challenge answers, or raw page HTML; bypassing a
          challenge; crawling hidden pages; or performing provider writes.
        </p>
        <p className='text-sm leading-7 text-muted-foreground'>
          Cabinet retains the accepted capture envelope and its provenance in
          the active profile so processing can recover after restart and avoid
          duplicate observations. Revoking a companion session stops later local
          API access but does not erase observations already accepted into the
          workspace.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Diagnostics</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Remote diagnostics are disabled by default. Cabinet keeps runtime and
          activity diagnostics locally until you explicitly opt in and a
          configured remote endpoint is available. An opted-in event can send
          its type, category, message, and recursively redacted details to that
          endpoint. Cabinet replaces sensitive keys and session identifiers and
          redacts known credential, private-page, and local-path patterns, but
          you should still review any exported diagnostics before sharing them.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Retention and deletion</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          The beta has no fixed automatic retention period for local workspace
          data, backups, or diagnostics logs. They remain in the configured data
          locations until you remove supported records or backups, or remove the
          data directory after closing Cabinet. Deleting the portable executable
          alone does not reliably delete data. Data already sent to ZITADEL, a
          provider, or an opted-in diagnostics endpoint is subject to that
          operator&apos;s retention and cannot be deleted by removing local
          Cabinet files.
        </p>
      </section>

      <section className='space-y-3'>
        <h2 className='text-xl font-semibold'>Export and support</h2>
        <p className='text-sm leading-7 text-muted-foreground'>
          Settings provides active-profile JSON and item CSV exports, local
          backup and restore, and a redacted diagnostics export. An export is a
          new copy under your control and is not removed when its source record
          changes. For access, export, deletion, or incident questions, contact
          the beta coordinator who supplied the candidate or the named
          deployment operator. Never include credentials, cookies, tokens,
          Browser Companion secrets, or unreviewed private page content in a
          support request.
        </p>
      </section>
    </main>
  )
}
