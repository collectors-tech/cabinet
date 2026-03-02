describe('UI-SCREEN-INVENTORY-PHOTOS', () => {
  function signIn() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-photos@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('PHOTOS-MEDIA-004 opens fullscreen photo viewer and navigates next/previous', () => {
    signIn()

    cy.get('[data-testid="inventory-photos-section"]').should('be.visible')
    cy.get('[data-testid="inventory-photo-thumb"]').first().click()
    cy.get('[data-testid="inventory-photo-fullscreen"]').should('be.visible')
    cy.get('[data-testid="inventory-photo-next"]').click()
    cy.get('[data-testid="inventory-photo-prev"]').click()
    cy.get('[data-testid="inventory-photo-fullscreen-close"]').click()
    cy.get('[data-testid="inventory-photo-fullscreen"]').should('not.exist')
  })
})
