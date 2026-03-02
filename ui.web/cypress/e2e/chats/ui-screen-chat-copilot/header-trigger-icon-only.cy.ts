describe('chats/ui-screen-chat-copilot/header-trigger', () => {
  function signInToChats() {
    cy.visit('/sign-in?redirect=%2Fchats%2F')
    cy.get('input[name="email"]').clear().type('e2e-chat-trigger@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
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
