describe('settings billing screen', () => {
  function signInToBilling() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Fbilling')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/billing\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToBilling()
  })

  it('UI-SCREEN-SETTINGS-BILLING-001 renders disabled static billing state without portal mutation', () => {
    cy.contains('h3', 'Billing').should('be.visible')
    cy.get('aside a[href="/settings/billing"]').should(
      'have.attr',
      'aria-current',
      'page'
    )

    cy.contains('p', 'Plan')
      .should('be.visible')
      .parent()
      .within(() => {
        cy.contains(
          'Billing controls are visible here and sync with cloud entitlement state.'
        ).should('be.visible')
      })

    cy.contains('p', 'License Status')
      .should('be.visible')
      .parent()
      .within(() => {
        cy.contains(
          'Check current license tier and renewal state for this account.'
        ).should('be.visible')
      })

    cy.contains('button', 'Open Billing Portal (Coming soon)')
      .should('be.visible')
      .and('be.disabled')

    cy.contains('a', 'Open Billing Portal').should('not.exist')
    cy.location('pathname').should('match', /^\/settings\/billing\/?$/)
  })
})
