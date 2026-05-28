describe('integrations/ui-screen-market-watch', () => {
  function signInToMarketWatch() {
    cy.visit('/sign-in?redirect=%2Fscanner%2F')
    cy.get('input[name="email"]').clear().type('e2e-market-watch@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/scanner\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
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
    cy.get('[data-testid="scanner-create-query"]').click()
    cy.wait('@createScopedQuery')
    cy.get('[data-testid="scanner-query-providers-qs-mw-1"]').should('contain', 'amazon')
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
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: {
        query_sets: [
          {
            id: 'qs-mw-failure',
            name: 'eBay Auth Recovery',
            keywords: ['pokemon'],
            provider_scope: ['ebay'],
            last_run_status: 'failed',
            last_run_at: '2026-05-28T03:51:00Z',
            last_run_message: 'eBay credentials need attention',
          },
        ],
      },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', {
      statusCode: 200,
      body: {
        failures: [
          {
            id: 'failure-mw-1',
            query_set_id: 'qs-mw-failure',
            provider: 'ebay',
            message: 'eBay credentials need attention',
            created_at: '2026-05-28T03:51:00Z',
          },
        ],
      },
    }).as('failures')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'degraded', message: 'Auth refresh required' },
    }).as('providerHealth')
    cy.intercept('POST', '/api/scanner/run', {
      statusCode: 500,
      body: { error: 'PROVIDER_AUTH_INVALID' },
    }).as('runFailure')
    cy.intercept('POST', '/api/scanner/failures/retry', (req) => {
      expect(req.body.query_set_id).to.equal('qs-mw-failure')
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
      .should('contain', 'Market Watch service is temporarily unavailable.')
      .and('contain', 'Check diagnostics for provider/runtime health.')
      .and('contain', 'PROVIDER_AUTH_INVALID')

    cy.get('[data-testid="scanner-failures"]').within(() => {
      cy.contains('ebay').should('be.visible')
      cy.contains('eBay credentials need attention').should('be.visible')
      cy.get('[data-testid="scanner-retry-qs-mw-failure"]').click()
    })
    cy.wait('@retryFailure')
    cy.get('[data-testid="scanner-action-feedback"]').should('not.exist')
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
            last_run_status: 'succeeded',
            last_run_at: '2026-05-26T06:41:00Z',
            last_candidate_count: 3,
          },
          {
            id: 'qs-mw-table-2',
            name: 'eBay HO Scan',
            keywords: ['HO slot'],
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

    signInToMarketWatch()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="market-watch-view-mode-table"]').click()
    cy.get('[data-testid="market-watch-query-table"]').should('be.visible')
    cy.get('[data-testid="market-watch-query-table"]').within(() => {
      cy.contains('th', 'Query Name').should('be.visible')
      cy.contains('th', 'Provider Scope').should('be.visible')
      cy.contains('th', 'Last Run Status').should('be.visible')
      cy.contains('th', 'Last Run Time').should('be.visible')
      cy.contains('th', 'Latest Output Summary').should('be.visible')
      cy.contains('td', 'Bonza AFX Watch').should('be.visible')
      cy.contains('td', 'bonzaslotcars').should('be.visible')
      cy.contains('td', 'succeeded').should('be.visible')
      cy.contains('td', 'Candidates: 3').should('be.visible')
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
      cy.contains('td', 'No output').should('be.visible')
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
      cy.contains('td', 'Candidates: 4').should('be.visible')
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
})
