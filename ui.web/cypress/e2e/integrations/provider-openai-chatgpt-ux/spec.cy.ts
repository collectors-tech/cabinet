describe('integrations/provider-openai-chatgpt-ux', () => {
  function openAPIKeyAdvanced() {
    cy.get('[data-testid="openai-api-key-advanced"]').click()
    cy.get('[data-testid="openai-api-key-section"]').should('be.visible')
  }

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
            model_options: ['gpt-5.6-luna', 'gpt-4o-mini', 'gpt-4.1-mini'],
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

    cy.intercept('DELETE', '/api/profiles/*/secrets?key=openai_api_key', {
      statusCode: 204,
      body: {},
    }).as('deleteSecret')

    cy.intercept('GET', '/api/provider/health*', {
      statusCode: 200,
      body: { status: 'ok', message: 'ok', updated_at: '2026-05-19T01:20:00Z' },
    }).as('providerHealth')

    cy.intercept('GET', '/api/providers/openai/browser-auth*', {
      statusCode: 200,
      body: {
        provider: 'openai',
        auth_method: 'browser_auth',
        state: 'signed_out',
        authenticated: false,
        profile_connected: false,
        recommended: true,
        message: 'Sign in with ChatGPT to use Cabinet Chat.',
      },
    }).as('browserAuthStatus')

    cy.intercept('POST', '/api/providers/openai/browser-auth', {
      statusCode: 200,
      body: {
        provider: 'openai',
        auth_method: 'browser_auth',
        state: 'connected',
        authenticated: true,
        profile_connected: true,
        recommended: true,
        message: 'ChatGPT is connected to Cabinet.',
      },
    }).as('browserAuthConnect')

    cy.intercept('POST', '/api/provider/test', (req) => {
      expect(req.body).to.deep.equal({
        provider: 'openai',
        profile_id: 'e2e-profile-001',
      })
      req.reply({
        statusCode: 200,
        body: {
          provider: 'openai',
          status: 'ready',
          code: 'OPENAI_BROWSER_AUTH_PROVIDER_TEST_PASSED',
          auth_method: 'browser_auth',
          credential_present: true,
          provider_test_passed: true,
          provider_test_state: 'passed',
          provider_test_artifact_id: 'local_codex_chatgpt_test',
          message: 'ChatGPT browser-auth runtime test passed.',
          checked_at: '2026-08-15T00:00:00Z',
        },
      })
    }).as('openAIProviderTest')
  })

  it('PROVIDER-OPENAI-UX-001/005 renders a clean card and dialog-owned OpenAI setup sections', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.get('button[aria-label="Switch to cards view"]').click()
    cy.get('[data-testid="provider-card-openai"]').should('be.visible')
    cy.get('[data-testid="provider-card-openai"]').should('contain', 'OpenAI / ChatGPT')
    cy.get('[data-testid="provider-card-openai"]').should('contain', 'Assistant')
    cy.get('[data-testid="provider-card-openai"]').should('not.contain', 'Validate')
    cy.get('[data-testid="provider-card-openai"]').should('not.contain', 'Test OpenAI')
    cy.get('[data-testid="provider-card-openai"]').should('not.contain', 'OpenAI is using:')

    cy.get('[data-testid="provider-open-openai"]').click()
    cy.get('[role="dialog"]').should('be.visible')
    cy.get('[data-testid="openai-config-dialog"]').should('exist')
    cy.get('[data-testid="openai-browser-auth-section"]')
      .should('contain', 'Sign in with ChatGPT')
      .and('contain', 'Recommended')
      .and('contain', 'No API key required')
    cy.get('[data-testid="openai-browser-auth-connect"]')
      .should('be.enabled')
      .and('contain', 'Continue with ChatGPT')
    cy.get('[data-testid="openai-api-key-section"]').should('not.be.visible')
    cy.get('[data-testid="openai-api-key-advanced"]').should(
      'contain',
      'Advanced: use an API key',
    )
    cy.get('[data-testid="provider-detail-api-family"]').should('not.exist')
    cy.get('[data-testid="provider-detail-config-schema"]').should('not.exist')

    cy.get('[data-testid="openai-config-dialog"]').within(() => {
      cy.contains('button', 'Sync').should('not.exist')
      cy.contains('OpenAI is using:').should('not.exist')
      cy.contains('API Family:').should('not.exist')
      cy.contains('Config Schema:').should('not.exist')
    })
  })

  it('PROVIDER-OPENAI-UX-006 makes verified ChatGPT browser sign-in the primary connection path', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.get('[data-testid="provider-open-openai"]').click()
    cy.wait('@browserAuthStatus')
    cy.get('[data-testid="openai-browser-auth-status"]').should('contain', 'Not connected')
    cy.get('[data-testid="openai-browser-auth-connect"]').click()
    cy.wait('@browserAuthConnect').its('request.body').should('deep.equal', {
      profile_id: 'e2e-profile-001',
    })
    cy.get('[data-testid="openai-browser-auth-status"]').should('contain', 'Connected')
    cy.get('[data-testid="openai-active-method"]').should('have.value', 'Browser Auth')
    cy.get('[data-testid="openai-test-prompt"]').should('not.exist')
    cy.get('[data-testid="openai-test-run"]')
      .should('be.enabled')
      .and('contain', 'Test connection')
      .click()
    cy.wait('@openAIProviderTest')
    cy.get('[data-testid="openai-provider-test-result"]')
      .should('be.visible')
      .and('contain', 'OpenAI connection is ready for Cabinet Chat.')
    cy.contains('button', 'Done').click()
    cy.get('[data-testid="provider-row-openai"]')
      .should('contain', 'Connected')
      .and('contain', 'Health: ready')
  })

  it('PROVIDER-OPENAI-UX-006 explains the local runtime prerequisite without promoting API keys', () => {
    cy.intercept('GET', '/api/providers/openai/browser-auth*', {
      statusCode: 200,
      body: {
        provider: 'openai',
        auth_method: 'browser_auth',
        state: 'runtime_missing',
        authenticated: false,
        profile_connected: false,
        recommended: true,
        message: 'Install the OpenAI Codex runtime to use ChatGPT browser sign-in.',
      },
    }).as('browserAuthRuntimeMissing')

    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')
    cy.get('[data-testid="provider-open-openai"]').click()
    cy.wait('@browserAuthRuntimeMissing')
    cy.get('[data-testid="openai-browser-auth-status"]').should(
      'contain',
      'Codex runtime required',
    )
    cy.get('[data-testid="openai-browser-auth-section"]')
      .should('contain', 'Install OpenAI Codex or the ChatGPT IDE extension, then retry.')
      .and('contain', 'No API key required')
    cy.get('[data-testid="openai-api-key-section"]').should('not.be.visible')
  })

  it('PROVIDER-OPENAI-UX-007 stores API key as a secret and non-secret defaults as settings', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.get('[data-testid="provider-open-openai"]').click()
    openAPIKeyAdvanced()
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
    openAPIKeyAdvanced()
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
    openAPIKeyAdvanced()
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

  it('PROVIDER-OPENAI-UX-009 disconnects the API key without clearing Browser Auth state', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.get('[data-testid="provider-open-openai"]').click()
    openAPIKeyAdvanced()
    cy.get('[data-testid="openai-api-key-disconnect"]').click()

    cy.wait('@deleteSecret')
    cy.wait('@saveSettings').then(({ request }) => {
      expect(request.body.settings).to.include({
        openai_active_auth_method: '',
        'openai.active_auth_method': '',
        'integration.openai.enabled': 'false',
      })
      expect(request.body.settings).to.include({
        'openai.browser_auth_state': 'setup_needed',
      })
    })
    cy.contains('OpenAI API key disconnected.').should('be.visible')
    cy.get('[data-testid="openai-api-key-status"]').should('contain', 'setup_needed')
    cy.get('[data-testid="openai-active-method"]').should('have.value', 'None connected')
    cy.get('[data-testid="openai-test-run"]').should('be.disabled')
  })
})
