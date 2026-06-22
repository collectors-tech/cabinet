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

  it('UI-SCREEN-INTEGRATIONS-001 + UI-SCREEN-INTEGRATIONS-006 + INTEGRATION-022: defaults to configured-only table and supports filter/sort/view using registry data', () => {
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

    cy.get('[data-testid="integrations-header-title"]').should(
      'contain',
      'Integrations'
    )
    cy.contains('Configure providers, credentials, and connector actions.').should(
      'not.exist'
    )
    cy.contains('integrations.title').should('not.exist')
    cy.contains('integrations.description').should('not.exist')

    cy.get('[data-testid="integrations-table-surface"]').should('be.visible')
    cy.contains('th', 'Provider').should('be.visible')
    cy.contains('th', 'Category / Type').should('exist')
    cy.contains('th', 'Connection').should('exist')
    cy.contains('th', 'Actions').should('exist')
    cy.contains('th', 'Health / Last run').should('exist')
    cy.contains('th', 'Row actions').should('exist')
    cy.get('[data-testid="provider-row-ebay"]').should('be.visible')
    cy.get('[data-testid="provider-row-au-webshop-bonzaslotcars-com-au"]').should(
      'not.exist'
    )
    cy.get('[data-testid="integrations-header-add"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Add integration')
      .and('not.contain.text', 'Add Integration')

    cy.get('input[placeholder="Filter providers..."]').clear().type('bonza')
    cy.get('[data-testid="provider-row-au-webshop-bonzaslotcars-com-au"]').should(
      'not.exist'
    )
    cy.get('[data-testid="provider-row-ebay"]').should('not.exist')

    cy.get('input[placeholder="Filter providers..."]').clear()
    cy.contains('button', 'Cards').click()
    cy.location('search').should('contain', 'view=cards')
    cy.get('[data-testid="provider-card-ebay"]').should('be.visible')
    cy.get('[data-testid="provider-card-au-webshop-bonzaslotcars-com-au"]').should(
      'not.exist'
    )
    cy.contains('button', 'Rows').click()
    cy.get('table').should('be.visible')
  })

  it('UI-SCREEN-INTEGRATIONS-015 + #1435: Add Integration is icon-only and opens provider selection first', () => {
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
            provider_id: 'au-webshop-acercmodels-com',
            display_name: 'acercmodels.com',
            base_domain: 'acercmodels.com',
            integration_mode: 'web_ingestion',
            auth_mode: 'none',
            state: 'ready',
            has_token: false,
            setup_instructions: 'Configure Acerc Models catalog ingestion.',
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

    cy.get('[data-testid="integrations-header-add"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Add integration')
      .and('not.contain.text', 'Add Integration')
      .click()

    cy.get('[data-testid="integrations-provider-selector"]')
      .should('be.visible')
      .and('contain', 'Add Integration')
      .and('contain', 'Choose a provider')
      .and('contain', 'acercmodels.com')
      .and('not.contain', 'Base URL')
    cy.get('[role="dialog"]').should('not.contain', 'Items per page')

    cy.get(
      '[data-testid="integrations-provider-selector-option-au-webshop-acercmodels-com"]'
    ).click()
    cy.get('[data-testid="integrations-provider-selector"]').should('not.exist')
    cy.get('[role="dialog"]')
      .should('contain', 'acercmodels.com')
      .and('contain', 'Base URL')
      .and('contain', 'Items per page')
  })

  it('UI-SCREEN-INTEGRATIONS-011 + #1112: paginates the full-page configured integrations table', () => {
    const providers = Array.from({ length: 12 }, (_, index) => {
      const number = index + 1
      return {
        provider_id: `provider-${String(number).padStart(2, '0')}`,
        display_name: `Provider ${String(number).padStart(2, '0')}`,
        base_domain: `provider-${number}.example.test`,
        integration_mode: number % 2 === 0 ? 'official_api' : 'web_ingestion',
        auth_mode: number % 2 === 0 ? 'api_key' : 'none',
        state: 'ready',
        has_token: number % 2 === 0,
        setup_instructions: 'Configure provider credentials.',
        capabilities: {
          search: true,
          stock_observation: number % 2 !== 0,
          pricing: true,
          health: true,
        },
        health: { status: number % 2 === 0 ? 'ok' : 'unknown' },
        last_run: { status: number % 2 === 0 ? 'success' : 'never' },
      }
    })

    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: { providers },
    }).as('registry')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: {
        settings: Object.fromEntries(
          providers.map((provider) => [
            `integration.${provider.provider_id}.enabled`,
            provider.provider_id === 'provider-12' ? 'false' : 'true',
          ])
        ),
      },
    }).as('settings')

    signIn()
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.get('[data-testid="integrations-table-pagination"]').should(
      'contain',
      'Showing 1-10 of 12 integrations'
    )
    cy.get('[data-testid="integrations-table-page-status"]').should(
      'contain',
      'Page 1 of 2'
    )
    cy.get('[data-testid="provider-row-provider-01"]').should('be.visible')
    cy.get('[data-testid="provider-row-provider-11"]').should('not.exist')

    cy.get('[data-testid="integrations-table-next-page"]').click()
    cy.get('[data-testid="integrations-table-pagination"]').should(
      'contain',
      'Showing 11-12 of 12 integrations'
    )
    cy.get('[data-testid="integrations-table-page-status"]').should(
      'contain',
      'Page 2 of 2'
    )
    cy.get('[data-testid="provider-row-provider-11"]').should('be.visible')
    cy.get('[data-testid="provider-row-provider-01"]').should('not.exist')

    cy.get('input[placeholder="Filter providers..."]').clear().type('12')
    cy.get('[data-testid="integrations-table-page-status"]').should(
      'contain',
      'Page 1 of 1'
    )
    cy.get('[data-testid="provider-row-provider-12"]').should('be.visible')
  })

  it('UI-SCREEN-INTEGRATIONS-011 + #1112: keeps controls stable while only table body scrolls', () => {
    const providers = Array.from({ length: 14 }, (_, index) => {
      const number = index + 1
      return {
        provider_id: `configured-${String(number).padStart(2, '0')}`,
        display_name: `Configured ${String(number).padStart(2, '0')}`,
        base_domain: `configured-${number}.example.test`,
        integration_mode: 'official_api',
        auth_mode: 'api_key',
        state: 'ready',
        has_token: true,
        setup_instructions: 'Configure provider credentials.',
        capabilities: {
          search: true,
          stock_observation: false,
          pricing: true,
          health: true,
        },
        health: { status: 'ok' },
        last_run: { status: 'success' },
      }
    })

    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          ...providers,
          {
            provider_id: 'catalog-only',
            display_name: 'Catalog Only',
            base_domain: 'catalog-only.example.test',
            integration_mode: 'web_ingestion',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: false,
            setup_instructions: 'Configure catalog provider.',
            capabilities: {
              search: true,
              stock_observation: true,
              pricing: true,
              health: true,
            },
            health: { status: 'unknown' },
            last_run: { status: 'never' },
          },
        ],
      },
    }).as('registry')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: {
        settings: Object.fromEntries(
          providers.map((provider) => [
            `integration.${provider.provider_id}.enabled`,
            'true',
          ])
        ),
      },
    }).as('settings')

    signIn()
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.get('[data-testid="integrations-header-add"]').should('be.visible')
    cy.get('[data-testid="provider-row-catalog-only"]').should('not.exist')
    cy.get('[data-testid="integrations-table-scroll-body"]')
      .should('have.css', 'overflow-y', 'auto')
      .and(($el) => {
        expect($el[0].clientHeight).to.be.lessThan($el[0].scrollHeight)
      })
    cy.get('[data-testid="integrations-table-surface"]').then(($surface) => {
      const surfaceTop = $surface[0].getBoundingClientRect().top
      cy.get('[data-testid="integrations-table-scroll-body"]').scrollTo('bottom')
      cy.get('[data-testid="integrations-table-pagination"]').should('be.visible')
      cy.get('[data-testid="integrations-table-surface"]').then(($after) => {
        expect($after[0].getBoundingClientRect().top).to.eq(surfaceTop)
      })
    })
  })

  it('UI-SCREEN-INTEGRATIONS-014 + UC-INT-UI-18: separates row details edit and action dialogs', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'connected-api',
            display_name: 'Connected API',
            base_domain: 'connected-api.example.test',
            integration_mode: 'official_api',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: true,
            setup_instructions: 'Configure connected API token.',
            capabilities: {
              search: true,
              stock_observation: false,
              pricing: true,
              health: true,
            },
            health: {
              status: 'ok',
              last_checked_at: '2026-03-01T00:00:00Z',
              next_action: 'retry_after_backoff',
            },
            last_run: { status: 'success', finished_at: '2026-03-01T00:00:00Z' },
          },
          {
            provider_id: 'offline-webshop',
            display_name: 'Offline Webshop',
            base_domain: 'offline-webshop.example.test',
            integration_mode: 'web_ingestion',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: false,
            setup_instructions: 'Configure offline webshop API credentials.',
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
      body: {
        settings: {
          'integration.connected-api.enabled': 'true',
          'integration.offline-webshop.enabled': 'false',
        },
      },
    }).as('settings')

    signIn()
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.get('[data-testid="provider-open-connected-api"]').click()
    cy.get('[role="dialog"]').should('contain', 'Connected API')
    cy.get('[data-testid="integrations-row-details-modal"]').should('not.exist')
    cy.get('[data-testid="integrations-row-edit-modal"]').should('not.exist')
    cy.location('search').should('not.contain', 'selected=connected-api')
    cy.get('body').type('{esc}')
    cy.get('[role="dialog"]').should('not.exist')

    cy.get('[data-testid="provider-row-connected-api"]').click()
    cy.get('[data-testid="integrations-row-details-modal"]', {
      timeout: 1000,
    })
      .should('be.visible')
      .and('contain', 'Connected API')
      .and('contain', 'State:')
      .and('contain', 'ready')
      .and('contain', 'Health:')
      .and('contain', 'ok')
    cy.location('search').should('contain', 'selected=connected-api')
    cy.get('[data-testid="integrations-row-edit-modal"]').should('not.exist')
    cy.get('body').type('{esc}')
    cy.get('[data-testid="integrations-row-details-modal"]').should('not.exist')

    cy.get('[data-testid="provider-row-offline-webshop"]').dblclick()
    cy.get('[data-testid="integrations-row-edit-modal"]')
      .should('be.visible')
      .and('contain', 'Offline Webshop')
      .and('contain', 'Mode:')
      .and('contain', 'web_ingestion')
    cy.location('search').should('contain', 'selected=offline-webshop')
    cy.get('[data-testid="integrations-row-details-modal"]').should('not.exist')
  })

  it('UI-SCREEN-INTEGRATIONS-012 + UC-INT-UI-08: filters rows by integration type selector', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'connected-api',
            display_name: 'Connected API',
            base_domain: 'connected-api.example.test',
            integration_mode: 'official_api',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: true,
            setup_instructions: 'Configure connected API token.',
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
            provider_id: 'offline-webshop',
            display_name: 'Offline Webshop',
            base_domain: 'offline-webshop.example.test',
            integration_mode: 'web_ingestion',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: false,
            setup_instructions: 'Configure offline webshop API credentials.',
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
      body: {
        settings: {
          'integration.connected-api.enabled': 'true',
          'integration.offline-webshop.enabled': 'false',
        },
      },
    }).as('settings')

    signIn()
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.get('[data-testid="provider-row-connected-api"]').should('be.visible')
    cy.get('[data-testid="provider-row-offline-webshop"]').should('be.visible')

    cy.contains('button', 'All Integrations').click()
    cy.get('[role="option"]').contains('Connected').click()
    cy.location('search').should('contain', 'type=connected')
    cy.contains('button', 'Connected').should('be.visible')
    cy.get('[data-testid="provider-row-connected-api"]').should('be.visible')
    cy.get('[data-testid="provider-row-offline-webshop"]').should('not.exist')

    cy.contains('button', 'Connected').click()
    cy.get('[role="option"]').contains('Not Connected').click()
    cy.location('search').should('contain', 'type=notConnected')
    cy.contains('button', 'Not Connected').should('be.visible')
    cy.get('[data-testid="provider-row-offline-webshop"]').should('be.visible')
    cy.get('[data-testid="provider-row-connected-api"]').should('not.exist')

    cy.contains('button', 'Not Connected').click()
    cy.get('[role="option"]').contains('All Integrations').click()
    cy.location('search').should('not.contain', 'type=')
    cy.get('[data-testid="provider-row-connected-api"]').should('be.visible')
    cy.get('[data-testid="provider-row-offline-webshop"]').should('be.visible')
  })

  it('UI-SCREEN-INTEGRATIONS-012 + UC-INT-UI-09: toggles rows and cards while preserving active filter context', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'connected-api',
            display_name: 'Connected API',
            base_domain: 'connected-api.example.test',
            integration_mode: 'official_api',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: true,
            setup_instructions: 'Configure connected API token.',
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
            provider_id: 'offline-webshop',
            display_name: 'Offline Webshop',
            base_domain: 'offline-webshop.example.test',
            integration_mode: 'web_ingestion',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: false,
            setup_instructions: 'Configure offline webshop API credentials.',
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
      body: {
        settings: {
          'integration.connected-api.enabled': 'true',
          'integration.offline-webshop.enabled': 'false',
        },
      },
    }).as('settings')

    signIn()
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.get('input[placeholder="Filter providers..."]').clear().type('api')
    cy.contains('button', 'All Integrations').click()
    cy.get('[role="option"]').contains('Connected').click()
    cy.location('search').should('contain', 'filter=api')
    cy.location('search').should('contain', 'type=connected')
    cy.contains('button', 'Rows').should('have.attr', 'aria-pressed', 'true')
    cy.get('[data-testid="integrations-table-surface"]').should('be.visible')
    cy.get('[data-testid="provider-row-connected-api"]').should('be.visible')
    cy.get('[data-testid="provider-row-offline-webshop"]').should('not.exist')

    cy.contains('button', 'Cards').click()
    cy.location('search').should('contain', 'filter=api')
    cy.location('search').should('contain', 'type=connected')
    cy.location('search').should('contain', 'view=cards')
    cy.contains('button', 'Cards').should('have.attr', 'aria-pressed', 'true')
    cy.get('[data-testid="integrations-table-surface"]').should('not.exist')
    cy.get('[data-testid="provider-card-connected-api"]').should('be.visible')
    cy.get('[data-testid="provider-card-offline-webshop"]').should('not.exist')

    cy.contains('button', 'Rows').click()
    cy.location('search').should('contain', 'filter=api')
    cy.location('search').should('contain', 'type=connected')
    cy.location('search').should('not.contain', 'view=')
    cy.contains('button', 'Rows').should('have.attr', 'aria-pressed', 'true')
    cy.get('[data-testid="integrations-table-surface"]').should('be.visible')
    cy.get('[data-testid="provider-row-connected-api"]').should('be.visible')
    cy.get('[data-testid="provider-row-offline-webshop"]').should('not.exist')
  })

  it('UI-SCREEN-INTEGRATIONS-013 + UC-INT-UI-17: applies direct route query state on first render', () => {
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
          {
            provider_id: 'au-webshop-hobbyco-com-au',
            display_name: 'hobbyco.com.au',
            base_domain: 'hobbyco.com.au',
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
      body: { settings: { 'integration.au-webshop-bonzaslotcars-com-au.enabled': 'true' } },
    }).as('settings')

    cy.visit(
      '/sign-in?redirect=%2Fintegrations%2F%3Ffilter%3Dbonza%26type%3Dconnected%26sort%3Ddesc%26view%3Drows'
    )
    cy.get('input[name="email"]').clear().type('e2e-inventory@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/integrations\/?$/
    )
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.location('search').should('contain', 'filter=bonza')
    cy.location('search').should('contain', 'type=connected')
    cy.location('search').should('contain', 'sort=desc')
    cy.location('search').should('contain', 'view=rows')
    cy.get('input[placeholder="Filter providers..."]').should('have.value', 'bonza')
    cy.contains('button', 'Connected').should('be.visible')
    cy.contains('button', 'Rows').should('have.attr', 'aria-pressed', 'true')
    cy.get('table').should('be.visible')
    cy.contains('td', 'bonzaslotcars.com.au').should('be.visible')
    cy.contains('td', 'eBay').should('not.exist')
    cy.contains('td', 'hobbyco.com.au').should('not.exist')
    cy.get('[data-testid="provider-card-au-webshop-bonzaslotcars-com-au"]').should(
      'not.exist'
    )
  })

  it('UI-SCREEN-INTEGRATIONS-013 + UC-INT-UI-19: shows deterministic empty state for direct route filters', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-empty-filter', name: 'E2E Empty Filter' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'connected-api',
            display_name: 'Connected API',
            base_domain: 'connected-api.example.test',
            integration_mode: 'official_api',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: true,
            setup_instructions: 'Configure connected API token.',
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
            provider_id: 'offline-webshop',
            display_name: 'Offline Webshop',
            base_domain: 'offline-webshop.example.test',
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
      body: { settings: { 'integration.connected-api.enabled': 'true' } },
    }).as('settings')

    cy.visit(
      '/sign-in?redirect=%2Fintegrations%2F%3Ffilter%3Darcade%26type%3Dconnected%26sort%3Dasc%26view%3Drows'
    )
    cy.get('input[name="email"]').clear().type('e2e-inventory@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/integrations\/?$/
    )
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.location('search').should('contain', 'filter=arcade')
    cy.location('search').should('contain', 'type=connected')
    cy.location('search').should('contain', 'sort=asc')
    cy.location('search').should('contain', 'view=rows')
    cy.get('input[placeholder="Filter providers..."]').should('have.value', 'arcade')
    cy.contains('button', 'Connected').should('be.visible')
    cy.contains('button', 'Rows').should('have.attr', 'aria-pressed', 'true')
    cy.get('[data-testid="integrations-table-surface"]')
      .should('be.visible')
      .and('contain', 'No configured integrations match current filters.')
    cy.get('[data-testid="integrations-table-pagination"]').should(
      'contain',
      'Showing 0-0 of 0 integrations'
    )
    cy.get('[data-testid="integrations-table-page-status"]').should(
      'contain',
      'Page 1 of 1'
    )
    cy.get('[data-testid^="provider-row-"]').should('not.exist')
    cy.get('[data-testid^="provider-card-"]').should('not.exist')
    cy.get('[role="dialog"]').should('not.exist')
  })

  it('UI-SCREEN-INTEGRATIONS-013 + UC-INT-UI-20: shows deterministic empty state for direct route filters in cards view', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-empty-cards', name: 'E2E Empty Cards' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'connected-api',
            display_name: 'Connected API',
            base_domain: 'connected-api.example.test',
            integration_mode: 'official_api',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: true,
            setup_instructions: 'Configure connected API token.',
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
            provider_id: 'offline-webshop',
            display_name: 'Offline Webshop',
            base_domain: 'offline-webshop.example.test',
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
      body: { settings: { 'integration.connected-api.enabled': 'true' } },
    }).as('settings')

    cy.visit(
      '/sign-in?redirect=%2Fintegrations%2F%3Ffilter%3Darcade%26type%3Dconnected%26sort%3Dasc%26view%3Dcards'
    )
    cy.get('input[name="email"]').clear().type('e2e-inventory@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/integrations\/?$/
    )
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.location('search').should('contain', 'filter=arcade')
    cy.location('search').should('contain', 'type=connected')
    cy.location('search').should('contain', 'sort=asc')
    cy.location('search').should('contain', 'view=cards')
    cy.get('input[placeholder="Filter providers..."]').should('have.value', 'arcade')
    cy.contains('button', 'Connected').should('be.visible')
    cy.contains('button', 'Cards').should('have.attr', 'aria-pressed', 'true')
    cy.get('[data-testid="integrations-cards-empty-state"]')
      .should('be.visible')
      .and('contain', 'No configured integrations match current filters.')
    cy.get('[data-testid="integrations-table-surface"]').should('not.exist')
    cy.get('[data-testid^="provider-row-"]').should('not.exist')
    cy.get('[data-testid^="provider-card-"]').should('not.exist')
    cy.get('[role="dialog"]').should('not.exist')
  })

  it('TELEGRAM-CATALOG-CAPTURE-025: exposes Telegram assistant capture channel status from profile authorization settings', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'telegram',
            display_name: 'Telegram',
            base_domain: 'telegram.org',
            api_family: 'messaging_channel',
            api_support_profile: 'bot_webhook_sender_chat_v1',
            active_mode: 'authorized_sender_chat',
            integration_mode: 'assistant_capture_channel',
            auth_mode: 'sender_chat',
            state: 'ready',
            has_token: false,
            setup_instructions:
              'Configure Telegram sender/chat authorization in Profile settings, then route bot messages through the governed preview-before-apply capture channel.',
            auth_methods: {
              sender_chat: {
                state: 'connected',
                connected: true,
                credential_present: true,
                setup_message:
                  'Store the Telegram sender id and chat id on the active profile before Cabinet accepts capture messages.',
              },
            },
            capabilities: {
              search: false,
              stock_observation: false,
              pricing: false,
              health: true,
              assistant: true,
              media_capture: true,
              text_capture: true,
            },
            health: { status: 'unknown', last_checked_at: null },
            last_run: { status: 'never', finished_at: null },
          },
        ],
      },
    }).as('registry')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: {
        settings: {
          'telegram.catalog_capture.sender_id': '12345',
          'telegram.catalog_capture.chat_id': '-5235769556',
        },
      },
    }).as('settings')

    signIn()
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.contains('button', 'Cards').click()
    cy.get('[data-testid="provider-card-telegram"]')
      .should('be.visible')
      .and('contain', 'Telegram')
      .and('contain', 'API Family: messaging_channel')
      .and('contain', 'Assistant')
      .and('contain', 'Media capture')

    cy.get('[data-testid="provider-open-telegram"]').should('contain', 'Edit')

    cy.get('[data-testid="provider-open-telegram"]').click()
    cy.contains('Manage provider credentials, validation, and setup controls.').should(
      'be.visible'
    )
    cy.contains('Mode: assistant_capture_channel').should('be.visible')
    cy.contains('Auth method: sender/chat authorization').should('be.visible')
    cy.contains('Sender/chat state: connected').should('be.visible')
    cy.contains('Profile settings: sender and chat authorized').should('be.visible')
    cy.contains('preview-before-apply capture channel').should('be.visible')
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
            setup_status: {
              auth_mode: 'api_key',
              marketplace: 'EBAY_AU',
              token_state: 'stored',
              validation_status: 'degraded',
              health_state: 'degraded',
              next_action: 'check_provider_health_and_credentials',
              base_url_set: true,
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
      'exist'
    )
    cy.contains('Configure eBay token and marketplace.').should('exist')
    cy.get('[data-testid="ebay-setup-status-panel"]')
      .scrollIntoView()
      .should('contain', 'eBay setup status')
      .and('contain', 'Auth mode: api_key')
      .and('contain', 'Marketplace / Region: EBAY_AU')
      .and('contain', 'Token state: stored token on file')
      .and('contain', 'Validation status: degraded')
      .and('contain', 'Health state: degraded')
      .and('contain', 'Base URL override configured')
      .and('contain', 'Check provider health and credentials')
    cy.contains('button', 'Validate').scrollIntoView().should('be.visible')
    cy.contains('button', 'Sync').scrollIntoView().should('be.disabled')
    cy.contains('Sync runs from Market Watch query sets.')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('button', 'Save Integration').scrollIntoView().should('be.visible')
    cy.contains('Mode: official_api').should('exist')
    cy.contains('Health: ok').should('exist')
    cy.contains('Last run: success').should('exist')
  })

  it('INTEGRATION-025 + INTEGRATION-026: previews and imports eBay buyer-interest sync without remote write-back claims', () => {
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
    cy.intercept('POST', '/api/providers/ebay/buyer-interest/preview', (req) => {
      expect(req.body.source_account).to.eq('buyer@example.test')
      expect(req.body.items).to.have.length(2)
      req.reply({
        statusCode: 200,
        body: {
          provider: 'ebay',
          mode: 'preview',
          total: 2,
          counts: { wishlist: 1, discovery: 1 },
          mappings: [
            {
              title: 'Watched eBay listing',
              listing_id: 'v1|watch|0',
              interest_state: 'watched',
              destination: 'wishlist',
              write_back_allowed: false,
              write_back_blocker: 'ebay_api_capability_not_verified',
            },
            {
              title: 'Cart eBay listing',
              listing_id: 'v1|cart|0',
              interest_state: 'cart_like',
              destination: 'discovery',
              write_back_allowed: false,
              write_back_blocker: 'ebay_api_capability_not_verified',
            },
          ],
        },
      })
    }).as('buyerInterestPreview')
    cy.intercept('POST', '/api/providers/ebay/buyer-interest/import', (req) => {
      expect(req.body.source_account).to.eq('buyer@example.test')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'ebay',
          mode: 'import',
          total: 2,
          counts: { wishlist: 1, discovery: 1 },
          results: [
            {
              title: 'Watched eBay listing',
              listing_id: 'v1|watch|0',
              interest_state: 'watched',
              destination: 'wishlist',
              persisted_id: 'wishlist-ebay-watch',
              item_id: 'item-ebay-watch',
              write_back_allowed: false,
              write_back_blocker: 'ebay_api_capability_not_verified',
            },
            {
              title: 'Cart eBay listing',
              listing_id: 'v1|cart|0',
              interest_state: 'cart_like',
              destination: 'discovery',
              persisted_id: 'candidate-ebay-cart',
              candidate_id: 'candidate-ebay-cart',
              write_back_allowed: false,
              write_back_blocker: 'ebay_api_capability_not_verified',
            },
          ],
        },
      })
    }).as('buyerInterestImport')

    signIn()

    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.get('[data-testid="ebay-buyer-interest-sync-panel"]')
      .scrollIntoView()
      .should('be.visible')
    cy.get('[data-testid="ebay-buyer-interest-sync-panel"] summary').click()
    cy.get('[data-testid="ebay-buyer-interest-writeback-status"]').should(
      'contain',
      'Write-back blocked until eBay capability is verified'
    )
    cy.get('[data-testid="ebay-buyer-interest-payload"]').should(
      'contain.value',
      'buyer@example.test'
    )

    cy.get('[data-testid="ebay-buyer-interest-preview"]')
      .scrollIntoView()
      .click()
    cy.wait('@buyerInterestPreview')
    cy.contains('Buyer-interest preview mapped without remote write-back.').should(
      'be.visible'
    )
    cy.get('[data-testid="ebay-buyer-interest-result"]')
      .should('contain', 'Mode: preview / Total: 2')
      .and('contain', 'Wishlist: 1 / Discoveries: 1')
      .and('contain', 'Watched eBay listing -> wishlist')
      .and('contain', 'ebay_api_capability_not_verified')

    cy.get('[data-testid="ebay-buyer-interest-import"]')
      .scrollIntoView()
      .click()
    cy.wait('@buyerInterestImport')
    cy.contains(
      'Buyer-interest import persisted local Wishlist and Discovery state.'
    ).should('be.visible')
    cy.get('[data-testid="ebay-buyer-interest-result"]')
      .should('contain', 'Mode: import / Total: 2')
      .and('contain', 'Cart eBay listing -> discovery')
      .and('contain', 'ebay_api_capability_not_verified')
  })

  it('INTEGRATION-028: previews listing lifecycle commands and executes local drafts without remote write claims', () => {
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
            seller_operations: [],
          },
        ],
      },
    })
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: { 'integration.ebay.enabled': 'true' } },
    })
    cy.intercept('POST', '/api/providers/ebay/listing-lifecycle/preview', (req) => {
      if (req.body.command === 'draft') {
        expect(req.body.capability).to.eq('draft_only')
        expect(req.body.confirmed).to.eq(false)
        req.reply({
          statusCode: 200,
          body: {
            provider: 'ebay',
            mode: 'listing_lifecycle_preview',
            preview: {
              command: 'draft',
              capability: 'draft_only',
              allowed: true,
              local_only: true,
              remote_write: false,
              confirmation_required: false,
            },
          },
        })
        return
      }

      expect(req.body.command).to.eq('publish')
      expect(req.body.capability).to.eq('confirmed_api')
      expect(req.body.confirmed).to.eq(false)
      req.reply({
        statusCode: 200,
        body: {
          provider: 'ebay',
          mode: 'listing_lifecycle_preview',
          preview: {
            command: 'publish',
            capability: 'confirmed_api',
            allowed: false,
            local_only: false,
            remote_write: true,
            confirmation_required: true,
            blocker: 'ebay_listing_lifecycle_confirmation_required',
          },
        },
      })
    }).as('listingLifecyclePreview')
    cy.intercept('POST', '/api/providers/ebay/listing-lifecycle/execute', (req) => {
      if (req.body.command === 'draft') {
        expect(req.body.capability).to.eq('draft_only')
        req.reply({
          statusCode: 200,
          body: {
            provider: 'ebay',
            mode: 'listing_lifecycle_execute',
            execution: {
              command: 'draft',
              capability: 'draft_only',
              allowed: true,
              local_only: true,
              remote_write: false,
              executed: true,
              status: 'local_draft_ready',
              response: {
                provider: 'cabinet',
                command: 'draft',
                draft_id: 'draft-local-item-local-1',
                status: 'local_draft_ready',
              },
            },
          },
        })
        return
      }

      expect(req.body.command).to.eq('publish')
      expect(req.body.confirmed).to.eq(true)
      req.reply({
        statusCode: 409,
        body: {
          provider: 'ebay',
          mode: 'listing_lifecycle_execute',
          execution: {
            command: 'publish',
            capability: 'confirmed_api',
            allowed: false,
            local_only: false,
            remote_write: true,
            executed: false,
            status: 'blocked',
            blocker: 'ebay_listing_lifecycle_adapter_required',
          },
        },
      })
    }).as('listingLifecycleExecute')

    signIn()

    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.get('[data-testid="ebay-listing-lifecycle-panel"]')
      .scrollIntoView()
      .should('be.visible')
      .and('contain', 'Publish, revise, end, and relist require confirmation')

    cy.get('[data-testid="ebay-listing-lifecycle-preview-draft"]').click()
    cy.wait('@listingLifecyclePreview')
    cy.contains('Listing lifecycle preview completed without remote write.').should(
      'be.visible'
    )
    cy.get('[data-testid="ebay-listing-lifecycle-preview-result"]')
      .should('contain', 'Preview: Create draft')
      .and('contain', 'Allowed: yes / Local only: yes / Remote write: no')

    cy.get('[data-testid="ebay-listing-lifecycle-execute-draft"]').click()
    cy.wait('@listingLifecycleExecute')
    cy.contains('Listing draft was created locally without eBay remote write.').should(
      'be.visible'
    )
    cy.get('[data-testid="ebay-listing-lifecycle-execute-result"]')
      .should('contain', 'Execute: Create draft')
      .and('contain', 'Executed: yes / Local only: yes / Remote write: no')
      .and('contain', 'draft-local-item-local-1')

    cy.get('[data-testid="ebay-listing-lifecycle-preview-publish"]').click()
    cy.wait('@listingLifecyclePreview')
    cy.get('[data-testid="ebay-listing-lifecycle-preview-result"]')
      .should('contain', 'Preview: Publish')
      .and('contain', 'Allowed: no / Local only: no / Remote write: yes')
      .and('contain', 'ebay_listing_lifecycle_confirmation_required')

    cy.get('[data-testid="ebay-listing-lifecycle-confirm-execute-publish"]').click()
    cy.wait('@listingLifecycleExecute')
    cy.get('[data-testid="ebay-listing-lifecycle-error"]').should(
      'contain',
      'ebay_listing_lifecycle_adapter_required'
    )
    cy.get('[data-testid="ebay-listing-lifecycle-execute-result"]')
      .should('contain', 'Execute: Publish')
      .and('contain', 'Executed: no / Local only: no / Remote write: yes')
      .and('contain', 'blocked')
  })

  it('INTEGRATION-027 + #842: previews and executes seller operation sync without remote write claims', () => {
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
            seller_operations: [
              {
                operation: 'messages',
                capability: 'read_only',
                read_available: true,
                write_available: false,
                confirmation_required: false,
                blocker: 'ebay_seller_write_capability_not_verified',
              },
              {
                operation: 'offers',
                capability: 'confirmed_api',
                read_available: true,
                write_available: true,
                confirmation_required: true,
              },
              {
                operation: 'notifications',
                capability: 'unverified',
                read_available: false,
                write_available: false,
                confirmation_required: false,
                blocker: 'ebay_seller_operation_capability_not_verified',
              },
            ],
          },
        ],
      },
    })
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: { 'integration.ebay.enabled': 'true' } },
    })
    cy.intercept('POST', '/api/providers/ebay/seller-operations/preview', (req) => {
      if (req.body.operation === 'messages') {
        expect(req.body.action).to.eq('sync')
        expect(req.body.confirmed).to.eq(false)
        req.reply({
          statusCode: 200,
          body: {
            provider: 'ebay',
            mode: 'seller_operation_preview',
            preview: {
              operation: 'messages',
              action: 'sync',
              capability: 'read_only',
              read_available: true,
              write_available: false,
              confirmed: false,
              allowed: true,
              remote_write: false,
              blocker: 'ebay_seller_write_capability_not_verified',
            },
          },
        })
        return
      }

      expect(req.body.operation).to.eq('offers')
      expect(req.body.action).to.eq('fulfill')
      expect(req.body.confirmed).to.eq(true)
      req.reply({
        statusCode: 200,
        body: {
          provider: 'ebay',
          mode: 'seller_operation_preview',
          preview: {
            operation: 'offers',
            action: 'fulfill',
            capability: 'confirmed_api',
            read_available: true,
            write_available: true,
            confirmation_required: true,
            confirmed: true,
            allowed: true,
            remote_write: true,
          },
        },
      })
    }).as('sellerOperationPreview')
    cy.intercept('POST', '/api/providers/ebay/seller-operations/execute', (req) => {
      if (req.body.operation === 'messages') {
        expect(req.body.action).to.eq('sync')
        expect(req.body.confirmed).to.eq(false)
        req.reply({
          statusCode: 200,
          body: {
            provider: 'ebay',
            mode: 'seller_operation_execute',
            execution: {
              operation: 'messages',
              action: 'sync',
              capability: 'read_only',
              allowed: true,
              remote_write: false,
              executed: true,
              local_only: true,
              status: 'local_read_sync_complete',
              result: {
                source: 'local_read_model',
                records: [
                  {
                    id: 'msg-1',
                    title: 'Buyer question about condition',
                    kind: 'seller_message',
                    status: 'needs_reply',
                  },
                ],
                summary: { total: 1 },
              },
            },
          },
        })
        return
      }

      expect(req.body.operation).to.eq('offers')
      expect(req.body.action).to.eq('fulfill')
      expect(req.body.confirmed).to.eq(true)
      req.reply({
        statusCode: 409,
        body: {
          provider: 'ebay',
          mode: 'seller_operation_execute',
          execution: {
            operation: 'offers',
            action: 'fulfill',
            capability: 'confirmed_api',
            allowed: false,
            remote_write: true,
            executed: false,
            local_only: false,
            status: 'blocked',
            blocker: 'ebay_seller_operation_adapter_required',
          },
        },
      })
    }).as('sellerOperationExecute')

    signIn()

    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.get('[data-testid="ebay-seller-operations-panel"]')
      .scrollIntoView()
      .should('be.visible')
      .and('contain', 'External writes require confirmation')
    cy.get('[data-testid="ebay-seller-operation-notifications"]').should(
      'contain',
      'ebay_seller_operation_capability_not_verified'
    )
    cy.get('[data-testid="ebay-seller-operation-preview-notifications"]').should(
      'be.disabled'
    )

    cy.get('[data-testid="ebay-seller-operation-preview-messages"]').click()
    cy.wait('@sellerOperationPreview')
    cy.contains('Seller operation preview completed without remote write.').should(
      'be.visible'
    )
    cy.get('[data-testid="ebay-seller-operation-preview-result"]')
      .should('contain', 'Preview: Messages')
      .and('contain', 'Allowed: yes / Remote write: no')
      .and('contain', 'ebay_seller_write_capability_not_verified')

    cy.get('[data-testid="ebay-seller-operation-execute-messages"]').click()
    cy.wait('@sellerOperationExecute')
    cy.contains(
      'Seller operation read-only sync completed locally without remote write.'
    ).should('be.visible')
    cy.get('[data-testid="ebay-seller-operation-execute-result"]')
      .should('contain', 'Execute: Messages')
      .and('contain', 'Executed: yes / Local only: yes')
      .and('contain', 'local_read_sync_complete')
    cy.get('[data-testid="ebay-seller-operation-read-result"]')
      .should('contain', 'local_read_model')
      .and('contain', 'Buyer question about condition')
      .and('contain', 'seller_message / needs_reply')

    cy.get('[data-testid="ebay-seller-operation-confirm-offers"]').click()
    cy.wait('@sellerOperationPreview')
    cy.get('[data-testid="ebay-seller-operation-preview-result"]')
      .should('contain', 'Preview: Offers')
      .and('contain', 'Allowed: yes / Remote write: yes')

    cy.get('[data-testid="ebay-seller-operation-confirm-execute-offers"]').click()
    cy.wait('@sellerOperationExecute')
    cy.get('[data-testid="ebay-seller-operation-preview-error"]').should(
      'contain',
      'ebay_seller_operation_adapter_required'
    )
    cy.get('[data-testid="ebay-seller-operation-execute-result"]')
      .should('contain', 'Execute: Offers')
      .and('contain', 'Executed: no / Local only: no')
      .and('contain', 'blocked')
  })

  it('COMMERCE-LANDED-COST-001 + COMMERCE-LANDED-COST-003: previews landed-cost recommendations without mutation', () => {
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
            seller_operations: [],
          },
        ],
      },
    })
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: { 'integration.ebay.enabled': 'true' } },
    })
    cy.intercept('POST', '/api/commerce/landed-cost/plan', (req) => {
      expect(req.body.items).to.have.length(2)
      expect(req.body.components[0].allocation_method).to.eq('weight')
      req.reply({
        statusCode: 200,
        body: {
          provider: 'cabinet',
          mode: 'landed_cost_plan',
          mutable: false,
          allocation: {
            total_direct_cents: 46000,
            total_shared_cents: 9200,
            total_landed_cents: 55200,
            items: [
              {
                item_id: 'card-b',
                direct_cost_cents: 34500,
                allocated_cost_cents: 6600,
                landed_cost_cents: 41100,
                allocation_provenance_id: [
                  'forwarder-shipment:SHIP-1',
                  'forwarder-invoice:INV-1',
                ],
              },
              {
                item_id: 'card-a',
                direct_cost_cents: 11500,
                allocated_cost_cents: 2600,
                landed_cost_cents: 14100,
                allocation_provenance_id: [
                  'forwarder-shipment:SHIP-1',
                  'forwarder-invoice:INV-1',
                ],
              },
            ],
          },
          consolidation: {
            item_ids: ['card-a', 'card-b'],
            estimated_value_cents: 55200,
            estimated_fee_cents: 2500,
            estimated_total_cents: 57700,
            threshold_state: 'under_limit',
            warnings: [],
            mutable: false,
          },
        },
      })
    }).as('landedCostPlan')

    signIn()

    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.get('[data-testid="ebay-landed-cost-planner-panel"]')
      .scrollIntoView()
      .should('be.visible')
    cy.get('[data-testid="ebay-landed-cost-planner-panel"] summary').click()
    cy.get('[data-testid="ebay-landed-cost-mutation-status"]').should(
      'contain',
      'Preview only / no mutation'
    )
    cy.get('[data-testid="ebay-landed-cost-payload"]').should(
      'contain.value',
      'forwarder-shipment:SHIP-1'
    )
    cy.get('[data-testid="ebay-landed-cost-preview"]').click()
    cy.wait('@landedCostPlan')
    cy.contains(
      'Landed-cost plan previewed without mutating inventory or shipment state.'
    ).should('be.visible')
    cy.get('[data-testid="ebay-landed-cost-result"]')
      .should('contain', 'Mode: landed_cost_plan / Mutable: no')
      .and('contain', 'Direct: $460.00 / Shared: $92.00 / Landed: $552.00')
      .and('contain', 'card-b: landed $411.00')
      .and('contain', 'forwarder-shipment:SHIP-1')
      .and('contain', 'Consolidation: under_limit / Total: $577.00')
      .and('contain', 'Items: card-a, card-b')
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
      delay: 1000,
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
    cy.contains('Health: unknown').scrollIntoView().should('be.visible')
    cy.contains('Last run: never').scrollIntoView().should('be.visible')
    cy.contains('Last checked: n/a').scrollIntoView().should('be.visible')
    cy.contains('button', 'Validate').click()
    cy.contains('button', 'Validating...').should('be.visible')
    cy.wait('@validate')
    cy.contains('Validated eBay health: ok.').scrollIntoView().should('be.visible')
    cy.contains('Health: ok').scrollIntoView().should('be.visible')
    cy.contains('Last run: success').scrollIntoView().should('be.visible')
    cy.contains('Last checked: 2026-03-01T00:01:00Z')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('button', 'Cancel').click()
    cy.get('[role="dialog"]').should('not.exist')
    cy.contains('[data-testid="provider-row-ebay"]', 'Health: ok').should('be.visible')
    cy.contains('[data-testid="provider-row-ebay"]', 'Last run: success').should('be.visible')
    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.get('[data-testid="replace-token"]').click()
    cy.get('[data-testid="provider-token-input"]').type('new-secret-token')
    cy.contains('button', 'Save Integration').click()
    cy.wait('@saveSettings')
    cy.contains('Provider configuration saved.').scrollIntoView().should('be.visible')
  })

  it('INTEGRATION-006 + #1289: displays eBay provider-health readiness aliases and recovery guidance', () => {
    const healthResponses = [
      {
        provider: 'ebay',
        status: 'ok',
        state: 'ready',
        message: 'eBay credentials are ready for Market Watch runs.',
        last_error: null,
        retry_after_seconds: null,
        next_action: 'run_market_watch_query_sets',
        updated_at: '2026-06-16T06:31:00Z',
      },
      {
        provider: 'ebay',
        status: 'error',
        state: 'disabled',
        message: 'eBay bearer token is missing.',
        last_error: 'PROVIDER_AUTH_MISSING',
        retry_after_seconds: null,
        next_action: 'review_provider_credentials_and_health',
        updated_at: '2026-06-16T06:32:00Z',
      },
      {
        provider: 'ebay',
        status: 'error',
        state: 'degraded',
        message: 'eBay rejected the configured token.',
        last_error: 'PROVIDER_AUTH_INVALID',
        retry_after_seconds: null,
        next_action: 'review_provider_credentials_and_health',
        updated_at: '2026-06-16T06:33:00Z',
      },
      {
        provider: 'ebay',
        status: 'error',
        state: 'degraded',
        message: 'eBay rate limit backoff is active.',
        last_error: 'PROVIDER_RATE_LIMITED',
        retry_after_seconds: 120,
        next_action: 'retry_after_backoff',
        updated_at: '2026-06-16T06:34:00Z',
      },
    ]
    let healthIndex = 0

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
            state: 'disabled',
            has_token: false,
            setup_instructions: 'Configure eBay token and marketplace.',
            capabilities: {
              search: true,
              stock_observation: false,
              pricing: true,
              health: true,
            },
            health: {
              status: 'unknown',
              state: 'disabled',
              last_checked_at: null,
            },
            last_run: { status: 'never', finished_at: null },
          },
        ],
      },
    })
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: { ebay_marketplace: 'EBAY-AU' } },
    })
    cy.intercept('GET', '/api/provider/health?provider=ebay', (req) => {
      req.reply({
        statusCode: 200,
        body: healthResponses[Math.min(healthIndex, healthResponses.length - 1)],
      })
      healthIndex += 1
    }).as('providerHealth')

    signIn()

    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.contains('Readiness: disabled').scrollIntoView().should('be.visible')

    cy.contains('button', 'Validate').click()
    cy.wait('@providerHealth')
    cy.contains('Validated eBay health: ok.').scrollIntoView().should('be.visible')
    cy.contains('Readiness: ready').scrollIntoView().should('be.visible')
    cy.contains('Message: eBay credentials are ready for Market Watch runs.')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Next action: run_market_watch_query_sets')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Last checked: 2026-06-16T06:31:00Z')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Last error:').should('not.exist')

    cy.contains('button', 'Validate').click()
    cy.wait('@providerHealth')
    cy.contains('Readiness: disabled').scrollIntoView().should('be.visible')
    cy.contains('Health: error').scrollIntoView().should('be.visible')
    cy.contains('Message: eBay bearer token is missing.')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Last error: PROVIDER_AUTH_MISSING')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Next action: review_provider_credentials_and_health')
      .scrollIntoView()
      .should('be.visible')

    cy.contains('button', 'Validate').click()
    cy.wait('@providerHealth')
    cy.contains('Readiness: degraded').scrollIntoView().should('be.visible')
    cy.contains('Message: eBay rejected the configured token.')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Last error: PROVIDER_AUTH_INVALID')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Last run: failed').scrollIntoView().should('be.visible')

    cy.contains('button', 'Validate').click()
    cy.wait('@providerHealth')
    cy.contains('Message: eBay rate limit backoff is active.')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Last error: PROVIDER_RATE_LIMITED')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Retry after: 120 seconds').scrollIntoView().should('be.visible')
    cy.contains('Next action: retry_after_backoff')
      .scrollIntoView()
      .should('be.visible')
  })

  it('UI-SCREEN-INTEGRATIONS-010 + UC-INT-UI-12: labels provider config fields programmatically', () => {
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
            has_token: false,
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
    }).as('registry')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: {
        settings: {
          ebay_base_url: 'https://api.ebay.com',
          ebay_marketplace: 'EBAY-AU',
          'integration.ebay.items_per_page': '48',
        },
      },
    }).as('settings')

    signIn()
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.get('[role="dialog"]').within(() => {
      cy.contains('label', 'Base URL')
        .should('have.attr', 'for', 'provider-base-url')
        .then(($label) => {
          cy.get(`#${$label.attr('for')}`).should(
            'have.value',
            'https://api.ebay.com'
          )
        })
      cy.contains('label', 'Marketplace / Region')
        .should('have.attr', 'for', 'provider-marketplace')
        .then(($label) => {
          cy.get(`#${$label.attr('for')}`).should('have.value', 'EBAY-AU')
        })
      cy.contains('label', 'Items per page')
        .should('have.attr', 'for', 'provider-items-per-page')
        .then(($label) => {
          cy.get(`#${$label.attr('for')}`).should('have.value', '48')
        })
      cy.contains('label', 'New token / API key')
        .should('have.attr', 'for', 'provider-token')
        .then(($label) => {
          cy.get(`#${$label.attr('for')}`)
            .should('have.attr', 'type', 'password')
            .and('have.value', '')
        })
    })
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
      body: { settings: { 'integration.ebay.enabled': 'false' } },
    }).as('settingsRecovered')

    signIn()

    cy.wait('@profilesList')
    cy.get('[data-testid="integrations-profile-recovery"]').should('be.visible')
    cy.get('[data-testid="integrations-recovery-profile-profile-e2e-002"]').click()
    cy.wait('@setActiveProfile')
    cy.wait('@registryRecovered')
    cy.wait('@settingsRecovered')
    cy.get('[data-testid="provider-row-ebay"]').should('be.visible')
    cy.get('[data-testid="integrations-bootstrap-error"]').should('not.exist')
  })

  it('UI-SCREEN-INTEGRATIONS-005 + UC-INT-UI-16: creates a missing active profile inline and reloads integrations', () => {
    let activeProfileRecovered = false

    cy.intercept('GET', '/api/profiles/active', (req) => {
      if (!activeProfileRecovered) {
        req.reply({ statusCode: 404, body: { error: 'active_profile_not_set' } })
        return
      }
      req.reply({ statusCode: 200, body: { id: 'profile-e2e-created', name: 'Created Profile' } })
    }).as('activeProfile')

    cy.intercept('GET', '/api/profiles', {
      statusCode: 200,
      body: { profiles: [] },
    }).as('profilesList')

    cy.intercept('POST', '/api/profiles', (req) => {
      expect(req.body.name).to.eq('Created Profile')
      req.reply({
        statusCode: 200,
        body: { id: 'profile-e2e-created', name: 'Created Profile' },
      })
    }).as('createProfile')

    cy.intercept('PUT', '/api/profiles/active', (req) => {
      expect(req.body.profile_id).to.eq('profile-e2e-created')
      activeProfileRecovered = true
      req.reply({
        statusCode: 200,
        body: { id: 'profile-e2e-created', name: 'Created Profile' },
      })
    }).as('setActiveProfile')

    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'created-profile-provider',
            display_name: 'Created Profile Provider',
            base_domain: 'created-profile.example.test',
            integration_mode: 'official_api',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: false,
            setup_instructions: 'Configure created profile provider credentials.',
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

    cy.intercept('GET', '/api/profiles/profile-e2e-created/settings', {
      statusCode: 200,
      body: {
        settings: {
          'integration.created-profile-provider.enabled': 'false',
        },
      },
    }).as('settingsRecovered')

    signIn()

    cy.wait('@profilesList')
    cy.location('pathname').should('match', /^\/integrations\/?$/)
    cy.get('[data-testid="integrations-profile-recovery"]').should('be.visible')
    cy.get('[data-testid="integrations-recovery-no-profiles"]').should(
      'contain',
      'No selectable profiles were found'
    )
    cy.get('[data-testid="integrations-recovery-create-input"]').type(
      'Created Profile'
    )
    cy.get('[data-testid="integrations-recovery-create-submit"]').click()
    cy.wait('@createProfile')
    cy.wait('@setActiveProfile')
    cy.wait('@registryRecovered')
    cy.wait('@settingsRecovered')
    cy.location('pathname').should('match', /^\/integrations\/?$/)
    cy.get('[data-testid="provider-row-created-profile-provider"]').should(
      'be.visible'
    )
    cy.get('[data-testid="integrations-bootstrap-error"]').should('not.exist')
    cy.get('[data-testid="integrations-profile-recovery"]').should('not.exist')
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
      body: {
        settings: {
          'integration.au-webshop-voglers-com-au.enabled': 'true',
        },
      },
    })

    signIn()
    cy.contains('button', 'Cards').click()
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
      body: {
        settings: {
          'integration.au-webshop-voglers-com-au.enabled': 'true',
        },
      },
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
