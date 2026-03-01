describe('settings/appearance', () => {
  function signInToSettings() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Fappearance')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/appearance\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()
  })

  it('UI-SCREEN-SETTINGS-APPEARANCE-001 applies and persists theme/font preferences', () => {
    cy.contains('button', 'Update preferences').should('not.be.disabled')
    cy.get('select[name="font"]').select('manrope')
    cy.contains('span', 'Dark').click()
    cy.contains('button', 'Update preferences').click()
    cy.contains('Appearance settings saved.').should('be.visible')

    cy.get('html').should('have.class', 'dark')
    cy.reload()
    cy.get('select[name="font"]').should('have.value', 'manrope')
    cy.get('html').should('have.class', 'dark')
  })
})
