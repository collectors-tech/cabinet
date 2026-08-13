describe('chats/agent-compact-accessibility', { retries: 0 }, () => {
  function bootstrap(path = '/inventory/') {
    cy.viewport(640, 360)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.e2eEnsureSignedOut()
    cy.stubLocalServerSession('e2e-profile-001')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path })
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      new RegExp(`^${path.replace(/\/$/, '')}/?$`)
    )
  }

  function expectInsideViewport(testID: string) {
    cy.get(`[data-testid="${testID}"]`)
      .scrollIntoView()
      .should('be.visible')
      .then(($element) => {
        const bounds = $element[0].getBoundingClientRect()
        expect(bounds.left, `${testID} left`).to.be.at.least(0)
        expect(bounds.right, `${testID} right`).to.be.at.most(
          Cypress.config('viewportWidth')
        )
        expect(bounds.top, `${testID} top`).to.be.at.least(0)
        expect(bounds.bottom, `${testID} bottom`).to.be.at.most(
          Cypress.config('viewportHeight')
        )
      })
  }

  it('#2100 keeps contextual Agent controls reachable and hands focus into full Chat', () => {
    bootstrap()

    cy.get('[data-testid="shell-chat-toggle"]', { timeout: 20000 })
      .focus()
      .should('be.focused')
      .type('{enter}')
    cy.get('[data-testid="shell-assistant-modal-content"]', {
      timeout: 20000,
    }).should('be.visible')
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 20000,
    }).should(($threadID) => {
      expect($threadID.text().trim()).not.to.be.oneOf(['', 'bootstrapping'])
    })

    expectInsideViewport('shell-assistant-attachment-picker')
    expectInsideViewport('shell-assistant-compose-input')
    expectInsideViewport('shell-assistant-send-button')

    cy.get('[data-testid="shell-assistant-attachment-picker"]')
      .should('not.be.disabled')
      .focus()
      .should('be.focused')
    cy.get('[data-testid="shell-assistant-attachment-upload"]').should(
      'be.disabled'
    )
    cy.press(Cypress.Keyboard.Keys.TAB)
    cy.get('[data-testid="shell-assistant-compose-input"]').should(
      'be.focused'
    )

    cy.get('[data-testid="shell-assistant-open-full-chat"]')
      .scrollIntoView()
      .focus()
      .should('be.visible')
      .and('be.focused')
      .type(' ')

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
    cy.location('search').should('contain', 'thread_id=')
    cy.get('[data-testid="chat-main-surface"]', { timeout: 20000 }).should(
      'be.visible'
    )
    cy.get('[data-testid="chat-compose-input"]').should('be.focused')
  })

  it('#2100 keeps full Agent result, Apply, Cancel, attachments and composer reachable at 200 percent zoom equivalent', () => {
    bootstrap('/chats/')
    cy.request('POST', '/api/chat/threads', {
      profile_id: 'e2e-profile-001',
      title: 'Compact Agent accessibility',
    }).then(({ body }) => {
      cy.visit(`/chats/?thread_id=${encodeURIComponent(String(body.id))}`)
    })

    cy.location('search').then((search) => {
      const threadID = new URLSearchParams(search).get('thread_id') ?? ''
      cy.request('POST', '/api/chat/actions/preview', {
        profile_id: 'e2e-profile-001',
        thread_id: threadID,
        action: 'update_inventory_item',
        payload: {
          item_id: 'e2e-item-001',
          part_number: 'A11Y-2100',
          title: 'Compact Agent result',
        },
      }).then(({ body }) => {
        cy.visit(
          `/chats/?thread_id=${encodeURIComponent(threadID)}&preview_id=${encodeURIComponent(String(body.id))}`
        )
      })
    })

    cy.get('[data-testid="chat-compose-input"]', { timeout: 20000 })
      .scrollIntoView()
      .focus()
      .should('be.focused')

    expectInsideViewport('chat-composer-attachment-button')
    expectInsideViewport('chat-compose-input')
    expectInsideViewport('chat-send-button')

    cy.get('[data-testid="chat-action-preview"]')
      .scrollIntoView()
      .should('be.visible')
      .and('contain', 'A11Y-2100')

    cy.get('[data-testid="chat-apply-action-button"]')
      .focus()
      .should('be.focused')
      .type(' ')
    cy.get('[data-testid="chat-apply-confirm-cancel"]')
      .should('be.visible')
      .focus()
      .type(' ')
    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('not.exist')
    cy.get('[data-testid="chat-apply-action-button"]').should('be.focused')
    expectInsideViewport('chat-cancel-action-button')
  })
})
