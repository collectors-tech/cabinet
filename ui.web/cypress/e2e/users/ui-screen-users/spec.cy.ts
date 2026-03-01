describe('ui-screen-users', () => {
  function signInToUsers() {
    cy.visit('/sign-in?redirect=%2Fusers%2F')
    cy.get('input[name="email"]').clear().type('e2e-users@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/users\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToUsers()
  })

  it('UI-SCREEN-USERS-001 supports filter, sort context, and pagination workflows', () => {
    cy.contains('h2', 'User List').should('be.visible')

    cy.contains('span', 'Go to next page').parents('button').first().click()
    cy.location('search').should('contain', 'page=2')

    cy.get('tbody tr').first().find('td').eq(1).invoke('text').then((usernameRaw) => {
      const username = usernameRaw.trim()
      const token = username.slice(0, Math.min(4, username.length))

      cy.get('input[placeholder="Filter users..."]').clear().type(token)
      cy.location('search').should('contain', 'username=')
      cy.get('tbody tr').first().find('td').eq(1).invoke('text').should('match', new RegExp(token, 'i'))
    })
  })

  it('UI-SCREEN-USERS-002 opens invite and add-user dialogs from primary actions', () => {
    cy.contains('button', 'Invite User').click()
    cy.contains('[role="dialog"]', 'Invite User').should('be.visible')
    cy.contains('[role="dialog"] button', 'Cancel').click()
    cy.contains('[role="dialog"]', 'Invite User').should('not.exist')

    cy.contains('button', 'Add User').click()
    cy.contains('[role="dialog"]', 'Add New User').should('be.visible')
    cy.get('body').type('{esc}')
    cy.contains('[role="dialog"]', 'Add New User').should('not.exist')
  })

  it('UI-SCREEN-USERS-003 opens edit/delete row action dialogs with selected row context', () => {
    cy.get('tbody tr').first().find('td').eq(1).invoke('text').then((usernameRaw) => {
      const username = usernameRaw.trim()

      cy.contains('span', 'Open menu').first().parents('button').first().click()
      cy.contains('[role="menuitem"]', 'Edit').click()
      cy.contains('[role="dialog"]', 'Edit User').should('be.visible')
      cy.get('body').type('{esc}')

      cy.contains('span', 'Open menu').first().parents('button').first().click()
      cy.contains('[role="menuitem"]', 'Delete').click()
      cy.contains('[role="alertdialog"], [role="dialog"]', 'Delete User')
        .should('be.visible')
        .as('deleteDialog')
      cy.get('@deleteDialog').contains(username).should('exist')
      cy.get('@deleteDialog').contains('button', 'Cancel').click()
      cy.contains('Delete User').should('not.exist')
    })
  })
})
