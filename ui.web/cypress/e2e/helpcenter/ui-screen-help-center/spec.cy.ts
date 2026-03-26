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

  it('UI-SCREEN-HELP-CENTER-001 renders article library on help-center route', () => {
    signInToHelpCenter()

    cy.contains('h1', 'Help Center').should('be.visible')
    cy.get('[data-testid="help-center-article-library"]').should('be.visible')
    cy.get('[data-testid="help-center-library-summary"]').should('contain.text', 'Articles available')
    cy.get('[data-testid="help-center-article-link-getting-started-login-database-setup"]').should('be.visible')
    cy.get('[data-testid="help-center-article-link-section-inventory"]').should('be.visible')
    cy.contains('Oops! Something went wrong').should('not.exist')
  })

  it('UI-SCREEN-HELP-CENTER-002 renders selected article content in-app', () => {
    signInToHelpCenter()

    cy.get('[data-testid="help-center-selected-article-title"]').should('contain.text', 'Login and Database Setup')
    cy.get('[data-testid="help-center-article-link-section-inventory"]').click()
    cy.get('[data-testid="help-center-selected-article-title"]').should('contain.text', 'Inventory')
    cy.get('[data-testid="help-center-article-viewer"]').should('contain.text', 'Managing owned items')
    cy.get('[data-testid="help-center-article-content-section-inventory"]').should('contain.text', 'Photos, barcodes, AI assist')
  })

  it('UI-SCREEN-HELP-CENTER-003 preserves shell controls on help-center route', () => {
    signInToHelpCenter()

    cy.contains('button', /Search/i).should('be.visible')
    cy.contains('span', /toggle theme/i).should('exist')
    cy.get('[data-slot="sidebar-trigger"]').should('be.visible').click()
    cy.get('[data-slot="sidebar-trigger"]').should('be.visible')
    cy.contains(/ACC001|Local Admin/i).should('be.visible')
  })
})
