describe('chats/chat-integration-management', { retries: 0 }, () => {
  it('AGENT-INTEGRATIONS-001/#2185 configures a provider through governed Chat preview and one-time apply', () => {
    cy.e2eReset()
    cy.e2eBootstrap().then((state) => {
      cy.e2eSetSetupState('present')
      cy.request('PUT', `/api/profiles/${state.profile_id}/settings`, {
        settings: {
          'agent.authority.mode': 'approved_external_actions',
          'agent.authority.external_write_approved': 'true',
        },
      })

      cy.request('POST', '/api/chat/messages', {
        profile_id: state.profile_id,
        thread_id: state.thread_id,
        role: 'user',
        content:
          'Configure provider Voglers AGENT-2185-SYNTHETIC for its public catalogue, enable it for this profile, and show me the governed preview before applying.',
        context: {
          route: { pathname: '/integrations' },
          workspace: { id: 'integrations' },
          setup: { state: 'ready' },
          assistant: {
            provider: 'fake',
            model: 'cabinet-e2e-planner',
          },
        },
      }).then(({ body }) => {
        expect(body.agent_planner.skill_id).to.eq(
          'cabinet.integrations.configure_provider'
        )
        expect(body.agent_planner.parameters).to.deep.include({
          provider_id: 'voglers',
          setup_payload: 'public_catalogue',
          setup_step: 'public_catalogue',
          marketplace: 'public',
        })
        expect(body.agent_planner.confirmation_state).to.eq('preview_required')
        expect(body.agent_planner.preview_result).to.include({
          skill_id: 'cabinet.integrations.configure_provider',
          preview_status: 'previewed',
          mutation_applied: false,
        })
        expect(JSON.stringify(body)).not.to.include('provider_secret')

        const previewID = String(body.agent_planner.preview_result.preview_id)
        cy.useBootstrappedProfile(state.profile_id, state.profile_name, {
          path: '/chats/',
          shellWorkspace: 'navigation',
        })
        cy.visit(`/chats/?thread_id=${encodeURIComponent(state.thread_id)}`)
        cy.get('[data-testid="chat-agent-response-state"]')
          .should('have.attr', 'data-agent-state', 'preview_required')
          .and('contain', 'Preview required')
        cy.get('[data-testid="chat-agent-planner-card"]')
          .should('contain', 'Review required')
          .and('contain', 'Voglers public catalogue')

        cy.intercept('POST', '/api/agent/skills/apply').as('applyProvider')
        cy.get('[data-testid="chat-agent-preview-apply"]').click()
        cy.wait('@applyProvider').then(({ request, response }) => {
          expect(request.body).to.deep.eq({
            profile_id: state.profile_id,
            preview_id: previewID,
            confirm: true,
          })
          expect(response?.statusCode).to.eq(200)
          expect(response?.body).to.include({
            preview_status: 'applied',
            mutation_applied: true,
          })
          expect(response?.body.target).to.deep.include({
            provider_id: 'voglers',
            operation: 'integrations.provider.configure',
          })
          expect(response?.body.target).not.to.have.property('provider_secret')
        })

        cy.request(`/api/profiles/${state.profile_id}/settings`).then(
          ({ body: settingsBody }) => {
            expect(settingsBody.settings).to.include({
              'integration.voglers.enabled': 'true',
              'integration.voglers.setup_step': 'public_catalogue',
              'integration.voglers.marketplace': 'public',
            })
          }
        )

        cy.request({
          method: 'POST',
          url: '/api/agent/skills/apply',
          failOnStatusCode: false,
          body: {
            profile_id: state.profile_id,
            preview_id: previewID,
            confirm: true,
          },
        }).then((replay) => {
          expect(replay.status).to.eq(409)
          expect(replay.body.error).to.eq('agent_skill_preview_already_applied')
        })
      })
    })
  })
})
