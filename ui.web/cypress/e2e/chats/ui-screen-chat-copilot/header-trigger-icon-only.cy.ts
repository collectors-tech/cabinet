describe('chats/ui-screen-chat-copilot/header-trigger', () => {
  function signInToChats() {
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true }).then((bootstrap) => {
      cy.request('PUT', '/api/profiles/active', { profile_id: bootstrap.profile_id })
        .its('status')
        .should('eq', 200)
      cy.visit('/sign-in?redirect=%2Fchats%2F', {
        onBeforeLoad(win) {
          win.localStorage.setItem(`cabinet.workspace.${bootstrap.profile_id}`, '1')
        },
      })
      cy.contains('button', 'Open local workspace').click()
      cy.get('body').then(($body) => {
        const profileButton = `Use ${bootstrap.profile_name}`
        if ($body.text().includes(profileButton)) {
          cy.contains('button', profileButton).click()
        }
      })
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
  }

  it('UI-SCREEN-CHAT-COPILOT-006 renders icon-only header trigger and preserves chat rail behavior', () => {
    signInToChats()

    cy.get('[data-testid="shell-chat-toggle"]')
      .should('be.visible')
      .and('have.attr', 'aria-label')
    cy.get('[data-testid="shell-chat-toggle"]').should('have.attr', 'title')

    cy.get('[data-testid="shell-chat-toggle"]')
      .invoke('text')
      .then((labelText) => {
        expect(labelText.trim()).to.equal('')
      })

    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-chat-rail"]').should('be.visible')
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-chat-rail"]').should('not.exist')
  })
})
