describe('ui-screen-help-center', () => {
  function signInToHelpCenter() {
    cy.visit('/sign-in?redirect=%2Fhelp-center%2F')
    cy.get('input[name="email"]').clear().type('e2e-help@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/help-center\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-HELP-CENTER-001 renders deterministic placeholder state on help-center route', () => {
    signInToHelpCenter()

    cy.contains('h1', 'Help Center').should('be.visible')
    cy.get('[data-testid="help-center-placeholder"]').should('be.visible')
    cy.contains('Documentation is being organized').should('be.visible')
    cy.contains('Oops! Something went wrong').should('not.exist')
  })

  it('UI-SCREEN-HELP-CENTER-002 preserves shell controls on help-center route', () => {
    signInToHelpCenter()

    cy.contains('button', /Search/i).should('be.visible')
    cy.contains('span', /toggle theme/i).should('exist')
    cy.get('[data-slot="sidebar-trigger"]').should('be.visible').click()
    cy.get('[data-slot="sidebar-trigger"]').should('be.visible')
    cy.contains(/ACC001|Local Admin/i).should('be.visible')
  })
})
