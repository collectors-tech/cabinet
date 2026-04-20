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
            integration_mode: 'assistant_default',
            auth_mode: 'api_key',
            state: 'needs_config',
            has_token: false,
            capabilities: {
              search: false,
              stock_observation: false,
              pricing: false,
              health: true,
              assistant: true,
            },
            health: { status: 'needs_config' },
            last_run: { status: 'not_run' },
          },
        ],
      },
    }).as('providersRegistry')

    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: {
        settings: {
          openai_api_key: '',
          assistant_default_provider: 'openai',
          assistant_default_model: 'gpt-4o-mini',
        },
      },
    }).as('profileSettings')

    cy.intercept('PUT', '/api/profiles/*/settings', (req) => {
      req.reply({ statusCode: 200, body: req.body })
    }).as('saveSettings')
  })

  it('PROVIDER-OPENAI-UX-001/002/003 renders visible provider, auth-mode guidance, and assistant capability narrative', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.contains('OpenAI / ChatGPT').should('be.visible')
    cy.contains('platform.openai.com').should('be.visible')
    cy.contains('assistant_default').should('be.visible')
    cy.contains('api_key').should('be.visible')
    cy.contains('needs_config').should('be.visible')
    cy.contains('Assistant').should('be.visible')

    cy.contains('OpenAI / ChatGPT').click()
    cy.contains('OpenAI / ChatGPT').should('be.visible')
    cy.contains('api_key').should('be.visible')
    cy.contains('platform.openai.com').should('be.visible')
    cy.contains('Assistant').should('be.visible')
    cy.contains('gpt-4o-mini').should('be.visible')
  })

  it('PROVIDER-OPENAI-UX-004 persists assistant default provider/model through integrations settings', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.contains('OpenAI / ChatGPT').click()
    cy.get('input').filter(':visible').first().type('sk-test-openai')
    cy.contains('gpt-4o-mini').should('be.visible')
    cy.contains('Save Integration').click()
    cy.wait('@saveSettings').then(({ request }) => {
      expect(request.body.assistant_default_provider).to.eq('openai')
      expect(request.body.assistant_default_model).to.eq('gpt-4o-mini')
    })
  })

  it('PROVIDER-OPENAI-UX-004 boots Assistant workspace from saved provider/model defaults', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.contains('OpenAI / ChatGPT').click()
    cy.get('[data-testid="provider-openai-default-provider"]').clear().type('openai')
    cy.get('[data-testid="provider-openai-default-model"]').select('gpt-4.1-mini')
    cy.contains('Save Integration').click()
    cy.wait('@saveSettings')

    cy.visit('/inventory')
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-thread-provider"]').should('contain', 'openai')
    cy.get('[data-testid="shell-assistant-thread-model"]').should('contain', 'gpt-4.1-mini')
  })

  it('PROVIDER-OPENAI-UX-004 syncs Assistant defaults immediately after integrations save without reload', () => {
    cy.wait('@activeProfile')
    cy.wait('@providersRegistry')
    cy.wait('@profileSettings')

    cy.visit('/inventory')
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-thread-provider"]').should('contain', 'openai')
    cy.get('[data-testid="shell-assistant-thread-model"]').should('contain', 'gpt-4o-mini')

    cy.visit('/integrations')
    cy.contains('OpenAI / ChatGPT').click()
    cy.get('[data-testid="provider-openai-default-provider"]').clear().type('openai')
    cy.get('[data-testid="provider-openai-default-model"]').select('gpt-4.1-mini')
    cy.contains('Save Integration').click()
    cy.wait('@saveSettings')

    cy.visit('/inventory')
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-thread-provider"]').should('contain', 'openai')
    cy.get('[data-testid="shell-assistant-thread-model"]').should('contain', 'gpt-4.1-mini')
  })
})
