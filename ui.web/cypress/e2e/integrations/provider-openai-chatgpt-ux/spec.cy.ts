describe('integrations/provider-openai-chatgpt-ux', () => {
  beforeEach(() => {
    cy.e2eReset()
    cy.e2eBootstrap().then((bootstrap) => {
      cy.e2eSetSetupState('present')
      cy.useBootstrappedProfile(bootstrap.profile_id, bootstrap.profile_name, {
        path: '/integrations',
      })
    })

    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'e2e-profile-001', name: 'E2E Local' },
    }).as('activeProfile')

    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'openai',
            display_name: 'OpenAI / ChatGPT',
            base_domain: 'platform.openai.com',
            api_family: 'ai_provider',
            api_support_profile: 'browser_auth_or_api_key',
            integration_mode: 'assistant_workflows',
            auth_mode: 'hybrid',
            state: 'needs_config',
            has_token: false,
            active_auth_method: '',
            auth_methods: {
              api_key: { state: 'setup_needed', connected: false, credential_present: false },
              browser_auth: {
                state: 'setup_needed',
                connected: false,
                credential_present: false,
                setup_message: 'Browser Auth requires a verifiable callback/artifact before Cabinet marks OpenAI connected.',
              },
            },
            model_options: ['gpt-4o-mini', 'gpt-4.1-mini', 'gpt-5.3-codex'],
            capabilities: {
              search: false,
              stock_observation: false,
              pricing: false,
              health: true,
              assistant: true,
              image_help: true,
              content_generation: true,
            },
            health: { status: 'needs_config' },
            last_run: { status: 'not_run' },
            setup_instructions: 'Configure OpenAI with Browser Auth or an API key.',
          },
        ],
      },
    }).as('providersRegistry')

    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: {
        settings: {
          assistant_default_provider: 'openai',
          assistant_default_model: 'gpt-4o-mini',
          'openai.browser_auth_state': 'setup_needed',
        },
      },
    }).as('profileSettings')

    cy.intercept('PUT', '/api/profiles/*/settings', (req) => {
      req.reply({ statusCode: 200, body: { settings: req.body.settings } })
    }).as('saveSettings')

    cy.intercept('PUT', '/api/profiles/*/secrets', (req) => {
      expect(req.body).to.deep.equal({ key: 'openai_api_key', value: 'sk-test-openai' })
      req.reply({ statusCode: 200, body: { ok: true } })
    }).as('saveSecret')

    cy.intercept('GET', '/api/provider/health*', {
      statusCode: 200,
      body: { status: 'ok', message: 'ok', updated_at: '2026-05-19T01:20:00Z' },
    }).as('providerHealth')
  })

  it('PROVIDER-OPENAI-UX-001/005 renders a clean card and dialog-owned OpenAI setup sections', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.get('[data-testid="provider-card-openai"]').should('be.visible')
    cy.get('[data-testid="provider-card-openai"]').should('contain', 'OpenAI / ChatGPT')
    cy.get('[data-testid="provider-card-openai"]').should('contain', 'Assistant')
    cy.get('[data-testid="provider-card-openai"]').should('not.contain', 'Validate')
    cy.get('[data-testid="provider-card-openai"]').should('not.contain', 'Test OpenAI')
    cy.get('[data-testid="provider-card-openai"]').should('not.contain', 'OpenAI is using:')

    cy.get('[data-testid="provider-open-openai"]').click()
    cy.get('[data-testid="openai-config-dialog"]').should('be.visible')
    cy.get('[data-testid="openai-browser-auth-section"]').should('contain', 'Browser Auth')
    cy.get('[data-testid="openai-api-key-section"]').should('contain', 'API key')
    cy.get('[data-testid="openai-test-section"]').should('contain', 'Test OpenAI')
    cy.contains('OpenAI is using:').should('not.exist')
  })

  it('PROVIDER-OPENAI-UX-006 keeps Browser Auth setup-needed until verifiable proof exists', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.get('[data-testid="provider-open-openai"]').click()
    cy.get('[data-testid="openai-browser-auth-status"]').should('contain', 'setup_needed')
    cy.get('[data-testid="openai-browser-auth-connect"]').should('be.disabled')
    cy.get('[data-testid="openai-browser-auth-setup-needed"]')
      .should('contain', 'callback/artifact')
      .and('contain', 'Navigation alone never marks OpenAI connected')
    cy.get('[data-testid="openai-active-method"]').should('have.value', 'None connected')
    cy.get('[data-testid="openai-test-run"]').should('be.disabled')
  })

  it('PROVIDER-OPENAI-UX-007 stores API key as a secret and non-secret defaults as settings', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.get('[data-testid="provider-open-openai"]').click()
    cy.get('[data-testid="provider-token-input"]').type('sk-test-openai')
    cy.get('[data-testid="openai-test-model"]').click()
    cy.contains('[role="option"]', 'gpt-4.1-mini').click()
    cy.get('[data-testid="openai-api-key-connect"]').click()

    cy.wait('@saveSettings').then(({ request }) => {
      expect(request.body.settings).to.include({
        openai_active_auth_method: 'api_key',
        'openai.active_auth_method': 'api_key',
        assistant_default_provider: 'openai',
        assistant_default_model: 'gpt-4.1-mini',
        'integration.openai.enabled': 'true',
      })
    })
    cy.wait('@saveSecret')
    cy.contains('OpenAI configuration saved.').should('be.visible')
  })

  it('PROVIDER-OPENAI-UX-008 binds empty API-key save validation to the token field', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.get('[data-testid="provider-open-openai"]').click()
    cy.get('[data-testid="openai-api-key-connect"]').click()

    cy.get('[data-testid="provider-token-input"]')
      .should('be.focused')
      .and('have.attr', 'aria-invalid', 'true')
      .and('have.attr', 'aria-describedby', 'provider-token-error')
    cy.get('#provider-token-error')
      .should('be.visible')
      .and('contain', 'OpenAI API key is required before connecting.')
    cy.get('@saveSettings.all').should('have.length', 0)
    cy.get('@saveSecret.all').should('have.length', 0)
  })

  it('PROVIDER-OPENAI-UX-008 blocks empty API-key validation before provider health', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.get('[data-testid="provider-open-openai"]').click()
    cy.get('[data-testid="openai-api-key-validate"]').click()

    cy.get('[data-testid="provider-token-input"]')
      .should('be.focused')
      .and('have.attr', 'aria-invalid', 'true')
      .and('have.attr', 'aria-describedby', 'provider-token-error')
    cy.get('#provider-token-error')
      .should('be.visible')
      .and('contain', 'OpenAI API key is required before validating.')
    cy.get('@providerHealth.all').should('have.length', 0)
  })
})
