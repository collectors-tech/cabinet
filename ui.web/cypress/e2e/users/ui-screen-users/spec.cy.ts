describe('ui-screen-users', () => {
  function visibleCellText(raw: string) {
    const value = raw.trim()
    const midpoint = value.length / 2
    if (
      value.length % 2 === 0 &&
      value.slice(0, midpoint) === value.slice(midpoint)
    ) {
      return value.slice(0, midpoint)
    }
    return value
  }

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

    cy.get('tbody tr').first().find('td').eq(1).invoke('text').then(() => {
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

  it('UI-SCREEN-USERS-003 opens the real edit dialog for the double-clicked user row', () => {
    cy.get('tbody tr').should('have.length.greaterThan', 1)

    cy.get('tbody tr')
      .eq(0)
      .find('td')
      .eq(1)
      .invoke('text')
      .then((firstUsernameRaw) => {
        const firstUsername = visibleCellText(firstUsernameRaw)

        cy.get('tbody tr').eq(0).click()
        cy.contains('[data-testid="users-row-details-modal"]', firstUsername)
          .should('be.visible')
        cy.get('body').type('{esc}')
      })

    cy.get('tbody tr')
      .eq(1)
      .find('td')
      .eq(1)
      .invoke('text')
      .then((secondUsernameRaw) => {
        const secondUsername = visibleCellText(secondUsernameRaw)

        cy.get('tbody tr').eq(1).dblclick()
        cy.contains('[role="dialog"]', 'Edit User')
          .should('be.visible')
          .within(() => {
            cy.get('input[placeholder="john_doe"]').should(
              'have.value',
              secondUsername
            )
          })
      })
  })

  it('UI-SCREEN-USERS-003 persists edit saves through Cabinet API and refreshes the edited row', () => {
    const originalUser = {
      id: 'edit-user-001',
      firstName: 'Editable',
      lastName: 'Collector',
      username: 'editable_collector',
      email: 'editable.collector@example.com',
      phoneNumber: '+61000000002',
      status: 'active',
      role: 'view',
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    }
    const editedUser = {
      ...originalUser,
      firstName: 'Updated',
      username: 'updated_collector',
      email: 'updated.collector@example.com',
      updatedAt: '2026-01-02T00:00:00Z',
    }
    let editSaved = false

    cy.intercept('GET', '/api/users*', (req) => {
      req.reply({
        statusCode: 200,
        body: { users: [editSaved ? editedUser : originalUser] },
      })
    }).as('editUsersList')
    cy.intercept('PUT', '/api/users/edit-user-001', (req) => {
      expect(req.body).to.include({
        firstName: 'Updated',
        lastName: 'Collector',
        username: 'updated_collector',
        email: 'updated.collector@example.com',
        phoneNumber: '+61000000002',
        role: 'view',
        status: 'active',
      })
      editSaved = true
      req.reply({
        statusCode: 200,
        body: { user: editedUser },
      })
    }).as('editUser')

    cy.reload()
    cy.wait('@editUsersList')
    cy.contains('editable_collector').should('be.visible')

    cy.get('tbody tr').first().dblclick()
    cy.contains('[role="dialog"]', 'Edit User')
      .should('be.visible')
      .within(() => {
        cy.get('input[placeholder="John"]').clear().type('Updated')
        cy.get('input[placeholder="john_doe"]')
          .clear()
          .type('updated_collector')
        cy.get('input[placeholder="john.doe@gmail.com"]')
          .clear()
          .type('updated.collector@example.com')
        cy.contains('button', 'Save changes').click()
      })

    cy.wait('@editUser').its('response.statusCode').should('eq', 200)
    cy.wait('@editUsersList')
    cy.contains('[role="dialog"]', 'Edit User').should('not.exist')
    cy.contains('updated_collector').should('be.visible')
    cy.contains('updated.collector@example.com').should('be.visible')
    cy.contains('editable_collector').should('not.exist')

    cy.reload()
    cy.wait('@editUsersList')
    cy.contains('updated_collector').should('be.visible')
    cy.contains('updated.collector@example.com').should('be.visible')
  })

  it('UI-SCREEN-USERS-004 retries users list after a fetch failure', () => {
    let listAttempt = 0

    cy.intercept('GET', '/api/users*', (req) => {
      listAttempt += 1
      if (listAttempt === 1) {
        req.reply({
          statusCode: 503,
          body: { error: 'users_unavailable' },
        })
        return
      }

      req.reply({
        statusCode: 200,
        body: {
          users: [
            {
              id: 'retry-user-001',
              firstName: 'Retry',
              lastName: 'User',
              username: 'retry_user',
              email: 'retry.user@example.com',
              phoneNumber: '+61000000001',
              status: 'active',
              role: 'admin',
              createdAt: '2026-01-01T00:00:00Z',
              updatedAt: '2026-01-01T00:00:00Z',
            },
          ],
        },
      })
    }).as('retryUsersList')

    cy.reload()
    cy.wait('@retryUsersList').its('response.statusCode').should('eq', 503)
    cy.get('[data-testid="users-load-error"]').should('be.visible')
    cy.contains('Users load failed').should('be.visible')
    cy.contains('users_fetch_failed_503').should('be.visible')
    cy.get('[data-testid="users-load-error"]')
      .contains('button', 'Retry')
      .click()
    cy.wait('@retryUsersList').its('response.statusCode').should('eq', 200)

    cy.get('[data-testid="users-load-error"]').should('not.exist')
    cy.contains('h2', 'User List').should('be.visible')
    cy.contains('retry_user').should('be.visible')
  })

  it('UI-SCREEN-USERS-006 renders deterministic loading and empty states', () => {
    cy.intercept('GET', '/api/users*', (req) => {
      req.reply({
        delay: 1000,
        statusCode: 200,
        body: { users: [] },
      })
    }).as('emptyUsersList')

    cy.reload()
    cy.contains('h2', 'User List').should('be.visible')
    cy.contains('button', 'Invite User').should('be.visible')
    cy.contains('button', 'Add User').should('be.visible')
    cy.contains('Loading users...').should('be.visible')

    cy.wait('@emptyUsersList').its('response.statusCode').should('eq', 200)

    cy.contains('Loading users...').should('not.exist')
    cy.get('[data-testid="users-load-error"]').should('not.exist')
    cy.contains('No results.').should('be.visible')
    cy.contains('retry_user').should('not.exist')
  })

  it('UI-SCREEN-USERS-007 hides optional table columns from the View menu', () => {
    cy.viewport(1280, 900)
    cy.contains('h2', 'User List').should('be.visible')
    cy.contains('th', 'Username').should('be.visible')
    cy.contains('th', 'Email').should('be.visible')
    cy.contains('th', 'Status').should('be.visible')
    cy.get('tbody tr').first().find('td').eq(1).invoke('text').then((usernameRaw) => {
      const username = visibleCellText(usernameRaw)

      cy.contains('button', 'View').click()
      cy.contains('[role="menuitemcheckbox"]', 'email').click()
      cy.contains('th', 'Email').should('not.exist')
      cy.contains('th', 'Username').should('be.visible')
      cy.contains('th', 'Status').should('be.visible')
      cy.contains(username).should('be.visible')
      cy.location('pathname').should('match', /^\/users\/?$/)
    })
  })

  it('UI-SCREEN-USERS-008 persists bulk invite, status, and delete actions through Cabinet API', () => {
    let users = [
      {
        id: 'bulk-user-001',
        firstName: 'Bulk',
        lastName: 'Alpha',
        username: 'bulk_alpha',
        email: 'bulk.alpha@example.com',
        phoneNumber: '+61000000101',
        status: 'active',
        role: 'view',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
      },
      {
        id: 'bulk-user-002',
        firstName: 'Bulk',
        lastName: 'Beta',
        username: 'bulk_beta',
        email: 'bulk.beta@example.com',
        phoneNumber: '+61000000102',
        status: 'active',
        role: 'admin',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
      },
    ]

    cy.intercept('GET', '/api/users*', (req) => {
      req.reply({ statusCode: 200, body: { users } })
    }).as('bulkUsersList')
    cy.intercept('POST', '/api/users/invite', (req) => {
      expect(req.body).to.include({ desc: 'Bulk invitation' })
      expect(['bulk.alpha@example.com', 'bulk.beta@example.com']).to.include(
        req.body.email
      )
      req.reply({ statusCode: 201, body: { user: req.body } })
    }).as('bulkInviteUser')
    cy.intercept('PUT', '/api/users/*', (req) => {
      const id = req.url.split('/').pop()
      expect(req.body).to.include({ status: 'inactive' })
      users = users.map((user) =>
        user.id === id ? { ...user, status: 'inactive' } : user
      )
      req.reply({
        statusCode: 200,
        body: { user: users.find((user) => user.id === id) },
      })
    }).as('bulkStatusUser')
    cy.intercept('DELETE', '/api/users/*', (req) => {
      const id = req.url.split('/').pop()
      users = users.filter((user) => user.id !== id)
      req.reply({ statusCode: 204 })
    }).as('bulkDeleteUser')

    cy.reload()
    cy.wait('@bulkUsersList')
    cy.contains('bulk_alpha').should('be.visible')
    cy.contains('bulk_beta').should('be.visible')

    cy.get('tbody tr')
      .eq(0)
      .find('[role="checkbox"]')
      .first()
      .click({ force: true })
    cy.get('tbody tr')
      .eq(1)
      .find('[role="checkbox"]')
      .first()
      .click({ force: true })
    cy.get('button[aria-label="Invite selected users"]').click()
    cy.wait(['@bulkInviteUser', '@bulkInviteUser'])
    cy.wait('@bulkUsersList')

    cy.get('tbody tr')
      .eq(0)
      .find('[role="checkbox"]')
      .first()
      .click({ force: true })
    cy.get('tbody tr')
      .eq(1)
      .find('[role="checkbox"]')
      .first()
      .click({ force: true })
    cy.get('button[aria-label="Deactivate selected users"]').click()
    cy.wait(['@bulkStatusUser', '@bulkStatusUser'])
    cy.wait('@bulkUsersList')
    cy.contains('bulk_alpha')
      .parents('tr')
      .contains('inactive')
      .should('be.visible')
    cy.contains('bulk_beta')
      .parents('tr')
      .contains('inactive')
      .should('be.visible')

    cy.get('tbody tr')
      .eq(0)
      .find('[role="checkbox"]')
      .first()
      .click({ force: true })
    cy.get('tbody tr')
      .eq(1)
      .find('[role="checkbox"]')
      .first()
      .click({ force: true })
    cy.get('button[aria-label="Delete selected users"]').click()
    cy.contains('[role="alertdialog"], [role="dialog"]', 'Delete 2 users')
      .should('be.visible')
      .within(() => {
        cy.get('input[placeholder=\'Type "DELETE" to confirm.\']').type('DELETE')
        cy.contains('button', 'Delete').click()
      })
    cy.wait(['@bulkDeleteUser', '@bulkDeleteUser'])
    cy.wait('@bulkUsersList')
    cy.contains('bulk_alpha').should('not.exist')
    cy.contains('bulk_beta').should('not.exist')
    cy.contains('No results.').should('be.visible')
  })
})
