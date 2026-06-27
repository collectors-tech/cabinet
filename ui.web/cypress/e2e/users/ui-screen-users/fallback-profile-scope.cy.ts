describe('ui-screen-users-fallback-profile-scope', () => {
  it('UI-SCREEN-USERS-005 keeps users route functional when active profile is missing', () => {
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.visit('/sign-in?redirect=%2Fusers%2F')
    cy.get('input[name="email"]').clear().type('e2e-users@example.com')
    cy.get('input[name="password"]').clear().type('password123')

    cy.intercept('GET', '/api/users*').as('listUsers')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/users\/?$/)

    cy.wait('@listUsers').its('response.statusCode').should('eq', 200)
    cy.contains('users_fetch_failed_404').should('not.exist')
    cy.get('[data-testid="users-header-title"]').should('contain', 'Users')
  })
})
