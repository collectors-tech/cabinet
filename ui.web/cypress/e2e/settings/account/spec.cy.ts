describe('settings/account', () => {
  function signInToSettings() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Faccount')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/account\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()
  })

  it('UI-SCREEN-SETTINGS-ACCOUNT-001 persists account fields across reload', () => {
    cy.contains('button', 'Update account').should('not.be.disabled')
    cy.get('input[name="name"]').clear().type('Settings Account Name')
    cy.get('[data-testid="settings-account-language-trigger"]').click()
    cy.contains('[role="option"]', 'Japanese').click()
    cy.contains('button', 'Update account').click()
    cy.contains('Account settings saved.').should('be.visible')

    cy.reload()
    cy.get('input[name="name"]').should('have.value', 'Settings Account Name')
    cy.contains('button[role="combobox"]', 'Japanese').should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-ACCOUNT-002 blocks invalid account submission with field errors', () => {
    cy.contains('button', 'Update account').should('not.be.disabled')
    cy.get('input[name="name"]').clear()
    cy.contains('button', 'Update account').click()
    cy.contains('Please enter your name.').should('be.visible')
  })
})
