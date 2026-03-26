describe('general/ui-data-contract-parity', () => {
  function bootstrapAndSignIn(path = '/') {
    cy.e2eReset()
    cy.e2eBootstrap().then((bootstrap) => {
      cy.useBootstrappedProfile(bootstrap.profile_id, bootstrap.profile_name, { path })
    })
  }

  it('UI-DATA-CONTRACT-PARITY-001 maps authenticated screens to explicit API contracts', () => {
    cy.intercept('GET', '/api/dashboard', {
      statusCode: 200,
      body: {
        new_discoveries: 1,
        wishlist_hits: 2,
        price_drops: 0,
        low_stock_discoveries: 0,
        restocks: 0,
        recently_added: ['Parity Item'],
        total_items: 10,
        total_instances: 20,
        estimated_value: 1200,
        cards: [{ title: 'Wishlist', value: 2, link: '/wishlist' }],
      },
    }).as('dashboard')
    cy.intercept('GET', '/api/items*', {
      statusCode: 200,
      body: { items: [] },
    }).as('items')
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'e2e-profile-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'ebay',
            display_name: 'eBay',
            base_domain: 'ebay.com',
            integration_mode: 'official_api',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: true,
            capabilities: {
              search: true,
              stock_observation: false,
              pricing: true,
              health: true,
            },
            health: { status: 'ok' },
            last_run: { status: 'success' },
          },
        ],
      },
    }).as('providers')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    }).as('settings')

    bootstrapAndSignIn('/')
    cy.wait('@dashboard')
    cy.contains('Home').should('be.visible')

    cy.get('[data-testid="sidebar-nav-link-inventory"]').click()
    cy.wait('@items')
    cy.location('pathname').should('match', /^\/inventory\/?$/)

    cy.get('[data-testid="sidebar-nav-link-integrations"]').click()
    cy.wait('@activeProfile')
    cy.wait('@providers')
    cy.wait('@settings')
    cy.location('pathname').should('match', /^\/integrations\/?$/)
  })

  it('UI-DATA-CONTRACT-PARITY-002 renders deterministic load error and recovery states', () => {
    let attempts = 0
    cy.intercept('GET', '/api/dashboard', (req) => {
      attempts += 1
      if (attempts === 1) {
        req.reply({ statusCode: 500, body: { error: 'dashboard_failed' } })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          new_discoveries: 0,
          wishlist_hits: 0,
          price_drops: 0,
          low_stock_discoveries: 0,
          restocks: 0,
          recently_added: [],
          total_items: 1,
          total_instances: 1,
          estimated_value: 10,
          cards: [],
        },
      })
    }).as('dashboardRetry')

    bootstrapAndSignIn('/')
    cy.wait('@dashboardRetry')
    cy.contains('Dashboard unavailable').should('be.visible')
    cy.contains('button', 'Retry').click()
    cy.wait('@dashboardRetry')
    cy.contains('Dashboard unavailable').should('not.exist')
    cy.contains('Inventory Items').should('be.visible')
    cy.contains('500').should('not.exist')
  })

  it('UI-DATA-CONTRACT-PARITY-003 preserves settings context and surfaces inline mutation failure', () => {
    cy.intercept('PUT', '/api/profiles/*/settings', {
      statusCode: 500,
      body: { error: 'profile_settings_save_500' },
    }).as('saveProfileSettings')

    bootstrapAndSignIn('/settings/')
    cy.get('input[placeholder="cabinet-user"]').clear().type('parity-user')
    cy.contains('button', 'Update profile').click()
    cy.wait('@saveProfileSettings')
    cy.location('pathname').should('match', /^\/settings\/profile\/?$/)
    cy.contains('profile_settings_save_500').should('be.visible')
  })

  it('UI-DATA-CONTRACT-PARITY-004 maps parity contracts to executable Cypress evidence', () => {
    cy.readFile('../openspec/specs/general/ui-data-contract-parity/spec.md').then((content) => {
      expect(content).to.contain('ui.web/cypress/e2e/general/ui-data-contract-parity/spec.cy.ts')
      expect(content).to.contain('screen identifier')
      expect(content).to.contain('endpoint(s) exercised')
      expect(content).to.contain('expected state assertion(s)')
    })
  })
})
