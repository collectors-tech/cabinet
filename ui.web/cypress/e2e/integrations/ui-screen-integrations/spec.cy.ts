describe('ui-screen-integrations', () => {
  function signIn() {
    cy.visit('/sign-in?redirect=%2Fintegrations%2F')
    cy.get('input[name="email"]').clear().type('e2e-inventory@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/integrations\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-INTEGRATIONS-001 + UI-SCREEN-INTEGRATIONS-006 + INTEGRATION-022: defaults to cards and supports filter/sort/view using registry data', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
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
            setup_instructions: 'Configure eBay token and marketplace.',
            capabilities: {
              search: true,
              stock_observation: false,
              pricing: true,
              health: true,
            },
            health: { status: 'ok', last_checked_at: '2026-03-01T00:00:00Z' },
            last_run: { status: 'success', finished_at: '2026-03-01T00:00:00Z' },
          },
          {
            provider_id: 'au-webshop-bonzaslotcars-com-au',
            display_name: 'bonzaslotcars.com.au',
            base_domain: 'bonzaslotcars.com.au',
            integration_mode: 'web_ingestion',
            auth_mode: 'none',
            state: 'ready',
            has_token: false,
            setup_instructions: 'Webshop ingestion requires no token.',
            capabilities: {
              search: true,
              stock_observation: true,
              pricing: true,
              health: true,
            },
            health: { status: 'unknown', last_checked_at: null },
            last_run: { status: 'never', finished_at: null },
          },
        ],
      },
    }).as('registry')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: { 'integration.ebay.enabled': 'true' } },
    }).as('settings')

    signIn()
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.contains('h1', 'Integrations').should('be.visible')
    cy.contains('Configure providers, credentials, and connector actions.').should(
      'be.visible'
    )
    cy.contains('integrations.title').should('not.exist')
    cy.contains('integrations.description').should('not.exist')

    cy.get('[data-testid="provider-card-ebay"]').should('be.visible')
    cy.get('[data-testid="provider-card-au-webshop-bonzaslotcars-com-au"]').should(
      'be.visible'
    )

    cy.get('input[placeholder="Filter providers..."]').clear().type('bonza')
    cy.get('[data-testid="provider-card-au-webshop-bonzaslotcars-com-au"]').should(
      'be.visible'
    )
    cy.get('[data-testid="provider-card-ebay"]').should('not.exist')

    cy.contains('button', 'Rows').click()
    cy.location('search').should('contain', 'view=rows')
    cy.get('table').should('be.visible')
  })

  it('UI-SCREEN-INTEGRATIONS-002 + UI-SCREEN-INTEGRATIONS-007 + INTEGRATION-020: opens provider detail panel with actions and status', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    })
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
            setup_instructions: 'Configure eBay token and marketplace.',
            capabilities: {
              search: true,
              stock_observation: false,
              pricing: true,
              health: true,
            },
            health: { status: 'ok', last_checked_at: '2026-03-01T00:00:00Z' },
            last_run: { status: 'success', finished_at: '2026-03-01T00:00:00Z' },
          },
        ],
      },
    })
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: { 'integration.ebay.enabled': 'true' } },
    })

    signIn()

    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.contains('Manage provider credentials, validation, and setup controls.').should(
      'be.visible'
    )
    cy.contains('Configure eBay token and marketplace.').should('be.visible')
    cy.contains('button', 'Validate').should('be.visible')
    cy.contains('button', 'Sync').should('be.disabled')
    cy.contains('Sync runs from Market Watch query sets.').should('be.visible')
    cy.contains('button', 'Save Integration').should('be.visible')
    cy.contains('Mode: official_api').should('be.visible')
    cy.contains('Health: ok').should('be.visible')
    cy.contains('Last run: success').should('be.visible')
  })

  it('UI-SCREEN-INTEGRATIONS-003 + UI-SCREEN-INTEGRATIONS-004 + UI-SCREEN-INTEGRATIONS-008: persists settings and reconciles validation health state', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    })
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
            setup_instructions: 'Configure eBay token and marketplace.',
            capabilities: {
              search: true,
              stock_observation: false,
              pricing: true,
              health: true,
            },
            health: { status: 'unknown', last_checked_at: null },
            last_run: { status: 'never', finished_at: null },
          },
        ],
      },
    })
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: {
        settings: {
          ebay_base_url: 'https://api.ebay.com',
          ebay_marketplace: 'EBAY-AU',
          'integration.ebay.enabled': 'true',
        },
      },
    })
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      delayMs: 250,
      body: {
        provider: 'ebay',
        status: 'ok',
        message: 'healthy',
        updated_at: '2026-03-01T00:01:00Z',
      },
    }).as('validate')
    cy.intercept('PUT', '/api/profiles/*/settings', (req) => {
      expect(req.body.settings).to.have.property('ebay_base_url')
      expect(req.body.settings).to.have.property('ebay_marketplace')
      expect(req.body.settings).to.have.property('ebay_bearer_token', 'new-secret-token')
      req.reply({
        statusCode: 200,
        body: {
          settings: {
            ...req.body.settings,
            'integration.ebay.enabled': 'true',
          },
        },
      })
    }).as('saveSettings')

    signIn()

    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.contains('Token on file.').should('be.visible')
    cy.contains('input[placeholder="New token / API key"]').should('not.exist')
    cy.get('[data-testid="replace-token"]').click()
    cy.get('[data-testid="provider-token-input"]').type('new-secret-token')
    cy.contains('Health: unknown').should('be.visible')
    cy.contains('Last run: never').should('be.visible')
    cy.contains('Last checked: n/a').should('be.visible')
    cy.contains('button', 'Validate').click()
    cy.contains('button', 'Validating...').should('be.visible')
    cy.wait('@validate')
    cy.contains('Validated eBay health: ok.').should('be.visible')
    cy.contains('Health: ok').should('be.visible')
    cy.contains('Last run: success').should('be.visible')
    cy.contains('Last checked: 2026-03-01T00:01:00Z').should('be.visible')
    cy.contains('button', 'Cancel').click()
    cy.contains('[data-testid="provider-card-ebay"]', 'Health: ok').should('be.visible')
    cy.contains('[data-testid="provider-card-ebay"]', 'Last run: success').should('be.visible')
    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.get('[data-testid="replace-token"]').click()
    cy.get('[data-testid="provider-token-input"]').type('new-secret-token')
    cy.contains('button', 'Save Integration').click()
    cy.wait('@saveSettings')
    cy.contains('Provider configuration saved.').should('be.visible')
  })

  it('UI-SCREEN-INTEGRATIONS-005: renders deterministic bootstrap error with retry control', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    })
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 500,
      body: { error: 'provider_registry_unavailable' },
    }).as('registryFail')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    })

    signIn()

    cy.wait('@registryFail')
    cy.get('[data-testid="integrations-bootstrap-error"]').should('be.visible')
    cy.contains('button', 'Retry').should('be.visible')
  })

  it('UI-SCREEN-INTEGRATIONS-005: recovers active-profile bootstrap inline by selecting or creating profile context', () => {
    let activeProfileRecovered = false

    cy.intercept('GET', '/api/profiles/active', (req) => {
      if (!activeProfileRecovered) {
        req.reply({ statusCode: 404, body: { error: 'active_profile_not_set' } })
        return
      }
      req.reply({ statusCode: 200, body: { id: 'profile-e2e-002', name: 'Recovered Profile' } })
    }).as('activeProfile')

    cy.intercept('GET', '/api/profiles', {
      statusCode: 200,
      body: {
        profiles: [{ id: 'profile-e2e-002', name: 'Recovered Profile' }],
      },
    }).as('profilesList')

    cy.intercept('PUT', '/api/profiles/active', (req) => {
      expect(req.body.profile_id).to.eq('profile-e2e-002')
      activeProfileRecovered = true
      req.reply({ statusCode: 200, body: { id: 'profile-e2e-002', name: 'Recovered Profile' } })
    }).as('setActiveProfile')

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
            has_token: false,
            setup_instructions: 'Configure eBay token and marketplace.',
            capabilities: {
              search: true,
              stock_observation: false,
              pricing: true,
              health: true,
            },
            health: { status: 'ok', last_checked_at: '2026-03-01T00:00:00Z' },
            last_run: { status: 'success', finished_at: '2026-03-01T00:00:00Z' },
          },
        ],
      },
    }).as('registryRecovered')

    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    }).as('settingsRecovered')

    signIn()

    cy.wait('@profilesList')
    cy.get('[data-testid="integrations-profile-recovery"]').should('be.visible')
    cy.get('[data-testid="integrations-recovery-profile-profile-e2e-002"]').click()
    cy.wait('@setActiveProfile')
    cy.wait('@registryRecovered')
    cy.wait('@settingsRecovered')
    cy.get('[data-testid="provider-card-ebay"]').should('be.visible')
    cy.get('[data-testid="integrations-bootstrap-error"]').should('not.exist')
  })

  it('INTEGRATION-018 + INTEGRATION-019: runtime provider registry includes configured shop domains and capability classification fields', () => {
    cy.request('/api/providers/registry').then((response) => {
      expect(response.status).to.eq(200)
      const providers = response.body.providers as Array<Record<string, unknown>>
      expect(providers.length).to.be.greaterThan(0)

      const domains = providers
        .map((provider) => String(provider.base_domain ?? ''))
        .filter(Boolean)

      ;[
        'bonzaslotcars.com.au',
        'frontlinehobbies.com.au',
        'hobbytechtoys.com.au',
        'andrewshobbies.com.au',
        'voglers.com.au',
        'acercmodels.com',
        'mrtoys.com.au',
        'hobbyco.com.au',
        'metrohobbies.com.au',
      ].forEach((domain) => {
        expect(domains).to.include(domain)
      })

      providers.forEach((provider) => {
        expect(provider).to.have.property('integration_mode')
        expect(provider).to.have.property('api_available')
        expect(provider).to.have.property('auth_requirement')
      })
    })
  })

  it('UI-SCREEN-INTEGRATIONS-009 + UC-INT-UI-10: cards show provider API family badges from registry mapping', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    })
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'au-webshop-voglers-com-au',
            display_name: 'voglers.com.au',
            base_domain: 'voglers.com.au',
            integration_mode: 'storefront_access',
            api_family: 'bigcommerce',
            api_support_profile: 'bigcommerce_storefront_v1',
            auth_mode: 'none',
            state: 'ready',
            has_token: false,
            setup_instructions: 'Storefront mode by default.',
            capabilities: {
              search: true,
              stock_observation: true,
              pricing: true,
              health: true,
            },
            health: { status: 'ok', last_checked_at: '2026-03-01T00:00:00Z' },
            last_run: { status: 'success', finished_at: '2026-03-01T00:00:00Z' },
          },
        ],
      },
    })
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    })

    signIn()
    cy.get('[data-testid="provider-card-au-webshop-voglers-com-au"]').should('be.visible')
    cy.get('[data-testid="provider-api-family-au-webshop-voglers-com-au"]')
      .should('be.visible')
      .and('contain.text', 'API Family: bigcommerce')
  })

  it('UI-SCREEN-INTEGRATIONS-009 + UC-INT-UI-11 + INTEGRATION-024: detail panel shows API family + support profile metadata from registry', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    })
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'au-webshop-voglers-com-au',
            display_name: 'voglers.com.au',
            base_domain: 'voglers.com.au',
            integration_mode: 'storefront_access',
            api_family: 'bigcommerce',
            api_support_profile: 'bigcommerce_storefront_v1',
            auth_mode: 'none',
            state: 'ready',
            has_token: false,
            setup_instructions: 'Storefront mode by default.',
            capabilities: {
              search: true,
              stock_observation: true,
              pricing: true,
              health: true,
            },
            health: { status: 'ok', last_checked_at: '2026-03-01T00:00:00Z' },
            last_run: { status: 'success', finished_at: '2026-03-01T00:00:00Z' },
          },
        ],
      },
    })
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: {} },
    })

    signIn()
    cy.get('[data-testid="provider-open-au-webshop-voglers-com-au"]').click()
    cy.get('[data-testid="provider-detail-api-family"]')
      .should('be.visible')
      .and('contain.text', 'API Family: bigcommerce')
    cy.get('[data-testid="provider-detail-api-support-profile"]')
      .should('be.visible')
      .and('contain.text', 'Support Profile: bigcommerce_storefront_v1')
  })
})
