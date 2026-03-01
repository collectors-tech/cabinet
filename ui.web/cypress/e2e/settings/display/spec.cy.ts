describe('settings/display', () => {
  function signInToSettings() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Fdisplay')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/display\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()
  })

  it('UI-SCREEN-SETTINGS-DISPLAY-001 persists sidebar visibility selection', () => {
    cy.contains('button', 'Update display').should('not.be.disabled')
    cy.get('[data-testid="settings-display-desktop"]').click()
    cy.contains('button', 'Update display').click()
    cy.contains('Display settings saved.').should('be.visible')

    cy.reload()
    cy.get('[data-testid="settings-display-desktop"]')
      .should('have.attr', 'data-state', 'checked')
  })

  it('UI-SCREEN-SETTINGS-DISPLAY-002 requires at least one selected display item', () => {
    cy.contains('button', 'Clear selection').click()
    cy.get('[data-testid^="settings-display-"][data-state="checked"]').each(
      ($checkbox) => {
        cy.wrap($checkbox).click()
      }
    )
    cy.get('[data-testid^="settings-display-"][data-state="checked"]').should(
      'have.length',
      0
    )
    cy.contains('button', 'Update display').click()
    cy.contains('You have to select at least one item.').should('be.visible')
  })
})
