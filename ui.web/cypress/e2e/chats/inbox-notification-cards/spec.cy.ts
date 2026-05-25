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
            id: 'inbox-telegram-capture',
            profile_id: 'e2e-profile-001',
            thread_id: 'thread-telegram-capture',
            source: 'telegram_catalog_capture',
            status: 'unread',
            title: 'Telegram capture needs review',
            summary:
              'Review the Telegram photo and barcode draft before applying it.',
            metadata: {
              review_url:
                '/chats?profile_id=e2e-profile-001&thread_id=thread-telegram-capture&preview_id=preview-telegram-capture',
              preview_id: 'preview-telegram-capture',
              confirmation_state: 'preview_required',
              telegram_reply: {
                review_url:
                  '/chats?profile_id=e2e-profile-001&thread_id=thread-telegram-capture&preview_id=preview-telegram-capture',
                confirmation_state: 'preview_required',
              },
            },
            created_at: '2026-04-29T09:00:00Z',
            updated_at: '2026-04-29T09:00:00Z',
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
      .should('have.length', 2)
      .first()
      .scrollIntoView({ block: 'center' })
      .should('be.visible')
    cy.get('[data-testid="shell-inbox-workspace"]')
      .invoke('text')
      .should('include', 'Price moved on Showcase Seed One')
      .and('include', 'Market price climbed 12% since the last check.')
      .and('include', '2 days old')
      .and('include', 'SHW-100')
      .and('not.include', 'Already handled')
    cy.contains(
      '[data-testid="shell-inbox-notification-card"]',
      'Telegram capture needs review'
    )
      .should('exist')
      .within(() => {
        cy.root().invoke('text').should('include', 'Telegram Catalog Capture')
        cy.get('[data-testid="shell-inbox-item-link"]')
          .should('contain', 'Review Telegram capture')
          .and(
            'have.attr',
            'href',
            '/chats?profile_id=e2e-profile-001&thread_id=thread-telegram-capture&preview_id=preview-telegram-capture'
          )
      })
    cy.contains(
      '[data-testid="shell-inbox-notification-card"]',
      'Price moved on Showcase Seed One'
    )
      .find('[data-testid="shell-inbox-item-link"]')
      .should('have.attr', 'href', '/inventory/?item=shw-100')

    cy.contains(
      '[data-testid="shell-inbox-notification-card"]',
      'Price moved on Showcase Seed One'
    )
      .scrollIntoView({ block: 'center' })
      .find('[data-testid="shell-inbox-mark-read"]')
      .click()
    cy.wait('@updateInboxStatus').its('request.body').should('deep.include', {
      profile_id: 'e2e-profile-001',
      status: 'read',
    })
    cy.contains(
      '[data-testid="shell-inbox-notification-card"]',
      'Price moved on Showcase Seed One'
    )
      .find('[data-testid="shell-inbox-item-status"]')
      .should('contain', 'read')

    cy.contains(
      '[data-testid="shell-inbox-notification-card"]',
      'Price moved on Showcase Seed One'
    )
      .scrollIntoView({ block: 'center' })
      .find('[data-testid="shell-inbox-mark-unread"]')
      .click()
    cy.wait('@updateInboxStatus').its('request.body').should('deep.include', {
      profile_id: 'e2e-profile-001',
      status: 'unread',
    })
    cy.contains(
      '[data-testid="shell-inbox-notification-card"]',
      'Price moved on Showcase Seed One'
    )
      .find('[data-testid="shell-inbox-item-status"]')
      .should('contain', 'unread')

    cy.contains(
      '[data-testid="shell-inbox-notification-card"]',
      'Price moved on Showcase Seed One'
    )
      .scrollIntoView({ block: 'center' })
      .find('[data-testid="shell-inbox-archive"]')
      .click()
    cy.wait('@updateInboxStatus').its('request.body').should('deep.include', {
      profile_id: 'e2e-profile-001',
      status: 'archived',
    })
    cy.get('[data-testid="shell-inbox-notification-card"]').should(
      'have.length',
      1
    )
    cy.contains(
      '[data-testid="shell-inbox-notification-card"]',
      'Telegram capture needs review'
    ).should('exist')
  })

  it('INBOX-NOTIFICATIONS-002 opens Telegram capture review URLs on the requested chat thread', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/chat/threads?profile_id=*', {
      statusCode: 200,
      body: {
        threads: [
          {
            id: 'thread-general',
            profile_id: 'e2e-profile-001',
            title: 'General chat',
            created_at: '2026-04-28T09:00:00Z',
            updated_at: '2026-04-28T09:00:00Z',
          },
          {
            id: 'thread-telegram-capture',
            profile_id: 'e2e-profile-001',
            title: 'Telegram capture: Barcode draft',
            created_at: '2026-04-29T09:00:00Z',
            updated_at: '2026-04-29T09:00:00Z',
          },
        ],
      },
    }).as('loadThreads')
    cy.intercept(
      'GET',
      '/api/chat/messages?profile_id=*&thread_id=thread-telegram-capture',
      {
        statusCode: 200,
        body: {
          messages: [
            {
              id: 'message-telegram-capture',
              profile_id: 'e2e-profile-001',
              thread_id: 'thread-telegram-capture',
              role: 'user',
              content: 'Telegram capture with photo and barcode 9312345678901',
              created_at: '2026-04-29T09:00:00Z',
            },
          ],
        },
      }
    ).as('loadTelegramMessages')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/chats',
    })
    cy.visit(
      '/chats?profile_id=e2e-profile-001&thread_id=thread-telegram-capture&preview_id=preview-telegram-capture'
    )
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
    cy.wait('@loadThreads')
    cy.wait('@loadTelegramMessages')
    cy.get('[data-testid="chat-thread-title"]').should(
      'contain',
      'Telegram capture: Barcode draft'
    )
    cy.contains(
      '[data-testid="chat-message-list"]',
      'Telegram capture with photo and barcode 9312345678901'
    ).should('be.visible')
  })
})
