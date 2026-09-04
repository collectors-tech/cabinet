describe('chats/assistant-workspace-agent-authority', () => {
  const profileID = 'e2e-profile-001'

  function openInventory() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile(profileID, 'E2E Local', {
      path: '/inventory/',
      shellWorkspace: 'navigation',
    })
  }

  function openAgent() {
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 20000,
    }).should(($threadID) => {
      expect($threadID.text().trim()).not.to.eq('')
      expect($threadID.text().trim()).not.to.eq('bootstrapping')
    })
  }

  it('AGENT-AUTHORITY-005/007/#1932 re-checks profile policy when a server-owned preview is applied', () => {
    openInventory()
    cy.request('PUT', `/api/profiles/${profileID}/settings`, {
      settings: {
        'agent.authority.external_write_approved': 'false',
        'agent.authority.mode': 'ask_before_local_changes',
      },
    })
    openAgent()

    cy.get('[data-testid="shell-assistant-thread-id"]')
      .invoke('text')
      .then((rawThreadID) => {
        const threadID = rawThreadID.trim()
        cy.request('POST', '/api/chat/messages', {
          profile_id: profileID,
          thread_id: threadID,
          role: 'user',
          content:
            'Add wishlist entry AGENT-2097-SYNTHETIC title Authority policy recheck',
          context: {
            route: { pathname: '/inventory/' },
            workspace: { id: 'inventory' },
            setup: { state: 'ready' },
            assistant: {
              provider: 'fake',
              model: 'cabinet-e2e-planner',
            },
          },
        }).then(({ body: messageResponse }) => {
          const previewID = messageResponse.agent_planner.preview_result
            .preview_id as string
          expect(previewID).to.match(/^asp_[a-f0-9]+$/)

          cy.request('PUT', `/api/profiles/${profileID}/settings`, {
            settings: {
              'agent.authority.external_write_approved': 'false',
              'agent.authority.mode': 'read_only',
            },
          })
          cy.reload()
          openAgent()
          cy.intercept('POST', '/api/agent/skills/apply').as('policyRecheck')
          cy.get('[data-testid="shell-assistant-agent-preview-apply"]').click()
          cy.wait('@policyRecheck').then(({ request, response }) => {
            expect(Object.keys(request.body).sort()).to.deep.eq([
              'confirm',
              'preview_id',
              'profile_id',
            ])
            expect(request.body.preview_id).to.eq(previewID)
            expect(response?.statusCode).to.eq(409)
            expect(response?.body.error).to.eq('agent_authority_read_only')
          })
          cy.get('[data-testid="shell-assistant-agent-planner-card"]')
            .should('contain', 'agent_authority_read_only')
            .and('not.contain', 'cabinet.wishlist.create_entry')
          cy.request('/api/items?profile_id=e2e-profile-001').then(
            ({ body }) => {
              expect(JSON.stringify(body)).not.to.include(
                'AGENT-2097-SYNTHETIC'
              )
            }
          )
        })
      })
  })

  it('AGENT-CONTEXT-003/#1714/#2359 sends selected inventory detail context with natural conversation', () => {
    cy.intercept('GET', '/api/items*', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-agent-context-1',
            part_number: 'CTX-1714-ROW',
            title: 'Agent Context Row',
            status: 'todo',
            condition: 'new',
            category: 'feature',
            item_type: 'car',
            packaging_grade_type: 'carded',
            brand: 'AFX',
            priority: 'medium',
            description: '',
            notes: '',
            tags: [],
            source_urls: [],
          },
        ],
      },
    }).as('itemsForAgentContext')
    openInventory()
    cy.wait('@itemsForAgentContext')
    cy.get('[data-testid="inventory-item-row-item-agent-context-1"]')
      .should('be.visible')
      .click()
    openAgent()
    cy.get('body').should('not.have.css', 'pointer-events', 'none')
    cy.get('[data-testid="shell-assistant-compose-input"]').should(
      'not.have.css',
      'pointer-events',
      'none'
    )

    const intent = 'find inventory item with part number CTX-1714-ROW'
    cy.intercept('POST', '/api/chat/messages', (request) => {
      request.body.context = {
        ...(request.body.context ?? {}),
        assistant: {
          provider: 'fake',
          model: 'cabinet-e2e-planner',
        },
      }
      request.continue()
    }).as('selectedContext')
    cy.get('[data-testid="shell-assistant-compose-input"]').type(intent)
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@selectedContext').then(({ request, response }) => {
      expect(request.body.content).to.eq(intent)
      expect(request.body.agent_context.route_id).to.match(/^\/inventory\/?$/)
      expect(request.body.agent_context.surface_id).to.eq(
        'inventory.item.detail'
      )
      expect(request.body.agent_context.source_channel).to.eq('in-app')
      expect(request.body.agent_context.selected_record).to.deep.eq({
        type: 'inventory_item',
        id: 'item-agent-context-1',
      })
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.agent_planner.intent_domain).to.eq('inventory')
    })
    cy.get('[data-testid="shell-assistant-modal-content"]').should(
      'not.contain.text',
      'Agent Skill'
    )
  })
})
