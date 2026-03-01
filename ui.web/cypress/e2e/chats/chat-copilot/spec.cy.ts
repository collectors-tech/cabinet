describe('chats/chat-copilot', () => {
  function signInToInventory() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-chat-rail@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('CHAT-COPILOT-001 toggles chat rail from header without route context loss', () => {
    signInToInventory()

    cy.location('pathname').then((initialPathname) => {
      cy.get('[data-testid="shell-chat-toggle"]').should('contain', 'Open Chat').click()
      cy.get('[data-testid="shell-chat-rail"]').should('be.visible')
      cy.location('pathname').should('eq', initialPathname)

      cy.get('[data-testid="shell-chat-toggle"]').should('contain', 'Close Chat').click()
      cy.get('[data-testid="shell-chat-rail"]').should('not.exist')
      cy.location('pathname').should('eq', initialPathname)
    })
  })
})
