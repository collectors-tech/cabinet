describe('ui-foundation-shell-navigation', () => {
  function visibleByTestId(testId: string) {
    return cy.get(`[data-testid="${testId}"]`).first()
  }

  function signInTo(path: string) {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('input[name="email"]').clear().type('e2e-shell-nav@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-001 keeps nav/header visible while content container handles scroll', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    cy.get('header').should('have.class', 'sticky')
    cy.get('header').should('have.class', 'top-0')
    cy.get('[data-slot="sidebar-container"]').then(($sidebarBefore) => {
      const topBefore = Math.round(
        $sidebarBefore[0].getBoundingClientRect().top
      )

      cy.scrollTo('bottom', { ensureScrollable: false })

      cy.get('[data-slot="sidebar-container"]').then(($sidebarAfter) => {
        const topAfter = Math.round($sidebarAfter[0].getBoundingClientRect().top)
        expect(Math.abs(topAfter - topBefore)).to.be.lte(2)
      })
    })
    cy.get('[data-slot="sidebar-container"]').should('be.visible')
    cy.get('[data-slot="sidebar-header"]').should('be.visible')
    cy.get('header').should('be.visible')
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-004 renders app version and build date metadata in sidebar footer', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: '2.3.0-e2e',
        build_date: '2026-03-02T00:00:00Z',
      },
    }).as('runtimeMeta')

    signInTo('/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/?$/)
    cy.wait('@runtimeMeta')

    cy.get('[data-testid="sidebar-runtime-meta"]').should('be.visible')
    cy.get('[data-testid="sidebar-app-version"]').should('have.text', '2.3.0-e2e')
    cy.get('[data-testid="sidebar-build-date"]').should(
      'have.text',
      '2026-03-02T00:00:00Z'
    )
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-002 supports nav edit mode reorder and visibility persistence', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    visibleByTestId('sidebar-nav-group-general').within(() => {
      cy.get('[data-testid^="sidebar-nav-link-"]')
        .then(($links) =>
          [...$links].map((link) => link.getAttribute('data-testid') || '')
        )
        .should('deep.equal', [
          'sidebar-nav-link-dashboard',
          'sidebar-nav-link-inventory',
          'sidebar-nav-link-collections',
          'sidebar-nav-link-wishlist',
          'sidebar-nav-link-discoveries',
          'sidebar-nav-link-scanner',
          'sidebar-nav-link-integrations',
          'sidebar-nav-link-chats',
          'sidebar-nav-link-users',
          'sidebar-nav-link-reports',
        ])
    })

    visibleByTestId('sidebar-nav-edit-toggle').click()
    visibleByTestId('sidebar-nav-edit-panel').should('be.visible')
    visibleByTestId('sidebar-nav-edit-panel').within(() => {
      cy.get('[data-testid="sidebar-nav-move-up-wishlist"]').click({
        force: true,
      })
      cy.get('[data-testid^="sidebar-nav-edit-item-"]')
        .filter(':visible')
        .should(($items) => {
          const ids = [...$items].map(
            (item) => item.getAttribute('data-testid') || ''
          )
          expect(ids).to.deep.equal([
            'sidebar-nav-edit-item-dashboard',
            'sidebar-nav-edit-item-inventory',
            'sidebar-nav-edit-item-wishlist',
            'sidebar-nav-edit-item-collections',
            'sidebar-nav-edit-item-discoveries',
            'sidebar-nav-edit-item-scanner',
            'sidebar-nav-edit-item-integrations',
            'sidebar-nav-edit-item-chats',
            'sidebar-nav-edit-item-users',
            'sidebar-nav-edit-item-reports',
          ])
        })
      cy.get('[data-testid="sidebar-nav-visibility-integrations"]').click()
    })
    visibleByTestId('sidebar-nav-edit-toggle').click()

    visibleByTestId('sidebar-nav-group-general').within(() => {
      cy.get('[data-testid^="sidebar-nav-link-"]')
        .then(($links) =>
          [...$links].map((link) => link.getAttribute('data-testid') || '')
        )
        .should('deep.equal', [
          'sidebar-nav-link-dashboard',
          'sidebar-nav-link-inventory',
          'sidebar-nav-link-wishlist',
          'sidebar-nav-link-collections',
          'sidebar-nav-link-discoveries',
          'sidebar-nav-link-scanner',
          'sidebar-nav-link-chats',
          'sidebar-nav-link-users',
          'sidebar-nav-link-reports',
        ])
    })

    cy.reload()
    visibleByTestId('sidebar-nav-group-general').within(() => {
      cy.get('[data-testid^="sidebar-nav-link-"]')
        .then(($links) =>
          [...$links].map((link) => link.getAttribute('data-testid') || '')
        )
        .should('deep.equal', [
          'sidebar-nav-link-dashboard',
          'sidebar-nav-link-inventory',
          'sidebar-nav-link-wishlist',
          'sidebar-nav-link-collections',
          'sidebar-nav-link-discoveries',
          'sidebar-nav-link-scanner',
          'sidebar-nav-link-chats',
          'sidebar-nav-link-users',
          'sidebar-nav-link-reports',
        ])
    })
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-007 reflects live nav edit order and saves the exact shown order', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    visibleByTestId('sidebar-nav-edit-toggle').click()
    visibleByTestId('sidebar-nav-edit-panel').should('be.visible')

    visibleByTestId('sidebar-nav-edit-panel').within(() => {
      cy.get('[data-testid="sidebar-nav-move-up-wishlist"]').click({
        force: true,
      })
      cy.get('[data-testid^="sidebar-nav-edit-item-"]')
        .filter(':visible')
        .should(($items) => {
          const ids = [...$items].map(
            (item) => item.getAttribute('data-testid') || ''
          )
          expect(ids).to.deep.equal([
            'sidebar-nav-edit-item-dashboard',
            'sidebar-nav-edit-item-inventory',
            'sidebar-nav-edit-item-wishlist',
            'sidebar-nav-edit-item-collections',
            'sidebar-nav-edit-item-discoveries',
            'sidebar-nav-edit-item-scanner',
            'sidebar-nav-edit-item-integrations',
            'sidebar-nav-edit-item-chats',
            'sidebar-nav-edit-item-users',
            'sidebar-nav-edit-item-reports',
          ])
        })
      cy.get('[data-testid="sidebar-nav-move-up-wishlist"]').click({
        force: true,
      })

      cy.get('[data-testid^="sidebar-nav-edit-item-"]')
        .filter(':visible')
        .should(($items) => {
          const ids = [...$items].map(
            (item) => item.getAttribute('data-testid') || ''
          )
          expect(ids).to.deep.equal([
            'sidebar-nav-edit-item-dashboard',
            'sidebar-nav-edit-item-wishlist',
            'sidebar-nav-edit-item-inventory',
            'sidebar-nav-edit-item-collections',
            'sidebar-nav-edit-item-discoveries',
            'sidebar-nav-edit-item-scanner',
            'sidebar-nav-edit-item-integrations',
            'sidebar-nav-edit-item-chats',
            'sidebar-nav-edit-item-users',
            'sidebar-nav-edit-item-reports',
          ])
        })
    })

    visibleByTestId('sidebar-nav-edit-toggle').click()
    visibleByTestId('sidebar-nav-group-general')
      .find('[data-testid^="sidebar-nav-link-"]')
      .then(($links) => [...$links].map((link) => link.getAttribute('data-testid') || ''))
      .should('deep.equal', [
        'sidebar-nav-link-dashboard',
        'sidebar-nav-link-wishlist',
        'sidebar-nav-link-inventory',
        'sidebar-nav-link-collections',
        'sidebar-nav-link-discoveries',
        'sidebar-nav-link-scanner',
        'sidebar-nav-link-integrations',
        'sidebar-nav-link-chats',
        'sidebar-nav-link-users',
        'sidebar-nav-link-reports',
      ])
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-003 updates collection context label when folder selection changes', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    cy.get('[data-testid="collection-context-label"]').should(
      'contain',
      'All Items'
    )
    cy.get('[data-testid="collection-active-context"]').should(
      'have.text',
      'All Items'
    )

    cy.get('[data-testid="collection-folder-store-1"]').click()
    cy.get('[data-testid="collection-context-label"]').should(
      'contain',
      'Store 1'
    )
    cy.get('[data-testid="collection-active-context"]').should(
      'have.text',
      'Store 1'
    )
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-005 keeps sidebar top area DB/profile switcher only', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    visibleByTestId('team-switcher-trigger').should('be.visible')
    cy.get('[data-testid="workspace-collections-panel"]').should('not.exist')
    cy.get('[data-testid="workspace-collection-item-all-items"]').should(
      'not.exist'
    )
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-006 manages collections from Collections section and inline picker quick-create', () => {
    signInTo('/collections/')
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/collections\/?$/
    )

    visibleByTestId('collections-section').should('be.visible')
    visibleByTestId('collections-new-input').clear().type('Quick Create Shelf')
    visibleByTestId('collections-new-save').click()
    visibleByTestId('collections-item-quick-create-shelf').should('be.visible')

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    visibleByTestId('collection-inline-picker').should('be.visible')
    visibleByTestId('collection-inline-add-new').click()
    visibleByTestId('collection-inline-new-name').clear().type('Inline Auto Select')
    visibleByTestId('collection-inline-save').click()
    visibleByTestId('collection-inline-picker-selected').should(
      'contain',
      'Inline Auto Select'
    )
    visibleByTestId('collection-inline-picker-option-inline-auto-select').should(
      'be.visible'
    )
  })
})
