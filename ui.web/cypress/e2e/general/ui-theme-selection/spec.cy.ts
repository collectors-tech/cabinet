describe('ui-theme-selection', () => {
  function openThemeMenu() {
    cy.get('button')
      .filter((_i, el) => (el.textContent || '').includes('Toggle theme'))
      .first()
      .click({ force: true })
  }

  function signInToHome() {
    cy.visit('/sign-in?redirect=%2F')
    cy.get('input[name="email"]').clear().type('e2e-theme@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('eq', '/')
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToHome()
  })

  it('UI-THEME-SELECTION-001 applies and persists light/dark/system theme modes across navigation', () => {
    openThemeMenu()
    cy.contains('[role="menuitem"]', 'Dark').click()
    cy.get('html').should('have.class', 'dark')
    cy.getCookie('vite-ui-theme').its('value').should('eq', 'dark')

    cy.visit('/settings')
    cy.location('pathname').should('match', /^\/settings\/profile\/?$/)
    cy.get('html').should('have.class', 'dark')

    openThemeMenu()
    cy.contains('[role="menuitem"]', 'Light').click()
    cy.get('html').should('have.class', 'light')
    cy.getCookie('vite-ui-theme').its('value').should('eq', 'light')

    cy.get('button[aria-label="Open theme settings"]').click()
    cy.get('[aria-label="Select system"]').click()
    cy.getCookie('vite-ui-theme').its('value').should('eq', 'system')
  })

  it('UI-THEME-SELECTION-002 keeps header and configuration drawer theme controls synchronized', () => {
    openThemeMenu()
    cy.contains('[role="menuitem"]', 'Light').click()
    cy.get('html').should('have.class', 'light')

    cy.get('button[aria-label="Open theme settings"]').click()
    cy.get('[aria-label="Select light"]').should('have.attr', 'data-state', 'checked')
    cy.get('[aria-label="Select dark"]').click()
    cy.get('html').should('have.class', 'dark')
    cy.get('[aria-label="Select dark"]').should('have.attr', 'data-state', 'checked')
    cy.getCookie('vite-ui-theme').its('value').should('eq', 'dark')
  })

  it('UI-THEME-SELECTION-003 updates layout density live and persists preference after navigation', () => {
    cy.location('pathname').should('eq', '/')

    cy.get('button[aria-label="Open theme settings"]').click()
    cy.get('[aria-label="Select compact"]').click()
    cy.getCookie('layout_collapsible').its('value').should('eq', 'icon')

    cy.get('[aria-label="Select full layout"]').click()
    cy.getCookie('layout_collapsible').its('value').should('eq', 'offcanvas')

    cy.location('pathname').should('eq', '/')
    cy.visit('/inventory')
    cy.location('pathname').should('match', /^\/inventory\/?$/)
    cy.getCookie('layout_collapsible').its('value').should('eq', 'offcanvas')
  })
})
