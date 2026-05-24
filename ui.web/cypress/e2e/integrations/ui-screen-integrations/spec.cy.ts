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
      'exist'
    )
    cy.contains('Configure eBay token and marketplace.').should('exist')
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
    cy.contains('[data-testid="provider-card-ebay"]', 'Health: ok').should('be.visible')
    cy.contains('[data-testid="provider-card-ebay"]', 'Last run: success').should('be.visible')
    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.get('[data-testid="replace-token"]').click()
    cy.get('[data-testid="provider-token-input"]').type('new-secret-token')
    cy.contains('button', 'Save Integration').click()
    cy.wait('@saveSettings')
    cy.contains('Provider configuration saved.').scrollIntoView().should('be.visible')
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
