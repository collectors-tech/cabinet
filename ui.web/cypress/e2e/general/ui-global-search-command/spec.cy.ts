describe('ui-global-search-command', () => {
  const commandInputSelector = 'input[placeholder="Type a command or search..."]'

  function signInToHome() {
    cy.visit('/sign-in?redirect=%2F')
    cy.get('input[name="email"]').clear().type('e2e-command@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('eq', '/dashboard')
  }

  function openCommandPaletteWithShortcut() {
    cy.get('body').click(0, 0).type('{ctrl}k')
  }

  function openCommandPaletteFromSearchButton() {
    cy.contains('button', /search/i).first().click()
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToHome()
  })

  it('UI-GLOBAL-SEARCH-COMMAND-001 opens command palette with Ctrl/Cmd+K and focuses input', () => {
    openCommandPaletteWithShortcut()
    cy.get(commandInputSelector).should('be.visible').and('be.focused')
  })

  it('UI-GLOBAL-SEARCH-COMMAND-002 executes navigation command and closes palette', () => {
    openCommandPaletteFromSearchButton()
    cy.get(commandInputSelector).type('Inventory')
    cy.contains('[cmdk-item]', 'Inventory').click()

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.get(commandInputSelector).should('not.exist')
  })

  it('UI-GLOBAL-SEARCH-COMMAND-003 executes theme action command and closes palette', () => {
    openCommandPaletteFromSearchButton()
    cy.get(commandInputSelector).type('Dark')
    cy.contains('[cmdk-item]', 'Dark').click()

    cy.get('html').should('have.class', 'dark')
    cy.get(commandInputSelector).should('not.exist')
  })
})
