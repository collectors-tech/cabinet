describe('integrations/ui-screen-market-watch', () => {
  function signInToMarketWatch(redirectPath = '/scanner/') {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(redirectPath)}`)
    cy.get('input[name="email"]').clear().type('e2e-market-watch@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/scanner\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-MARKET-WATCH-004 shows deterministic workspace states', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      delay: 500,
      statusCode: 200,
      body: { query_sets: [] },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'auth_required' },
    }).as('providerHealth')

    signInToMarketWatch()
    cy.get('[data-testid="scanner-loading-state"]').should('be.visible')
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-empty-state"]')
      .should('be.visible')
      .and('contain', 'Create your first query set')
    cy.get('[data-testid="market-watch-provider-attention-state"]')
      .should('be.visible')
      .and('contain', 'Provider needs attention')
      .and('contain', 'auth_required')
      .and('contain', 'Review provider credentials')
  })

  it('UI-SCREEN-MARKET-WATCH-004 shows load failure with retry recovery', () => {
    let failedOnce = false

    cy.intercept('GET', '/api/scanner/query-sets', (req) => {
      if (!failedOnce) {
        failedOnce = true
        req.reply({ statusCode: 500, body: { error: 'scanner unavailable' } })
        return
      }
      req.reply({ statusCode: 200, body: { query_sets: [] } })
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-error-state"]')
      .should('be.visible')
      .and('contain', 'Market Watch data is unavailable.')
      .and('contain', 'query_sets_500')
    cy.get('[data-testid="scanner-error-retry"]').click()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.get('[data-testid="scanner-error-state"]').should('not.exist')
    cy.get('[data-testid="scanner-empty-state"]').should('be.visible')
  })

  it('UI-SCREEN-MARKET-WATCH-001 creates provider-scoped query sets from selector controls', () => {
    cy.intercept('GET', '/api/scanner/query-sets', { statusCode: 200, body: { query_sets: [] } }).as(
      'querySets'
    )
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/query-sets', (req) => {
      expect(req.body.provider_scope).to.deep.equal(['amazon'])
      req.reply({
        statusCode: 201,
        body: {
          id: 'qs-mw-1',
          name: 'Amazon Scope',
          keywords: ['slot'],
          provider_scope: ['amazon'],
        },
      })
    }).as('createScopedQuery')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-provider-mode"]').should('have.value', 'single')
    cy.get('[data-testid="market-watch-provider-single"]').select('amazon')
    cy.get('[data-testid="scanner-new-query-name"]').type('Amazon Scope')
    cy.get('[data-testid="scanner-new-query-keywords"]').type('slot')
    cy.get('[data-testid="market-watch-toolbar-create-query"]').click()
    cy.wait('@createScopedQuery')
    cy.get('[data-testid="scanner-query-providers-qs-mw-1"]').should('contain', 'amazon')
  })

  it('UI-SCREEN-MARKET-WATCH-011 creates saved query from route barcode handoff state', () => {
    cy.intercept('GET', '/api/scanner/query-sets', { statusCode: 200, body: { query_sets: [] } }).as(
      'querySets'
    )
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/query-sets', (req) => {
      expect(req.body.name).to.equal('Barcode 9312345678901')
      expect(req.body.keywords).to.deep.equal(['9312345678901'])
      expect(req.body.provider_scope).to.deep.equal(['ebay'])
      req.reply({
        statusCode: 201,
        body: {
          id: 'qs-mw-barcode',
          name: 'Barcode 9312345678901',
          keywords: ['9312345678901'],
          provider_scope: ['ebay'],
        },
      })
    }).as('createBarcodeQuery')

    signInToMarketWatch('/scanner/?barcode=9312345678901')
    cy.location('search').should('contain', 'barcode=9312345678901')
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-new-query-name"]').should(
      'have.value',
      'Barcode 9312345678901'
    )
    cy.get('[data-testid="scanner-new-query-keywords"]').should('have.value', '9312345678901')
    cy.get('[data-testid="scanner-action-status"]').should(
      'contain',
      'barcode_lookup_ready_9312345678901'
    )
    cy.get('[data-testid="scanner-action-feedback"]')
      .should('contain', 'Barcode lookup is ready for Market Watch.')
      .and('contain', 'Review provider scope before creating the query set.')
    cy.get('[data-testid="scanner-create-query"]').click()
    cy.wait('@createBarcodeQuery')
    cy.get('[data-testid="scanner-query-providers-qs-mw-barcode"]').should('contain', 'ebay')
  })

  it('UI-SCREEN-MARKET-WATCH-001 manages saved-query create edit and delete lifecycle', () => {
    let querySets: Array<{
      id: string
      name: string
      keywords: string[]
      provider_scope: string[]
      schedule_cron?: string
      enabled?: boolean
    }> = []

    cy.intercept('GET', '/api/scanner/query-sets', () => ({
      statusCode: 200,
      body: { query_sets: querySets },
    })).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/query-sets', (req) => {
      expect(req.body.provider_scope).to.deep.equal(['bonzaslotcars'])
      expect(req.body.schedule_cron).to.equal('0 */6 * * *')
      querySets = [
        {
          id: 'qs-mw-lifecycle',
          name: req.body.name,
          keywords: req.body.keywords,
          provider_scope: req.body.provider_scope,
          schedule_cron: req.body.schedule_cron,
          enabled: req.body.enabled,
        },
      ]
      req.reply({ statusCode: 201, body: querySets[0] })
    }).as('createQuerySet')
    cy.intercept('PUT', '/api/scanner/query-sets/qs-mw-lifecycle', (req) => {
      expect(req.body.name).to.equal('Bonza AFX Edited')
      expect(req.body.keywords).to.deep.equal(['AFX', 'Mega G+'])
      expect(req.body.provider_scope).to.deep.equal(['bonzaslotcars'])
      expect(req.body.schedule_cron).to.equal('15 */4 * * *')
      querySets = [
        {
          id: 'qs-mw-lifecycle',
          name: req.body.name,
          keywords: req.body.keywords,
          provider_scope: req.body.provider_scope,
          schedule_cron: req.body.schedule_cron,
          enabled: req.body.enabled,
        },
      ]
      req.reply({ statusCode: 200, body: querySets[0] })
    }).as('updateQuerySet')
    cy.intercept('DELETE', '/api/scanner/query-sets/qs-mw-lifecycle', (req) => {
      querySets = []
      req.reply({ statusCode: 204, body: '' })
    }).as('deleteQuerySet')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-provider-single"]').select('bonzaslotcars')
    cy.get('[data-testid="scanner-new-query-name"]').type('Bonza AFX')
    cy.get('[data-testid="scanner-new-query-keywords"]').type('AFX')
    cy.get('[data-testid="scanner-create-query"]').click()
    cy.wait('@createQuerySet')
    cy.get('[data-testid="scanner-query-providers-qs-mw-lifecycle"]').should(
      'contain',
      'bonzaslotcars'
    )
    cy.get('[data-testid="scanner-query-schedule-qs-mw-lifecycle"]').should(
      'contain',
      '0 */6 * * *'
    )

    cy.get('[data-testid="scanner-edit-qs-mw-lifecycle"]').click()
    cy.get('[data-testid="scanner-edit-name-qs-mw-lifecycle"]').clear().type('Bonza AFX Edited')
    cy.get('[data-testid="scanner-edit-keywords-qs-mw-lifecycle"]').clear().type('AFX, Mega G+')
    cy.get('[data-testid="scanner-edit-schedule-qs-mw-lifecycle"]').clear().type('15 */4 * * *')
    cy.get('[data-testid="scanner-save-qs-mw-lifecycle"]').click()
    cy.wait('@updateQuerySet')
    cy.contains('Bonza AFX Edited').should('be.visible')
    cy.contains('AFX, Mega G+').should('be.visible')
    cy.get('[data-testid="scanner-query-providers-qs-mw-lifecycle"]').should(
      'contain',
      'bonzaslotcars'
    )
    cy.get('[data-testid="scanner-query-schedule-qs-mw-lifecycle"]').should(
      'contain',
      '15 */4 * * *'
    )

    cy.get('[data-testid="scanner-delete-qs-mw-lifecycle"]').click()
    cy.wait('@deleteQuerySet')
    cy.contains('Bonza AFX Edited').should('not.exist')
    cy.get('[data-testid="scanner-empty-state"]').should('be.visible')
  })

  it('INTEGRATION-005 + #827 manages eBay saved-query create edit schedule and delete lifecycle', () => {
    let querySets: Array<{
      id: string
      name: string
      keywords: string[]
      provider_scope: string[]
      schedule_cron?: string
      enabled?: boolean
    }> = []

    cy.intercept('GET', '/api/scanner/query-sets', () => ({
      statusCode: 200,
      body: { query_sets: querySets },
    })).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok', state: 'ready', provider: 'ebay' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/query-sets', (req) => {
      expect(req.body.name).to.equal('eBay Slot Cars')
      expect(req.body.keywords).to.deep.equal(['AFX', 'Mega G+'])
      expect(req.body.provider_scope).to.deep.equal(['ebay'])
      expect(req.body.schedule_cron).to.equal('0 */6 * * *')
      querySets = [
        {
          id: 'qs-mw-ebay-lifecycle',
          name: req.body.name,
          keywords: req.body.keywords,
          provider_scope: req.body.provider_scope,
          schedule_cron: req.body.schedule_cron,
          enabled: req.body.enabled,
        },
      ]
      req.reply({ statusCode: 201, body: querySets[0] })
    }).as('createEbayQuerySet')
    cy.intercept('PUT', '/api/scanner/query-sets/qs-mw-ebay-lifecycle', (req) => {
      expect(req.body.name).to.equal('eBay Slot Cars Edited')
      expect(req.body.keywords).to.deep.equal(['AFX', 'Mega G+', 'Tomy'])
      expect(req.body.provider_scope).to.deep.equal(['ebay'])
      expect(req.body.schedule_cron).to.equal('30 */8 * * *')
      querySets = [
        {
          id: 'qs-mw-ebay-lifecycle',
          name: req.body.name,
          keywords: req.body.keywords,
          provider_scope: req.body.provider_scope,
          schedule_cron: req.body.schedule_cron,
          enabled: req.body.enabled,
        },
      ]
      req.reply({ statusCode: 200, body: querySets[0] })
    }).as('updateEbayQuerySet')
    cy.intercept('DELETE', '/api/scanner/query-sets/qs-mw-ebay-lifecycle', (req) => {
      querySets = []
      req.reply({ statusCode: 204, body: '' })
    }).as('deleteEbayQuerySet')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-provider-single"]').select('ebay')
    cy.get('[data-testid="scanner-new-query-name"]').type('eBay Slot Cars')
    cy.get('[data-testid="scanner-new-query-keywords"]').type('AFX, Mega G+')
    cy.get('[data-testid="scanner-new-query-schedule"]').clear().type('0 */6 * * *')
    cy.get('[data-testid="scanner-create-query"]').click()
    cy.wait('@createEbayQuerySet')
    cy.get('[data-testid="scanner-query-providers-qs-mw-ebay-lifecycle"]').should(
      'contain',
      'ebay'
    )
    cy.get('[data-testid="scanner-query-schedule-qs-mw-ebay-lifecycle"]').should(
      'contain',
      '0 */6 * * *'
    )

    cy.get('[data-testid="scanner-edit-qs-mw-ebay-lifecycle"]').click()
    cy.get('[data-testid="scanner-edit-name-qs-mw-ebay-lifecycle"]')
      .clear()
      .type('eBay Slot Cars Edited')
    cy.get('[data-testid="scanner-edit-keywords-qs-mw-ebay-lifecycle"]')
      .clear()
      .type('AFX, Mega G+, Tomy')
    cy.get('[data-testid="scanner-edit-schedule-qs-mw-ebay-lifecycle"]')
      .clear()
      .type('30 */8 * * *')
    cy.get('[data-testid="scanner-save-qs-mw-ebay-lifecycle"]').click()
    cy.wait('@updateEbayQuerySet')
    cy.contains('eBay Slot Cars Edited').should('be.visible')
    cy.contains('AFX, Mega G+, Tomy').should('be.visible')
    cy.get('[data-testid="scanner-query-providers-qs-mw-ebay-lifecycle"]').should(
      'contain',
      'ebay'
    )
    cy.get('[data-testid="scanner-query-schedule-qs-mw-ebay-lifecycle"]').should(
      'contain',
      '30 */8 * * *'
    )

    cy.get('[data-testid="scanner-delete-qs-mw-ebay-lifecycle"]').click()
    cy.wait('@deleteEbayQuerySet')
    cy.contains('eBay Slot Cars Edited').should('not.exist')
    cy.get('[data-testid="scanner-empty-state"]').should('be.visible')
  })

  it('UI-SCREEN-MARKET-WATCH-002 sends provider scope in run payload and shows provider-attributed results', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-2',
            name: 'Dual Scope',
            keywords: ['ho slot'],
            provider_scope: ['ebay', 'amazon'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-2')
      expect(req.body.provider_scope).to.deep.equal(['ebay', 'amazon'])
      req.reply({ statusCode: 200, body: { run_id: 'run-mw-2', status: 'ok' } })
    }).as('runScopedQuery')
    cy.intercept('GET', '/api/scanner/candidates?query_set_id=qs-mw-2', {
      statusCode: 200,
      body: {
        candidates: [
          {
            id: 'cand-1',
            query_set_id: 'qs-mw-2',
            listing_id: 'L-1',
            title: 'AFX Camaro',
            source: 'amazon',
            price: 42.5,
            shipping: 6.25,
            observed_currency: 'USD',
          },
        ],
      },
    }).as('listCandidates')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.get('[data-testid="scanner-run-qs-mw-2"]').click()
    cy.wait('@runScopedQuery')
    cy.wait('@listCandidates')
    cy.get('[data-testid="scanner-candidates-qs-mw-2"]').should('contain', 'amazon')
    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-mw-2"]').click()
    cy.get('[data-testid="market-watch-output-results-table"]').within(() => {
      cy.contains('td', 'AFX Camaro').should('be.visible')
      cy.contains('td', '42.50 USD').should('be.visible')
      cy.contains('td', '6.25 USD').should('be.visible')
    })
  })

  it('UI-SCREEN-MARKET-WATCH-002 runs eBay-only saved searches through the provider route', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-ebay',
            name: 'eBay Scoped',
            keywords: ['afx'],
            provider_scope: ['ebay'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/run', {
      statusCode: 500,
      body: { error: 'unexpected_generic_scanner_run_for_ebay' },
    }).as('genericScannerRun')
    cy.intercept('POST', '/api/providers/ebay/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-ebay')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'ebay',
          candidates: [
            {
              id: 'cand-mw-ebay',
              query_set_id: 'qs-mw-ebay',
              listing_id: 'ebay-1',
              title: 'eBay AFX Camaro',
              source: 'ebay',
            },
          ],
        },
      })
    }).as('runEbayQuery')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-mw-ebay"]').click()
    cy.wait('@runEbayQuery')
    cy.get('@genericScannerRun.all').should('have.length', 0)
    cy.get('[data-testid="scanner-action-status"]').should(
      'contain',
      'ebay_run_started_qs-mw-ebay'
    )
    cy.get('[data-testid="scanner-candidates-qs-mw-ebay"]')
      .should('contain', 'eBay AFX Camaro')
      .and('contain', 'ebay')
  })

  it('UI-SCREEN-MARKET-WATCH-003 blocks create when provider scope is empty in multi-provider mode', () => {
    cy.intercept('GET', '/api/scanner/query-sets', { statusCode: 200, body: { query_sets: [] } }).as(
      'querySets'
    )
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-provider-mode"]').select('multi')
    cy.get('[data-testid="market-watch-provider-checkbox-ebay"]').uncheck({ force: true })
    cy.get('[data-testid="market-watch-provider-checkbox-amazon"]').uncheck({ force: true })
    cy.get('[data-testid="scanner-new-query-name"]').type('No Provider Query')
    cy.get('[data-testid="scanner-new-query-keywords"]').type('slot')
    cy.get('[data-testid="scanner-create-query"]').click()

    cy.get('[data-testid="market-watch-provider-validation"]')
      .should('be.visible')
      .and('contain', 'Select at least one provider')
  })

  it('UI-SCREEN-MARKET-WATCH-003 surfaces run failure guidance and retry action', () => {
    let retryRequested = false

    cy.intercept('GET', '/api/scanner/query-sets', (req) => {
      req.reply({
        statusCode: 200,
        body: {
          query_sets: [
            {
              id: 'qs-mw-failure',
              name: 'eBay Auth Recovery',
              keywords: ['pokemon'],
              provider_scope: ['ebay'],
              last_run_status: retryRequested ? 'succeeded' : 'failed',
              last_run_at: retryRequested
                ? '2026-05-28T04:03:00Z'
                : '2026-05-28T03:51:00Z',
              last_run_message: retryRequested
                ? 'Retry completed successfully'
                : 'eBay credentials need attention',
              last_candidate_count: retryRequested ? 2 : 0,
            },
          ],
        },
      })
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', (req) => {
      req.reply({
        statusCode: 200,
        body: {
          failures: retryRequested
            ? []
            : [
                {
                  id: 'failure-mw-1',
                  query_set_id: 'qs-mw-failure',
                  provider: 'ebay',
                  message: 'eBay credentials need attention',
                  retry_guidance: 'Check provider health, credentials, and retry the operation.',
                  next_action: 'check_provider_health_and_credentials',
                  created_at: '2026-05-28T03:51:00Z',
                },
              ],
        },
      })
    }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'degraded', message: 'Auth refresh required' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/ebay/run', {
      statusCode: 401,
      body: {
        error: 'failed_to_run_scanner',
        error_code: 'PROVIDER_AUTH_INVALID',
        provider: 'ebay',
        query_set_id: 'qs-mw-failure',
        next_action: 'review_provider_credentials_and_health',
      },
    }).as('runFailure')
    cy.intercept('POST', '/api/scanner/failures/retry', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-failure')
      retryRequested = true
      req.reply({ statusCode: 200, body: { status: 'retry_requested' } })
    }).as('retryFailure')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-provider-health"]').should('contain', 'degraded')
    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('td', 'eBay Auth Recovery').should('be.visible')
      cy.contains('td', 'failed').should('be.visible')
    })
    cy.get('[data-testid="market-watch-open-output-qs-mw-failure"]').click()
    cy.get('[data-testid="market-watch-output-detail"]')
      .should('contain', 'Latest failure: eBay credentials need attention')
      .and('contain', 'Last run status: failed')

    cy.get('[data-testid="market-watch-view-mode-cards"]').click()
    cy.get('[data-testid="scanner-run-qs-mw-failure"]').click()
    cy.wait('@runFailure')
    cy.get('[data-testid="scanner-action-feedback"]')
      .should('contain', 'Market Watch action was denied.')
      .and('contain', 'Review provider health and credentials before retrying.')
      .and('contain', 'PROVIDER_AUTH_INVALID')
      .and('not.contain', 'failed_to_run_scanner')

    cy.get('[data-testid="scanner-failures"]').within(() => {
      cy.contains('ebay').should('be.visible')
      cy.contains('eBay credentials need attention').should('be.visible')
      cy.contains('Retry guidance: Check provider health, credentials, and retry the operation.').should(
        'be.visible'
      )
      cy.contains('Next action: check_provider_health_and_credentials').should('be.visible')
      cy.get('[data-testid="scanner-retry-qs-mw-failure"]').click()
    })
    cy.wait('@retryFailure')
    cy.wait(['@querySets', '@failures'])
    cy.get('[data-testid="scanner-action-feedback"]')
      .should('contain', 'Retry requested.')
      .and('contain', 'Refreshing Market Watch failure state.')
    cy.get('[data-testid="scanner-failures"]').should('not.exist')
    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('td', 'succeeded').should('be.visible')
      cy.contains('td', '2').should('be.visible')
    })
  })

  it('INTEGRATION-005 + #827 distinguishes eBay provider search failures from credential denials', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-ebay-search-failure',
            name: 'eBay Browse Recovery',
            keywords: ['afx'],
            provider_scope: ['ebay'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', {
      statusCode: 200,
      body: { failures: [] },
    }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'degraded', message: 'Browse API throttled' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/ebay/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-ebay-search-failure')
      req.reply({
        statusCode: 429,
        body: {
          error: 'failed_to_run_ebay_provider',
          error_code: 'PROVIDER_SEARCH_FAILED',
          provider: 'ebay',
          query_set_id: 'qs-mw-ebay-search-failure',
          message:
            'eBay Browse API request failed: 12000 API rate limit reached',
          next_action: 'check_provider_health_and_credentials',
        },
      })
    }).as('runEbaySearchFailure')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-mw-ebay-search-failure"]').click()
    cy.wait('@runEbaySearchFailure')
    cy.get('[data-testid="scanner-action-feedback"]')
      .should('contain', 'Provider search failed.')
      .and(
        'contain',
        'Check provider health and retry guidance before running this query again.'
      )
      .and(
        'contain',
        'Review credentials only if provider health reports an auth problem.'
      )
      .and('contain', 'PROVIDER_SEARCH_FAILED')
      .and('not.contain', 'Market Watch action was denied.')
      .and('not.contain', 'Sign in again')
  })

  it('INTEGRATION-005 + #827 surfaces eBay setup page-size validation diagnostics', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-ebay-page-size-invalid',
            name: 'eBay Invalid Page Size',
            keywords: ['afx'],
            provider_scope: ['ebay'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', {
      statusCode: 200,
      body: { failures: [] },
    }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok', state: 'ready', provider: 'ebay' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/ebay/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-ebay-page-size-invalid')
      req.reply({
        statusCode: 400,
        body: {
          error: 'invalid_ebay_items_per_page',
          provider: 'ebay',
          query_set_id: 'qs-mw-ebay-page-size-invalid',
          setting: 'integration.ebay.items_per_page',
          next_action: 'update_ebay_items_per_page',
        },
      })
    }).as('runEbayInvalidPageSize')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.get('[data-testid="scanner-run-qs-mw-ebay-page-size-invalid"]').click()
    cy.wait('@runEbayInvalidPageSize')

    cy.get('[data-testid="scanner-action-feedback"]')
      .should('contain', 'Run failed due to query validation.')
      .and('contain', 'Check query keywords and exclusions.')
      .and('contain', 'Review provider health and credentials before retrying.')
      .and('contain', 'invalid_ebay_items_per_page')
      .and('contain', 'setting: integration.ebay.items_per_page')
      .and('contain', 'next_action: update_ebay_items_per_page')
      .and('not.contain', 'Market Watch action was denied.')
      .and('not.contain', 'Sign in again')
  })

  it('INTEGRATION-005 + #827 surfaces eBay invalid query-set diagnostics', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-ebay-stale',
            name: 'eBay Stale Saved Query',
            keywords: ['afx'],
            provider_scope: ['ebay'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', {
      statusCode: 200,
      body: { failures: [] },
    }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok', state: 'ready', provider: 'ebay' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/ebay/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-ebay-stale')
      req.reply({
        statusCode: 400,
        body: {
          error: 'invalid_query_set_id',
          provider: 'ebay',
          query_set_id: 'qs-mw-ebay-stale',
          next_action: 'select_existing_ebay_query_set',
        },
      })
    }).as('runEbayInvalidQuerySet')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.get('[data-testid="scanner-run-qs-mw-ebay-stale"]').click()
    cy.wait('@runEbayInvalidQuerySet')

    cy.get('[data-testid="scanner-action-feedback"]')
      .should('contain', 'Run failed due to query validation.')
      .and('contain', 'Check query keywords and exclusions.')
      .and('contain', 'Review provider health and credentials before retrying.')
      .and('contain', 'invalid_query_set_id')
      .and('contain', 'provider: ebay')
      .and('contain', 'query_set_id: qs-mw-ebay-stale')
      .and('contain', 'next_action: select_existing_ebay_query_set')
      .and('not.contain', 'Market Watch action was denied.')
      .and('not.contain', 'Sign in again')
  })

  it('INTEGRATION-005 + #827 surfaces eBay provider run method diagnostics', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-ebay-method',
            name: 'eBay Method Diagnostic',
            keywords: ['afx'],
            provider_scope: ['ebay'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', {
      statusCode: 200,
      body: { failures: [] },
    }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok', state: 'ready', provider: 'ebay' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/ebay/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-ebay-method')
      req.reply({
        statusCode: 405,
        headers: { Allow: 'POST' },
        body: { error: 'method_not_allowed' },
      })
    }).as('runEbayMethodError')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.get('[data-testid="scanner-run-qs-mw-ebay-method"]').click()
    cy.wait('@runEbayMethodError')

    cy.get('[data-testid="scanner-action-feedback"]')
      .should('contain', 'Run failed.')
      .and('contain', 'Verify provider health and credentials.')
      .and('contain', 'Validate query set configuration, then retry.')
      .and('contain', 'method_not_allowed')
      .and('not.contain', 'Market Watch action was denied.')
      .and('not.contain', 'Sign in again')
  })

  it('INTEGRATION-005 + #827 surfaces eBay provider run pagination metadata and observed-currency output', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-ebay-pagination',
            name: 'eBay Paginated HO Run',
            keywords: ['ho slot car'],
            provider_scope: ['ebay'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/ebay/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-ebay-pagination')
      req.reply({
        statusCode: 200,
        body: {
          query_set_id: 'qs-mw-ebay-pagination',
          provider: 'ebay',
          candidates: [
            {
              id: 'cand-ebay-pagination-1',
              query_set_id: 'qs-mw-ebay-pagination',
              listing_id: 'ebay-ho-1',
              title: 'Aurora HO slot car lot',
              source: 'ebay',
              price: 42,
              shipping: 7.5,
              observed_currency: 'AUD',
              url: 'https://www.ebay.example/itm/ebay-ho-1',
              stock_state: 'in_stock',
              stock_count: 2,
            },
            {
              id: 'cand-ebay-pagination-2',
              query_set_id: 'qs-mw-ebay-pagination',
              listing_id: 'ebay-ho-2',
              title: 'Tyco HO track bundle',
              source: 'ebay',
              price: 64,
              shipping: 0,
              observed_currency: 'AUD',
              url: 'https://www.ebay.example/itm/ebay-ho-2',
              stock_state: 'low_stock',
              stock_count: 1,
            },
          ],
          run: {
            saved: 2,
            attempts: 2,
            items_per_page_requested: 60,
            items_per_page_effective: 48,
            observed_page_size: 24,
            page_count: 3,
            items_per_page_warning: 'requested page size capped at 48',
          },
        },
      })
    }).as('runEbayPagination')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.get('[data-testid="scanner-run-qs-mw-ebay-pagination"]').click()
    cy.wait('@runEbayPagination')

    cy.get('[data-testid="scanner-run-summary-qs-mw-ebay-pagination"]')
      .should('contain', 'Pages: 3')
      .and('contain', 'Candidates: 2')
      .and('contain', 'Observed page size: 24')

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-mw-ebay-pagination"]').click()
    cy.get('[data-testid="market-watch-output-detail"]').within(() => {
      cy.contains('eBay Paginated HO Run').should('be.visible')
      cy.contains('Pages scanned').should('be.visible')
      cy.contains('3').should('be.visible')
      cy.contains('Candidates').should('be.visible')
      cy.contains('2').should('be.visible')
      cy.contains('Observed page size').should('be.visible')
      cy.contains('24').should('be.visible')
      cy.contains('Aurora HO slot car lot').should('be.visible')
      cy.contains('ebay').should('be.visible')
      cy.get('[data-testid="market-watch-output-results-table"]').within(() => {
        cy.contains('td', '42.00 AUD').should('be.visible')
        cy.contains('td', '7.50 AUD').should('be.visible')
        cy.contains('td', '64.00 AUD').should('be.visible')
        cy.contains('td', '0.00 AUD').should('be.visible')
      })
    })
  })

  it('UI-SCREEN-MARKET-WATCH-006 runs Bonza AFX query and surfaces aggregated run summary', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-bonza',
            name: 'AFX',
            keywords: ['AFX'],
            provider_scope: ['bonzaslotcars'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/bonza/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-bonza')
      req.reply({
        statusCode: 200,
        body: {
          query_set_id: 'qs-mw-bonza',
          page_count: 2,
          observed_page_size: 2,
          items_per_page_used: 36,
          candidates: [
            { listing_id: 'bonza-1', title: 'AFX Camaro', source: 'bonzaslotcars' },
            { listing_id: 'bonza-2', title: 'AFX Mustang', source: 'bonzaslotcars' },
            { listing_id: 'bonza-3', title: 'AFX Mega G+', source: 'bonzaslotcars' },
          ],
          run_summary: {
            page_count: 2,
            observed_page_size: 2,
            candidates_total: 3,
          },
        },
      })
    }).as('runBonzaQuery')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])
    cy.get('[data-testid="scanner-run-qs-mw-bonza"]').click()
    cy.wait('@runBonzaQuery')
    cy.get('[data-testid="scanner-run-summary-qs-mw-bonza"]')
      .should('contain', 'Pages: 2')
      .and('contain', 'Candidates: 3')
      .and('contain', 'Observed page size: 2')
  })

  it('UI-SCREEN-MARKET-WATCH-005 renders query table view with saved-query columns for rapid inspection', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-table-1',
            name: 'Bonza AFX Watch',
            keywords: ['AFX', 'Mega G+'],
            provider_scope: ['bonzaslotcars'],
            schedule_cron: '0 */6 * * *',
            enabled: true,
            last_run_status: 'succeeded',
            last_run_at: '2026-05-26T06:41:00Z',
            last_candidate_count: 3,
          },
          {
            id: 'qs-mw-table-2',
            name: 'eBay HO Scan',
            keywords: ['HO slot'],
            provider_scope: ['ebay'],
            enabled: false,
            last_run_status: 'failed',
            last_run_message: 'Provider credentials expired',
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-query-table"]').should('be.visible')
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('th', 'Query Name').should('be.visible')
      cy.contains('th', 'Terms').should('be.visible')
      cy.contains('th', 'Provider Scope').should('be.visible')
      cy.contains('th', 'Schedule').should('be.visible')
      cy.contains('th', 'Latest Status').should('be.visible')
      cy.contains('th', 'Last Run Time').should('be.visible')
      cy.contains('th', 'Result Count').should('be.visible')
      cy.contains('td', 'Bonza AFX Watch').should('be.visible')
      cy.contains('td', 'AFX, Mega G+').should('be.visible')
      cy.contains('td', 'bonzaslotcars').should('be.visible')
      cy.contains('td', 'Scheduled: 0 */6 * * *').should('be.visible')
      cy.contains('td', 'succeeded').should('be.visible')
      cy.contains('td', '3').should('be.visible')
      cy.contains('td', 'eBay HO Scan').should('be.visible')
      cy.contains('td', 'Manual / paused').should('be.visible')
      cy.contains('td', 'failed: Provider credentials expired').should('be.visible')
    })
  })

  it('UI-SCREEN-MARKET-WATCH-007 filters query table rows by provider status schedule attention and result state', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-filter-bonza',
            name: 'Bonza Scheduled Results',
            keywords: ['AFX'],
            provider_scope: ['bonzaslotcars'],
            schedule_cron: '0 */6 * * *',
            enabled: true,
            last_run_status: 'succeeded',
            last_run_at: '2026-05-26T06:41:00Z',
            last_candidate_count: 4,
          },
          {
            id: 'qs-mw-filter-ebay',
            name: 'eBay Failed Manual',
            keywords: ['HO slot'],
            provider_scope: ['ebay'],
            enabled: false,
            last_run_status: 'failed',
            last_run_message: 'Provider credentials expired',
            last_candidate_count: 0,
          },
          {
            id: 'qs-mw-filter-amazon',
            name: 'Amazon Never Run',
            keywords: ['diecast'],
            provider_scope: ['amazon'],
            enabled: true,
            last_run_status: 'never',
            last_candidate_count: 0,
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-filter-summary"]').should('contain', 'Showing 3 of 3')
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('td', 'Bonza Scheduled Results').should('be.visible')
      cy.contains('td', 'eBay Failed Manual').should('be.visible')
      cy.contains('td', 'Amazon Never Run').should('be.visible')
    })
    cy.get('[data-testid="market-watch-run-history"]').within(() => {
      cy.contains('Bonza Scheduled Results').should('be.visible')
      cy.contains('eBay Failed Manual').should('be.visible')
      cy.contains('Amazon Never Run').should('be.visible')
    })

    cy.get('[data-testid="market-watch-filter-provider"]').select('bonzaslotcars')
    cy.get('[data-testid="market-watch-filter-status"]').select('succeeded')
    cy.get('[data-testid="market-watch-filter-schedule"]').select('scheduled')
    cy.get('[data-testid="market-watch-filter-results"]').check()
    cy.get('[data-testid="market-watch-filter-summary"]').should('contain', 'Showing 1 of 3')
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('td', 'Bonza Scheduled Results').should('be.visible')
      cy.contains('td', 'eBay Failed Manual').should('not.exist')
      cy.contains('td', 'Amazon Never Run').should('not.exist')
    })
    cy.get('[data-testid="market-watch-run-history"]').within(() => {
      cy.contains('Bonza Scheduled Results').should('be.visible')
      cy.contains('eBay Failed Manual').should('not.exist')
    })

    cy.get('[data-testid="market-watch-filter-status"]').select('failed')
    cy.get('[data-testid="market-watch-filter-empty"]')
      .should('be.visible')
      .and('contain', 'No Market Watch queries match')
    cy.get('[data-testid="market-watch-filter-empty-reset"]').click()
    cy.get('[data-testid="market-watch-filter-summary"]').should('contain', 'Showing 3 of 3')

    cy.get('[data-testid="market-watch-filter-attention"]').check()
    cy.get('[data-testid="market-watch-filter-summary"]').should('contain', 'Showing 1 of 3')
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('td', 'eBay Failed Manual').should('be.visible')
      cy.contains('td', 'Provider credentials expired').should('be.visible')
      cy.contains('td', 'Bonza Scheduled Results').should('not.exist')
    })
  })

  it('UI-SCREEN-MARKET-WATCH-005 refreshes table run history after scheduled refresh', () => {
    let scheduledCompleted = false

    cy.intercept('GET', '/api/scanner/query-sets', (req) => {
      req.reply({
        statusCode: 200,
        body: {
          query_sets: [
            {
              id: 'qs-mw-scheduled',
              name: 'Scheduled AFX Watch',
              keywords: ['AFX'],
              provider_scope: ['bonzaslotcars'],
              schedule_cron: '0 */6 * * *',
              enabled: true,
              ...(scheduledCompleted
                ? {
                    last_run_status: 'succeeded',
                    last_run_at: '2026-05-28T07:58:00Z',
                    last_candidate_count: 4,
                  }
                : {}),
            },
          ],
        },
      })
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/run/scheduled', (req) => {
      scheduledCompleted = true
      req.reply({
        statusCode: 200,
        body: {
          run_id: 'scheduled-mw-1',
          query_sets_executed: 1,
          candidates_collected: 4,
          failures: 0,
        },
      })
    }).as('scheduledRefresh')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('td', 'Scheduled AFX Watch').should('be.visible')
      cy.contains('td', 'never').should('be.visible')
      cy.contains('td', '0').should('be.visible')
    })

    cy.get('[data-testid="scanner-run-scheduled-refresh"]').click()
    cy.wait('@scheduledRefresh')
    cy.wait('@querySets')
    cy.get('[data-testid="scanner-action-status"]').should(
      'contain',
      'scheduled_run_completed_scheduled-mw-1'
    )
    cy.get('[data-testid="scanner-action-feedback"]')
      .should('contain', 'Scheduled refresh completed.')
      .and('contain', 'Candidates collected: 4')
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('td', 'succeeded').should('be.visible')
      cy.contains('td', '4').should('be.visible')
    })
    cy.get('[data-testid="market-watch-run-history-qs-mw-scheduled"]')
      .should('contain', 'Scheduled AFX Watch')
      .and('contain', 'succeeded')
      .and('contain', 'Candidates: 4')
  })

  it('UI-SCREEN-MARKET-WATCH-005 opens deterministic output details from query-table row action', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-detail-1',
            name: 'Bonza AFX Watch',
            keywords: ['AFX'],
            provider_scope: ['bonzaslotcars'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/bonza/run', {
      statusCode: 200,
      body: {
        query_set_id: 'qs-mw-detail-1',
        page_count: 2,
        observed_page_size: 2,
        items_per_page_used: 36,
        candidates: [
          { listing_id: 'bonza-1', title: 'AFX Camaro', source: 'bonzaslotcars' },
          { listing_id: 'bonza-2', title: 'AFX Mustang', source: 'bonzaslotcars' },
        ],
        run_summary: {
          page_count: 2,
          observed_page_size: 2,
          candidates_total: 2,
        },
      },
    }).as('runBonzaQuery')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-mw-detail-1"]').click()
    cy.wait('@runBonzaQuery')

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-mw-detail-1"]').click()
    cy.get('[data-testid="market-watch-output-detail"]').within(() => {
      cy.contains('Bonza AFX Watch').should('be.visible')
      cy.contains('Provider Scope').should('be.visible')
      cy.contains('bonzaslotcars').should('be.visible')
      cy.contains('Pages scanned').should('be.visible')
      cy.contains('2').should('be.visible')
      cy.contains('Candidates').should('be.visible')
    })
  })

  it('UI-SCREEN-MARKET-WATCH-009 persists output-detail Wishlist handoff provenance', () => {
    let wishlistEntries: Array<Record<string, unknown>> = []
    let wishlistItems: Array<Record<string, unknown>> = []

    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-provenance-1',
            name: 'Bonza AFX Provenance',
            keywords: ['AFX'],
            provider_scope: ['bonzaslotcars'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/bonza/run', {
      statusCode: 200,
      body: {
        query_set_id: 'qs-mw-provenance-1',
        candidates: [
          {
            id: 'cand-mw-provenance-1',
            query_set_id: 'qs-mw-provenance-1',
            listing_id: 'bonza-afx-camaro',
            title: 'AFX Camaro Mega-G+',
            source: 'bonzaslotcars',
            price: 89.95,
            currency: 'AUD',
            url: 'https://bonzaslotcars.example/products/afx-camaro',
            stock_status: 'in_stock',
            handoff_state: 'wishlist_ready',
          },
        ],
        run_summary: {
          page_count: 1,
          observed_page_size: 1,
          candidates_total: 1,
        },
      },
    }).as('runBonzaQuery')
    cy.intercept('POST', '/api/discovery/action', (req) => {
      expect(req.body.candidate_id).to.equal('cand-mw-provenance-1')
      expect(req.body.type).to.equal('add_to_wishlist')
      expect(req.body.payload).to.deep.equal({
        source: 'market_watch',
        query_set_id: 'qs-mw-provenance-1',
      })
      wishlistEntries = [
        {
          id: 'wish-mw-provenance-1',
          item_id: 'item-mw-provenance-1',
          priority: 'medium',
          target_price: 89.95,
          notes:
            'source_provider=bonzaslotcars; query_set_id=qs-mw-provenance-1; query_name=Bonza AFX Provenance; provider_scope=bonzaslotcars',
          created_at: '2026-06-11T12:52:00Z',
          updated_at: '2026-06-11T12:52:00Z',
        },
      ]
      wishlistItems = [
        {
          id: 'item-mw-provenance-1',
          title: 'AFX Camaro Mega-G+',
          part_number: 'bonza-afx-camaro',
          status: 'wishlist',
          category: 'Slot Cars',
          priority: 'medium',
        },
      ]
      req.reply({ statusCode: 200, body: { ok: true } })
    }).as('wishlistHandoff')
    cy.intercept('GET', '/api/wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } })
    }).as('wishlistEntries')
    cy.intercept('GET', '/api/items?status=wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } })
    }).as('wishlistItems')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    })
    cy.intercept('GET', '/api/pricing/stats?item_id=item-mw-provenance-1', {
      statusCode: 200,
      body: { min: 89.95, median: 89.95, latest: 89.95 },
    }).as('wishlistPriceStats')
    cy.intercept('GET', '/api/pricing/trend?item_id=item-mw-provenance-1', {
      statusCode: 200,
      body: { points: [] },
    }).as('wishlistPriceTrend')
    cy.intercept('GET', '/api/pricing/history?item_id=item-mw-provenance-1', {
      statusCode: 200,
      body: { history: [] },
    }).as('wishlistPriceHistory')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-mw-provenance-1"]').click()
    cy.wait('@runBonzaQuery')

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-mw-provenance-1"]').click()
    cy.get('[data-testid="market-watch-output-results-table"]').within(() => {
      cy.contains('th', 'Provider').should('be.visible')
      cy.contains('th', 'Handoff').should('be.visible')
      cy.contains('td', 'bonzaslotcars').should('be.visible')
      cy.contains('td', 'AFX Camaro Mega-G+').should('be.visible')
      cy.contains('td', '89.95 AUD').should('be.visible')
      cy.contains('td', 'https://bonzaslotcars.example/products/afx-camaro').should(
        'be.visible'
      )
      cy.contains('td', 'in_stock').should('be.visible')
      cy.contains('td', 'wishlist_ready').should('be.visible')
    })
    cy.get('[data-testid="scanner-handoff-wishlist-qs-mw-provenance-1"]').click()
    cy.wait('@wishlistHandoff')
    cy.get('[data-testid="scanner-handoff-status"]').should(
      'contain',
      'wishlist_handoff_ok_cand-mw-provenance-1'
    )

    cy.visit('/wishlist/')
    cy.wait(['@wishlistEntries', '@wishlistItems'])
    cy.contains('AFX Camaro Mega-G+').should('be.visible')
    cy.contains('source_provider=bonzaslotcars').should('be.visible')
    cy.contains('query_set_id=qs-mw-provenance-1').should('be.visible')
    cy.contains('query_name=Bonza AFX Provenance').should('be.visible')
    cy.contains('provider_scope=bonzaslotcars').should('be.visible')
  })

  it('UI-SCREEN-MARKET-WATCH-010 persists output-detail Inventory handoff provenance', () => {
    let inventoryItems: Array<Record<string, unknown>> = []

    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-inventory-1',
            name: 'Bonza Inventory Provenance',
            keywords: ['AFX'],
            provider_scope: ['bonzaslotcars'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/bonza/run', {
      statusCode: 200,
      body: {
        query_set_id: 'qs-mw-inventory-1',
        candidates: [
          {
            id: 'cand-mw-inventory-1',
            query_set_id: 'qs-mw-inventory-1',
            listing_id: 'bonza-afx-mustang',
            title: 'AFX Mustang GT',
            source: 'bonzaslotcars',
            price: 76.5,
            currency: 'AUD',
            url: 'https://bonzaslotcars.example/products/afx-mustang',
            stock_status: 'in_stock',
            handoff_state: 'inventory_ready',
          },
        ],
        run_summary: {
          page_count: 1,
          observed_page_size: 1,
          candidates_total: 1,
        },
      },
    }).as('runBonzaQuery')
    cy.intercept('POST', '/api/discovery/action', (req) => {
      expect(req.body.candidate_id).to.equal('cand-mw-inventory-1')
      expect(req.body.type).to.equal('create_item')
      expect(req.body.payload).to.deep.equal({
        source: 'market_watch',
        query_set_id: 'qs-mw-inventory-1',
      })
      inventoryItems = [
        {
          id: 'item-mw-inventory-1',
          title: 'AFX Mustang GT',
          part_number: 'bonza-afx-mustang',
          status: 'owned',
          category: 'Slot Cars',
          priority: 'medium',
          notes:
            '{"source_provider":"bonzaslotcars","query_set_id":"qs-mw-inventory-1","query_name":"Bonza Inventory Provenance","provider_scope":"bonzaslotcars","source_result_url":"https://bonzaslotcars.example/products/afx-mustang"}',
          description:
            '{"source_provider":"bonzaslotcars","query_set_id":"qs-mw-inventory-1","query_name":"Bonza Inventory Provenance","provider_scope":"bonzaslotcars","source_result_url":"https://bonzaslotcars.example/products/afx-mustang"}',
        },
      ]
      req.reply({ statusCode: 200, body: { ok: true } })
    }).as('inventoryHandoff')
    cy.intercept('GET', '/api/items', (req) => {
      req.reply({ statusCode: 200, body: { items: inventoryItems } })
    }).as('inventoryItems')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    })

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-mw-inventory-1"]').click()
    cy.wait('@runBonzaQuery')

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-mw-inventory-1"]').click()
    cy.get('[data-testid="market-watch-output-results-table"]').within(() => {
      cy.contains('td', 'bonzaslotcars').should('be.visible')
      cy.contains('td', 'AFX Mustang GT').should('be.visible')
      cy.contains('td', '76.50 AUD').should('be.visible')
      cy.contains('td', 'inventory_ready').should('be.visible')
    })
    cy.get('[data-testid="scanner-handoff-inventory-qs-mw-inventory-1"]').click()
    cy.wait('@inventoryHandoff')
    cy.get('[data-testid="scanner-handoff-status"]').should(
      'contain',
      'inventory_handoff_ok_cand-mw-inventory-1'
    )

    cy.visit('/inventory/')
    cy.wait('@inventoryItems')
    cy.contains('button', 'Cards').click()
    cy.get('[data-testid="inventory-item-row-item-mw-inventory-1"]')
      .should('contain', 'AFX Mustang GT')
    cy.get('[data-testid="inventory-card-notes-item-mw-inventory-1"]')
      .should('be.visible')
      .and('contain', 'source_provider')
      .and('contain', 'bonzaslotcars')
      .and('contain', 'query_set_id')
      .and('contain', 'qs-mw-inventory-1')
      .and('contain', 'source_result_url')
      .and('contain', 'https://bonzaslotcars.example/products/afx-mustang')
  })

  it('INTEGRATION-005 + UI-SCREEN-MARKET-WATCH-009 + UI-SCREEN-MARKET-WATCH-010 preserves eBay output handoff response provenance', () => {
    let wishlistEntries: Array<Record<string, unknown>> = []
    let wishlistItems: Array<Record<string, unknown>> = []
    let inventoryItems: Array<Record<string, unknown>> = []
    const ebayHandoffAudit = {
      source: 'market_watch',
      source_provider: 'ebay',
      query_set_id: 'qs-mw-ebay-handoff',
      query_name: 'eBay AFX Handoff',
      provider_scope: ['ebay'],
      listing_id: 'ebay-afx-camaro-1',
      source_result_url: 'https://www.ebay.com/itm/ebay-afx-camaro-1',
      observed_price: 112.5,
      observed_currency: 'AUD',
      seller: 'ebay-seller-1',
    }

    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-ebay-handoff',
            name: 'eBay AFX Handoff',
            keywords: ['AFX Camaro'],
            provider_scope: ['ebay'],
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok', state: 'ready', provider: 'ebay' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/providers/ebay/run', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-ebay-handoff')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'ebay',
          query_set_id: 'qs-mw-ebay-handoff',
          candidates: [
            {
              id: 'cand-mw-ebay-handoff-1',
              query_set_id: 'qs-mw-ebay-handoff',
              listing_id: 'ebay-afx-camaro-1',
              title: 'eBay AFX Camaro Collector Lot',
              source: 'ebay',
              price: 112.5,
              shipping: 8.75,
              currency: 'AUD',
              url: 'https://www.ebay.com/itm/ebay-afx-camaro-1',
              seller: 'ebay-seller-1',
              stock_status: 'available',
              handoff_state: 'wishlist_inventory_ready',
            },
          ],
          run_summary: {
            page_count: 1,
            observed_page_size: 1,
            candidates_total: 1,
          },
        },
      })
    }).as('runEbayQuery')
    cy.intercept('POST', '/api/discovery/action', (req) => {
      expect(req.body.candidate_id).to.equal('cand-mw-ebay-handoff-1')
      expect(req.body.payload).to.deep.equal({
        source: 'market_watch',
        query_set_id: 'qs-mw-ebay-handoff',
      })
      if (req.body.type === 'add_to_wishlist') {
        wishlistEntries = [
          {
            id: 'wish-mw-ebay-handoff-1',
            item_id: 'item-mw-ebay-handoff-wishlist-1',
            priority: 'medium',
            target_price: 112.5,
            notes:
              'source_provider=ebay; query_set_id=qs-mw-ebay-handoff; query_name=eBay AFX Handoff; provider_scope=ebay',
            created_at: '2026-06-16T08:24:00Z',
            updated_at: '2026-06-16T08:24:00Z',
          },
        ]
        wishlistItems = [
          {
            id: 'item-mw-ebay-handoff-wishlist-1',
            title: 'eBay AFX Camaro Collector Lot',
            part_number: 'ebay-afx-camaro-1',
            status: 'wishlist',
            category: 'Slot Cars',
            priority: 'medium',
          },
        ]
        req.reply({
          statusCode: 200,
          body: {
            ok: true,
            action: 'add_to_wishlist',
            candidate_id: 'cand-mw-ebay-handoff-1',
            audit: ebayHandoffAudit,
          },
        })
        return
      }
      expect(req.body.type).to.equal('create_item')
      inventoryItems = [
        {
          id: 'item-mw-ebay-handoff-inventory-1',
          title: 'eBay AFX Camaro Collector Lot',
          part_number: 'ebay-afx-camaro-1',
          status: 'owned',
          category: 'Slot Cars',
          priority: 'medium',
          notes:
            '{"source_provider":"ebay","query_set_id":"qs-mw-ebay-handoff","query_name":"eBay AFX Handoff","provider_scope":"ebay","source_result_url":"https://www.ebay.com/itm/ebay-afx-camaro-1"}',
          description:
            '{"source_provider":"ebay","query_set_id":"qs-mw-ebay-handoff","query_name":"eBay AFX Handoff","provider_scope":"ebay","source_result_url":"https://www.ebay.com/itm/ebay-afx-camaro-1"}',
        },
      ]
      req.reply({
        statusCode: 200,
        body: {
          ok: true,
          action: 'create_item',
          candidate_id: 'cand-mw-ebay-handoff-1',
          audit: ebayHandoffAudit,
        },
      })
    }).as('discoveryHandoff')
    cy.intercept('GET', '/api/wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } })
    }).as('wishlistEntries')
    cy.intercept('GET', '/api/items?status=wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } })
    }).as('wishlistItems')
    cy.intercept('GET', '/api/items', (req) => {
      req.reply({ statusCode: 200, body: { items: inventoryItems } })
    }).as('inventoryItems')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    })
    cy.intercept('GET', '/api/pricing/stats?item_id=item-mw-ebay-handoff-wishlist-1', {
      statusCode: 200,
      body: { min: 112.5, median: 112.5, latest: 112.5 },
    }).as('wishlistPriceStats')
    cy.intercept('GET', '/api/pricing/trend?item_id=item-mw-ebay-handoff-wishlist-1', {
      statusCode: 200,
      body: { points: [] },
    }).as('wishlistPriceTrend')
    cy.intercept('GET', '/api/pricing/history?item_id=item-mw-ebay-handoff-wishlist-1', {
      statusCode: 200,
      body: { history: [] },
    }).as('wishlistPriceHistory')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="scanner-run-qs-mw-ebay-handoff"]').click()
    cy.wait('@runEbayQuery')

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-mw-ebay-handoff"]').click()
    cy.get('[data-testid="market-watch-output-results-table"]').within(() => {
      cy.contains('th', 'Shipping').should('be.visible')
      cy.contains('td', 'ebay').should('be.visible')
      cy.contains('td', 'eBay AFX Camaro Collector Lot').should('be.visible')
      cy.contains('td', '112.50 AUD').should('be.visible')
      cy.contains('td', '8.75 AUD').should('be.visible')
      cy.contains('td', 'https://www.ebay.com/itm/ebay-afx-camaro-1').should(
        'be.visible'
      )
      cy.contains('td', 'available').should('be.visible')
      cy.contains('td', 'wishlist_inventory_ready').should('be.visible')
    })

    cy.get('[data-testid="scanner-handoff-wishlist-qs-mw-ebay-handoff"]').click()
    cy.wait('@discoveryHandoff').then((interception) => {
      expect(interception.response?.body).to.deep.include({
        ok: true,
        action: 'add_to_wishlist',
        candidate_id: 'cand-mw-ebay-handoff-1',
      })
      expect(interception.response?.body.audit).to.deep.equal(ebayHandoffAudit)
    })
    cy.get('[data-testid="scanner-handoff-status"]').should(
      'contain',
      'wishlist_handoff_ok_cand-mw-ebay-handoff-1'
    )
    cy.get('[data-testid="scanner-handoff-inventory-qs-mw-ebay-handoff"]').click()
    cy.wait('@discoveryHandoff').then((interception) => {
      expect(interception.response?.body).to.deep.include({
        ok: true,
        action: 'create_item',
        candidate_id: 'cand-mw-ebay-handoff-1',
      })
      expect(interception.response?.body.audit).to.deep.equal(ebayHandoffAudit)
    })
    cy.get('[data-testid="scanner-handoff-status"]').should(
      'contain',
      'inventory_handoff_ok_cand-mw-ebay-handoff-1'
    )

    cy.visit('/wishlist/')
    cy.wait(['@wishlistEntries', '@wishlistItems'])
    cy.contains('eBay AFX Camaro Collector Lot').should('be.visible')
    cy.contains('source_provider=ebay').should('be.visible')
    cy.contains('query_set_id=qs-mw-ebay-handoff').should('be.visible')
    cy.contains('query_name=eBay AFX Handoff').should('be.visible')
    cy.contains('provider_scope=ebay').should('be.visible')

    cy.visit('/inventory/')
    cy.wait('@inventoryItems')
    cy.contains('button', 'Cards').click()
    cy.get('[data-testid="inventory-item-row-item-mw-ebay-handoff-inventory-1"]')
      .should('contain', 'eBay AFX Camaro Collector Lot')
    cy.get('[data-testid="inventory-card-notes-item-mw-ebay-handoff-inventory-1"]')
      .should('be.visible')
      .and('contain', 'source_provider')
      .and('contain', 'ebay')
      .and('contain', 'query_set_id')
      .and('contain', 'qs-mw-ebay-handoff')
      .and('contain', 'source_result_url')
      .and('contain', 'https://www.ebay.com/itm/ebay-afx-camaro-1')
  })

  it('UI-SCREEN-MARKET-WATCH-004 keeps no-output detail state explicit', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-no-output',
            name: 'No Output Watch',
            keywords: ['AFX'],
            provider_scope: ['bonzaslotcars'],
            last_run_status: 'succeeded',
            last_run_at: '2026-06-11T10:30:00Z',
            last_candidate_count: 0,
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-open-output-qs-mw-no-output"]').click()
    cy.get('[data-testid="market-watch-output-no-results"]')
      .should('be.visible')
      .and('contain', 'No output available yet.')
      .and('contain', 'Run this query or adjust provider scope')
  })
})
