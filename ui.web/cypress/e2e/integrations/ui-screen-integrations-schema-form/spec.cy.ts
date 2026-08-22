describe('ui-screen-integrations-schema-form', () => {
  function signIn() {
    cy.stubLocalServerSession('profile-e2e-001')
    cy.visit('/sign-in?redirect=%2Fintegrations%2F')
    cy.contains('button', 'Open local workspace').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/integrations\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('#1464 + UI-SCREEN-INTEGRATIONS-015: renders schema-driven setup fields for unconfigured registry providers', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-e2e-001', name: 'E2E Local' },
    }).as('activeProfile')
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
            setup_instructions:
              'Use public BigCommerce storefront data before adding a token for deeper stock fields.',
            setup_schema: {
              schema_ref: 'integrations/au-webshop/setup',
              persistence_scope: 'active_profile',
              submit_target: '/api/profiles/:profileId/settings',
              validate_action: 'provider.family_detect',
              fields: [
                {
                  key: 'base_domain',
                  label: 'Store domain',
                  type: 'text',
                  required: true,
                  read_only: true,
                  persistence: 'provider_manifest',
                },
                {
                  key: 'provider_family',
                  label: 'Provider family',
                  type: 'select',
                  required: false,
                  persistence: 'profile_settings',
                  options: [
                    { value: 'auto', label: 'Auto-detect' },
                    { value: 'bigcommerce', label: 'BigCommerce' },
                  ],
                },
                {
                  key: 'crawl_interval_minutes',
                  label: 'Polling interval',
                  type: 'number',
                  required: false,
                  persistence: 'profile_settings',
                  default: 1440,
                },
              ],
            },
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
          'integration.au-webshop-voglers-com-au.enabled': 'false',
        },
      },
    }).as('settings')

    signIn()
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')

    cy.get('[data-testid="provider-row-au-webshop-voglers-com-au"]').should(
      'not.exist'
    )
    cy.get('[data-testid="integrations-header-add"]').click()
    cy.get(
      '[data-testid="integrations-provider-selector-option-au-webshop-voglers-com-au"]'
    )
      .should('be.visible')
      .and('contain.text', 'voglers.com.au')
      .click()
    cy.get('[role="dialog"]')
      .should('contain.text', 'voglers.com.au')
      .and(
        'contain.text',
        'Use public BigCommerce storefront data before adding a token'
      )
    cy.get('[data-testid="provider-detail-api-family"]')
      .should('be.visible')
      .and('contain.text', 'API Family: bigcommerce')
    cy.get('[data-testid="provider-detail-api-support-profile"]')
      .should('be.visible')
      .and('contain.text', 'Support Profile: bigcommerce_storefront_v1')
    cy.get('[data-testid="integration-schema-form"]')
      .should('be.visible')
      .and('contain.text', 'Store domain')
      .and('contain.text', 'Provider family')
      .and('contain.text', 'Polling interval')
    cy.get('[data-testid="provider-schema-field-base_domain"]')
      .should('have.value', 'voglers.com.au')
      .and('have.attr', 'readonly')
    cy.get('[data-testid="provider-schema-field-crawl_interval_minutes"]')
      .should('have.value', '1440')
  })
})
