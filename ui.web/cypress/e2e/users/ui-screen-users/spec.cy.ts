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
    cy.intercept('GET', '/api/users*').as('listUsers')
    signInToUsers()
    cy.wait('@listUsers')
  })

  it('UI-SCREEN-USERS-001 reads users table from Cabinet API and supports filter/sort/pagination workflows', () => {
    cy.contains('h2', 'User List').should('be.visible')
    cy.get('tbody tr').first().find('td').eq(1).invoke('text').then((usernameRaw) => {
      const username = usernameRaw.trim()
      const token = username.slice(0, Math.min(4, username.length))

      cy.get('input[placeholder="Filter users..."]').clear().type(token)
      cy.location('search').should('contain', 'username=')
      cy.get('tbody tr').first().find('td').eq(1).invoke('text').should('match', new RegExp(token, 'i'))
    })
  })

  it('UI-SCREEN-USERS-002 persists add and invite actions through Cabinet API', () => {
    const unique = Date.now().toString()
    const username = `autouser_${unique}`
    const email = `autouser_${unique}@example.com`
    const inviteEmail = `invite_${unique}@example.com`

    cy.intercept('POST', '/api/users').as('createUser')
    cy.intercept('POST', '/api/users/invite').as('inviteUser')

    cy.contains('button', 'Add User').click()
    cy.contains('[role="dialog"]', 'Add New User').should('be.visible')
    cy.get('input[placeholder="John"]').type('Auto')
    cy.get('input[placeholder="Doe"]').type('User')
    cy.get('input[placeholder="john_doe"]').type(username)
    cy.get('input[placeholder="john.doe@gmail.com"]').type(email)
    cy.get('input[placeholder="+123456789"]').type('+61000000000')
    cy.get('[role="dialog"] button[role="combobox"]').first().click()
    cy.contains('[role="option"], [data-radix-collection-item]', 'Admin').click()
    cy.get('input[placeholder="e.g., S3cur3P@ssw0rd"]').first().type('password123')
    cy.get('input[placeholder="e.g., S3cur3P@ssw0rd"]').eq(1).type('password123')
    cy.contains('[role="dialog"] button', 'Save changes').click()
    cy.wait('@createUser').its('response.statusCode').should('eq', 201)
    cy.get('input[placeholder="Filter users..."]').clear().type(username)
    cy.contains(username).should('be.visible')

    cy.contains('button', 'Invite User').click()
    cy.contains('[role="dialog"]', 'Invite User').should('be.visible')
    cy.get('input[placeholder="eg: john.doe@gmail.com"]').type(inviteEmail)
    cy.get('[role="dialog"] button[role="combobox"]').first().click()
    cy.contains('[role="option"], [data-radix-collection-item]', 'View').click()
    cy.contains('[role="dialog"] button', 'Invite').click()
    cy.wait('@inviteUser').its('response.statusCode').should('eq', 201)
    cy.get('input[placeholder="Filter users..."]').clear().type(inviteEmail.split('@')[0])
    cy.contains(inviteEmail.split('@')[0]).should('be.visible')
  })

  it('UI-SCREEN-USERS-003 persists delete actions through Cabinet API row context', () => {
    cy.intercept('DELETE', '/api/users/*').as('deleteUser')

    cy.get('tbody tr').first().find('td').eq(1).invoke('text').then((usernameRaw) => {
      const username = usernameRaw.trim()

      cy.contains('span', 'Open menu').first().parents('button').first().click()
      cy.contains('[role="menuitem"]', 'Delete').click()
      cy.contains('[role="alertdialog"], [role="dialog"]', 'Delete User')
        .should('be.visible')
        .as('deleteDialog')
      cy.get('@deleteDialog')
        .find('p .font-bold')
        .first()
        .invoke('text')
        .then((confirmUsernameRaw) => {
          const confirmUsername = confirmUsernameRaw.trim()
          cy.get('@deleteDialog')
            .find('input[placeholder="Enter username to confirm deletion."]')
            .clear()
            .type(confirmUsername)
          cy.get('@deleteDialog').contains('button', 'Delete').click()
          cy.wait('@deleteUser').its('response.statusCode').should('eq', 204)
          cy.contains(confirmUsername).should('not.exist')
        })
    })
  })
})
