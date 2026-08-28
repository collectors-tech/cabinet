describe('chats/agent-admin-session-authority', () => {
  it('AGENT-SKILLS-REGISTRY-013/#2088 derives users authority from the server session', () => {
    cy.e2eReset()
    cy.e2eBootstrap().then((state) => {
      const requestBody = {
        skill_id: 'cabinet.users.search',
        profile_id: state.profile_id,
        agent_context: {
          profile_id: state.profile_id,
          workspace_id: 'client-spoof-must-not-authorize',
          route_id: '/settings/users',
          surface_id: 'users.table',
          setup_state: 'ready',
          admin_session: 'authorized',
          role: 'owner',
          authority: { allowed: true },
        },
        parameters: {
          query: 'owner',
          admin_session: 'authorized',
          role: 'admin',
        },
      }

      cy.request({
        method: 'POST',
        url: '/api/agent/skills/preview',
        body: requestBody,
        failOnStatusCode: false,
      }).then((response) => {
        expect(response.status).to.eq(401)
        expect(response.body.error).to.eq(
          'agent_admin_authentication_required'
        )
        expect(JSON.stringify(response.body)).not.to.contain('users')
        expect(JSON.stringify(response.body)).not.to.contain(
          'client-spoof-must-not-authorize'
        )
      })

      cy.request({
        method: 'POST',
        url: '/api/agent/skills/preview',
        headers: { 'X-Cabinet-Session': state.session_token },
        body: requestBody,
      }).then((response) => {
        expect(response.status).to.eq(200)
        expect(response.body.skill_id).to.eq('cabinet.users.search')
        const serialized = JSON.stringify(response.body)
        expect(serialized).not.to.contain(state.session_token)
        expect(serialized).not.to.contain('admin_session')
        expect(serialized).not.to.contain('client-spoof-must-not-authorize')
      })
    })
  })

  it('AGENT-SKILLS-REGISTRY-014/#2089 uses an action-specific one-time confirmation for Chat user removal', () => {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then((state) => {
      cy.useBootstrappedProfile(state.profile_id, state.profile_name, {
        path: '/chats/',
        shellWorkspace: 'navigation',
      })
      cy.visit(`/chats/?thread_id=${encodeURIComponent(state.thread_id)}`)

      cy.request('POST', '/api/users', {
        firstName: 'Removable',
        lastName: 'Operator',
        username: 'removable-operator',
        email: 'removable.operator@example.test',
        phoneNumber: '+61 400 000 208',
        role: 'view',
      }).then((userResponse) => {
        expect(userResponse.status).to.eq(201)
        const targetUserID = String(userResponse.body.id)

        cy.request({
          method: 'POST',
          url: '/api/chat/messages',
          headers: { 'X-Cabinet-Session': state.session_token },
          body: {
            profile_id: state.profile_id,
            thread_id: state.thread_id,
            role: 'user',
            content: `Remove user AGENT-2089-SYNTHETIC target ${targetUserID}`,
            context: {
              route: { pathname: '/settings/users' },
              workspace: { id: 'settings' },
              setup: { state: 'ready' },
              assistant: {
                provider: 'fake',
                model: 'cabinet-e2e-planner',
              },
            },
          },
        }).then((messageResponse) => {
          expect(messageResponse.status).to.eq(201)
          const preview = messageResponse.body.agent_planner.preview_result
          expect(preview.skill_id).to.eq('cabinet.users.remove_user')
          expect(preview.preview_status).to.eq('previewed')
          expect(preview.strong_confirmation_required).to.eq(true)
          expect(preview.strong_confirmation_endpoint).to.eq(
            '/api/agent/skills/confirm-destructive'
          )

          cy.intercept('GET', '/api/agent/skills/preview*').as(
            'durablePreviewVerification'
          )
          cy.reload()
          cy.wait('@durablePreviewVerification')
          cy.get('[data-testid="chat-agent-planner-card"]').should(
            'contain',
            'Review required'
          )
          cy.intercept('POST', '/api/agent/skills/confirm-destructive').as(
            'strongReview'
          )
          cy.get('[data-testid="chat-agent-preview-apply"]')
            .should('contain', 'Review destructive action')
            .and('be.enabled')
          cy.get('[data-testid="chat-agent-preview-apply"]')
            .click()
          cy.wait('@strongReview').then(({ request, response }) => {
            expect(request.body).to.deep.eq({
              profile_id: state.profile_id,
              preview_id: preview.preview_id,
            })
            expect(response?.body.action).to.eq('remove_user')
            expect(response?.body.target).to.include({
              target_user: targetUserID,
              target_email: 'removable.operator@example.test',
              protected: false,
            })
            expect(response?.body.confirmation_token).to.match(/^[a-f0-9]{64}$/)
          })
          cy.get('[data-testid="chat-agent-strong-confirmation-impact"]')
            .should('be.visible')
            .and('contain', 'removable.operator@example.test')
            .and('contain', 'Revoke that user')

          cy.intercept('POST', '/api/agent/skills/apply').as('strongApply')
          cy.get('[data-testid="chat-agent-preview-apply"]')
            .should('contain', 'Confirm destructive action')
            .click()
          cy.wait('@strongApply').then(({ request, response }) => {
            expect(request.body.profile_id).to.eq(state.profile_id)
            expect(request.body.preview_id).to.eq(preview.preview_id)
            expect(request.body.confirm).to.eq(true)
            expect(request.body.strong_confirmation_token).to.match(
              /^[a-f0-9]{64}$/
            )
            expect(response?.body.preview_status).to.eq('applied')
            expect(response?.body.mutation_applied).to.eq(true)
          })
          cy.get('[data-testid="chat-agent-planner-card"]').should(
            'contain',
            'Applied once'
          )
          cy.request('/api/users').then((usersResponse) => {
            expect(usersResponse.body.users).not.to.deep.include({
              id: targetUserID,
            })
            expect(
              usersResponse.body.users.some(
                (user: { id: string }) => user.id === targetUserID
              )
            ).to.eq(false)
          })

          cy.get('[data-testid="shell-workspace-assistant"]').click()
          cy.get('[data-testid="shell-assistant-thread-select"]', {
            timeout: 20000,
          }).select(state.thread_id)
          cy.get('[data-testid="shell-assistant-thread-id"]').should(
            'have.text',
            state.thread_id
          )
          cy.get('[data-testid="shell-assistant-agent-planner-card"]')
            .should('contain', 'Applied once')
            .and('not.contain', 'Confirm destructive action')
          cy.get(
            '[data-testid="shell-assistant-agent-strong-confirmation-impact"]'
          ).should('not.exist')
        })
      })
    })
  })
})
