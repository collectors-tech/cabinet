describe('chats/inbox-notification-cards', () => {
  function openInboxWithStubbedNotifications() {
    cy.viewport(1400, 900)
    cy.clock(new Date('2026-04-29T12:00:00Z').getTime(), ['Date'])
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'inbox-price-move',
            profile_id: 'e2e-profile-001',
            thread_id: 'thread-price-move',
            source: 'pricing_workflow',
            status: 'unread',
            title: 'Price moved on Showcase Seed One',
            summary: 'Market price climbed 12% since the last check.',
            metadata: {
              item: {
                id: 'shw-100',
                part_number: 'SHW-100',
                title: 'Showcase Seed One',
                href: '/inventory/?item=shw-100',
              },
            },
            created_at: '2026-04-27T09:00:00Z',
            updated_at: '2026-04-27T09:00:00Z',
          },
          {
            id: 'inbox-old-archived',
            profile_id: 'e2e-profile-001',
            thread_id: 'thread-old',
            source: 'assistant_handoff',
            status: 'archived',
            title: 'Already handled',
            summary: 'This notification should stay out of the catch-up list.',
            created_at: '2026-04-20T09:00:00Z',
            updated_at: '2026-04-20T09:00:00Z',
          },
        ],
      },
    }).as('loadInbox')
    cy.intercept('PATCH', '/api/chat/inbox/inbox-price-move', (req) => {
      req.reply({
        statusCode: 200,
        body: {
          id: 'inbox-price-move',
          profile_id: 'e2e-profile-001',
          thread_id: 'thread-price-move',
          source: 'pricing_workflow',
          status: req.body.status,
          title: 'Price moved on Showcase Seed One',
          summary: 'Market price climbed 12% since the last check.',
          metadata: {
            item: {
              id: 'shw-100',
              part_number: 'SHW-100',
              title: 'Showcase Seed One',
              href: '/inventory/?item=shw-100',
            },
          },
          created_at: '2026-04-27T09:00:00Z',
          updated_at: '2026-04-29T12:00:00Z',
        },
      })
    }).as('updateInboxStatus')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
      shellWorkspace: 'inbox',
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.wait('@loadInbox')
  }

  it('INBOX-NOTIFICATIONS-001 shows catch-up cards and supports read, unread, archive actions', () => {
    openInboxWithStubbedNotifications()

    cy.get('[data-testid="shell-inbox-notification-card"]')
      .should('have.length', 1)
      .first()
      .scrollIntoView()
      .should('be.visible')
    cy.contains('[data-testid="shell-inbox-notification-card"]', 'Price moved on Showcase Seed One').should('be.visible')
    cy.contains('[data-testid="shell-inbox-notification-card"]', 'Market price climbed 12% since the last check.').should('be.visible')
    cy.contains('[data-testid="shell-inbox-notification-card"]', '2 days old').should('be.visible')
    cy.contains('[data-testid="shell-inbox-notification-card"]', 'SHW-100').should('be.visible')
    cy.contains('[data-testid="shell-inbox-workspace"]', 'Already handled').should('not.exist')
    cy.get('[data-testid="shell-inbox-item-link"]').should('have.attr', 'href', '/inventory/?item=shw-100')

    cy.get('[data-testid="shell-inbox-mark-read"]').click()
    cy.wait('@updateInboxStatus').its('request.body').should('deep.include', {
      profile_id: 'e2e-profile-001',
      status: 'read',
    })
    cy.get('[data-testid="shell-inbox-item-status"]').should('contain', 'read')

    cy.get('[data-testid="shell-inbox-mark-unread"]').click()
    cy.wait('@updateInboxStatus').its('request.body').should('deep.include', {
      profile_id: 'e2e-profile-001',
      status: 'unread',
    })
    cy.get('[data-testid="shell-inbox-item-status"]').should('contain', 'unread')

    cy.get('[data-testid="shell-inbox-archive"]').click()
    cy.wait('@updateInboxStatus').its('request.body').should('deep.include', {
      profile_id: 'e2e-profile-001',
      status: 'archived',
    })
    cy.get('[data-testid="shell-inbox-notification-card"]').should('not.exist')
    cy.contains('[data-testid="shell-inbox-workspace"]', 'Caught up.').should('be.visible')
  })
})
