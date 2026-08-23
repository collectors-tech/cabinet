describe('ui-screen-users-fallback-profile-scope', () => {
  it('UI-SCREEN-USERS-005 keeps users route functional when active profile is missing', () => {
    cy.e2eReset()
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200)
    cy.visit('/sign-in?redirect=%2Fusers%2F')

    cy.intercept('GET', '/api/users*').as('listUsers')
    cy.contains('button', 'Open local workspace').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/users\/?$/)

    cy.wait('@listUsers').its('response.statusCode').should('eq', 200)
    cy.contains('users_fetch_failed_404').should('not.exist')
    cy.get('[data-testid="users-header-title"]').should('contain', 'Users')
  })
})
