describe('general/ui-shell-workspaces', () => {
  function openInventory() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
      shellWorkspace: 'navigation',
    })
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/inventory\/?$/
    )
    cy.get('[data-testid="shell-workspace-icon-rail"]').should('be.visible')
    cy.get('[data-testid="shell-workspace-bell"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Open notification inbox')
    cy.get('[data-testid="shell-workspace-label"]').should('not.exist')
    cy.get('[data-testid="sidebar-nav-link-inventory"]').should('be.visible')
  }

  it('UI-SHELL-WORKSPACES-001 switches Navigation, Search, and Chat with an icon-only rail', () => {
    openInventory()
    cy.get('[data-testid="shell-workspace-navigation"]').should(
      'have.attr',
      'data-active',
      'true'
    )
    cy.get('[data-testid="sidebar-nav-link-inventory"]').should('be.visible')
    cy.get('[data-testid="shell-workspace-icon-rail"]')
      .should('be.visible')
      .within(() => {
        cy.get('[data-testid="shell-workspace-navigation"]')
          .should('have.attr', 'aria-label', 'Navigation workspace')
          .and('have.attr', 'title', 'Navigation workspace')
        cy.get('[data-testid="shell-workspace-search"]')
          .should('have.attr', 'aria-label', 'Search workspace')
          .and('have.attr', 'title', 'Search workspace')
        cy.get('[data-testid="shell-workspace-assistant"]')
          .should('have.attr', 'aria-label', 'Chat workspace')
          .and('have.attr', 'title', 'Chat workspace')
        cy.get('[data-testid="shell-workspace-bell"]')
          .should('have.attr', 'aria-label', 'Open notification inbox')
          .and('have.attr', 'title', 'Open notification inbox')
        cy.contains('Nav').should('not.exist')
        cy.contains('Search').should('not.exist')
        cy.contains('Chat').should('not.exist')
        cy.contains('Assistant').should('not.exist')
        cy.contains('Inbox').should('not.exist')
      })
    cy.get('[data-testid="shell-workspace-switcher"]').should(
      'not.contain',
      'Workspace'
    )
    cy.get('[data-testid="shell-workspace-icon-rail"] [data-active="true"]')
      .should('have.length', 1)

    cy.get('[data-testid="shell-workspace-search"]').click()
    cy.get('[data-testid="shell-search-workspace"]').should('be.visible')
    cy.get('[data-testid="shell-search-workspace-input"]')
      .should('be.visible')
      .and('be.focused')
    cy.get('[data-testid="shell-workspace-search"]').should(
      'have.attr',
      'data-active',
      'true'
    )
    cy.get('[data-testid="shell-workspace-icon-rail"] [data-active="true"]')
      .should('have.length', 1)
    cy.get('[data-testid="shell-search-nav-results"]').should('be.visible')
    cy.get('[data-testid="shell-search-nav-result"]')
      .first()
      .should('contain', 'Dashboard')
      .and('contain', 'General')
    cy.get('[data-testid="shell-search-nav-result"]').should(
      'contain',
      'Settings / Profile'
    )
    cy.location('pathname').should('match', /^\/inventory\/?$/)

    cy.get('[data-testid="shell-workspace-assistant"]').click()
    cy.get('[data-testid="shell-workspace-assistant"]').should(
      'have.attr',
      'data-active',
      'true'
    )
    cy.get('[data-testid="shell-workspace-icon-rail"] [data-active="true"]')
      .should('have.length', 1)
    cy.get('[data-testid="shell-assistant-workspace"]').should('exist')
    cy.get('[data-testid="shell-assistant-compose-input"]').should('exist')

    cy.get('[data-testid="shell-workspace-icon-rail"] [data-testid="shell-workspace-bell"]')
      .should('have.attr', 'aria-label', 'Open notification inbox')
      .and('have.attr', 'title', 'Open notification inbox')
      .within(() => {
        cy.get('[data-testid="shell-workspace-bell-badge"]').should('be.visible')
      })
    cy.get(
      '[data-testid="shell-workspace-icon-rail"] [data-testid="shell-workspace-bell"]'
    ).click({ force: true })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inbox\/?$/)
    cy.get('[data-testid="shell-workspace-inbox"]').should('not.exist')

    cy.get('[data-testid="shell-workspace-bell"]').should(
      'have.attr',
      'data-active',
      'true'
    )
    cy.get('[data-testid="shell-workspace-icon-rail"] [data-active="true"]')
      .should('have.length', 1)
  })

  it('UI-SHELL-WORKSPACES-002 activates Assistant workspace from header without route loss', () => {
    openInventory()
    cy.location('pathname').then((initialPathname) => {
      cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'false')
      cy.get('[data-testid="shell-chat-toggle"]').click({ force: true })
      cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
      cy.get('[data-testid="shell-assistant-workspace"]').should('exist')
      cy.get('[data-testid="shell-assistant-route-context"]').should('contain', '/inventory')
      cy.location('pathname').should('eq', initialPathname)
      cy.get('[data-testid="shell-chat-toggle"]').click({ force: true })
      cy.get('[data-testid="shell-workspace-navigation"]').should('have.attr', 'data-active', 'true')
      cy.get('[data-testid="shell-assistant-workspace"]').should('not.exist')
      cy.location('pathname').should('eq', initialPathname)
    })
  })

  it('UI-SHELL-WORKSPACES-003 preserves Assistant workspace across authenticated route changes', () => {
    openInventory()
    cy.get('[data-testid="shell-chat-toggle"]').click({ force: true })
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="shell-assistant-route-context"]').should('contain', '/inventory')

    cy.visit('/wishlist')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/wishlist\/?$/)
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="shell-assistant-workspace"]').should('exist')
    cy.get('[data-testid="shell-assistant-route-context"]').should('contain', '/wishlist')
  })

  it('UI-SHELL-WORKSPACES-004 keeps Assistant and /chats semantics distinct while Inbox is bell-routed', () => {
    openInventory()
    cy.get('[data-testid="shell-workspace-assistant"]').click()
    cy.get('[data-testid="shell-assistant-workspace"]').should('exist')
    cy.get('[data-testid="shell-assistant-compose-input"]').should('exist')

    cy.get('[data-testid="shell-workspace-icon-rail"] [data-testid="shell-workspace-bell"]').click({ force: true })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inbox\/?$/)
    cy.contains('Notification Inbox').should('be.visible')
    cy.get('[data-testid="shell-workspace-inbox"]').should('not.exist')

    cy.visit('/chats')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
    cy.contains(
      'Persistent profile-scoped conversation threads backed by Cabinet runtime.'
    ).should('exist')
    cy.contains(
      'Use Assistant for AI-guided help and actions; use Chats for durable conversation threads.'
    ).should('exist')
  })

  it('UI-SHELL-WORKSPACES-005 opens the durable Inbox page from the bell-only top affordance', () => {
    openInventory()
    cy.get('[data-testid="shell-workspace-icon-rail"] [data-testid="shell-workspace-bell"]').click({ force: true })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inbox\/?$/)
    cy.get('[data-testid="notification-inbox-page"]').should('be.visible')
    cy.get('[data-testid="notification-inbox-filters"]').should('be.visible')
    cy.get('[data-testid="notification-inbox-list-pane"]').should('be.visible')
    cy.get('[data-testid="notification-inbox-detail-pane"]').should('be.visible')
  })

  it('UI-SHELL-WORKSPACES-006 persists the Assistant workspace to the active profile across reload and section changes', () => {
    openInventory()
    cy.get('[data-testid="shell-workspace-assistant"]').click()
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.window().its('localStorage').invoke('getItem', 'cabinet.shell.workspace.active.e2e-profile-001').should('eq', 'assistant')

    cy.reload()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="shell-assistant-workspace"]').should('exist')

    cy.visit('/settings/profile')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/settings\/profile\/?$/)
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.contains('Profile settings').should('be.visible')
  })

  it('UI-SHELL-WORKSPACES-007 persists Search workspace as a real shell panel', () => {
    openInventory()
    cy.get('[data-testid="shell-workspace-search"]').click()
    cy.get('[data-testid="shell-search-workspace"]').should('be.visible')
    cy.window()
      .its('localStorage')
      .invoke(
        'getItem',
        'cabinet.shell.workspace.active.e2e-profile-001'
      )
      .should('eq', 'search')

    cy.reload()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/inventory\/?$/
    )
    cy.get('[data-testid="shell-workspace-search"]').should(
      'have.attr',
      'data-active',
      'true'
    )
    cy.get('[data-testid="shell-search-workspace"]').should('be.visible')
  })

  it('UI-SHELL-WORKSPACES-008/#1456 filters dense navigation results and navigates from Search workspace', () => {
    openInventory()
    cy.get('[data-testid="shell-workspace-search"]').click()
    cy.get('[data-testid="shell-search-workspace"]').should('be.visible')
    cy.get('[data-testid="shell-workspace-search"]').should(
      'have.attr',
      'data-active',
      'true'
    )
    cy.get('[data-testid="shell-search-workspace-input"]')
      .should('have.attr', 'placeholder', 'Search nav, settings, help...')
      .type('appearance')
    cy.get('[data-testid="shell-search-nav-result"]')
      .should('have.length', 1)
      .and('contain', 'Settings / Appearance')
      .and('contain', 'Other · /settings/appearance')
      .click()

    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/appearance\/?$/
    )
    cy.get('[data-testid="shell-workspace-search"]').should(
      'have.attr',
      'data-active',
      'true'
    )
  })
})
