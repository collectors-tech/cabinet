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

  function openCustomiseNav() {
    visibleByTestId('shell-workspace-menu-trigger').click()
    visibleByTestId('shell-workspace-menu-customise-nav').click()
    visibleByTestId('sidebar-nav-edit-panel').should('be.visible')
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

  it('UI-FOUNDATION-SHELL-NAVIGATION-005 keeps browser titles in Cabinet - <Page Title> format', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.title().should('eq', 'Cabinet - Inventory')

    visibleByTestId('sidebar-nav-link-integrations').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/integrations\/?$/)
    cy.title().should('eq', 'Cabinet - Integrations')

    visibleByTestId('sidebar-nav-link-dashboard').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/dashboard\/?$/)
    cy.title().should('eq', 'Cabinet - Home')
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-011 centers primary page titles with route icons and no inline context', () => {
    function assertCenteredHeader(testId: string, title: string) {
      cy.get(`[data-testid="${testId}-header-title"]`)
        .should('be.visible')
        .and('have.attr', 'data-centered', 'true')
        .and('contain', title)
      cy.get(`[data-testid="${testId}-page-icon"]`).should('be.visible')
      cy.get('header').should('not.contain', 'Active:')
      cy.get('header').should('not.contain', 'Collection:')
      cy.get('header').should('not.contain', 'Planning list')
    }

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    assertCenteredHeader('inventory', 'Inventory')

    visibleByTestId('sidebar-nav-link-collections').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/collections\/?$/)
    assertCenteredHeader('collections', 'Collections')

    visibleByTestId('sidebar-nav-link-wishlist').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/wishlist\/?$/)
    assertCenteredHeader('wishlist', 'Wishlist')
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
          'sidebar-nav-link-media',
          'sidebar-nav-link-collections',
          'sidebar-nav-link-wishlist',
          'sidebar-nav-link-discoveries',
          'sidebar-nav-link-market-watch',
          'sidebar-nav-link-inbox',
          'sidebar-nav-link-purchases',
          'sidebar-nav-link-integrations',
          'sidebar-nav-link-chats',
          'sidebar-nav-link-users',
          'sidebar-nav-link-reports',
        ])
    })

    visibleByTestId('shell-workspace-menu-trigger').click()
    visibleByTestId('shell-workspace-menu-settings').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/display\/?$/
    )
    cy.go('back')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    openCustomiseNav()
    visibleByTestId('sidebar-nav-edit-panel').within(() => {
      cy.get('[data-testid="sidebar-nav-move-up-wishlist"]').click({
        force: true,
      })
      cy.get('[data-testid^="sidebar-nav-edit-item-"]')
        .should(($items) => {
          const ids = [...$items].map(
            (item) => item.getAttribute('data-testid') || ''
          )
          expect(ids).to.deep.equal([
            'sidebar-nav-edit-item-dashboard',
            'sidebar-nav-edit-item-inventory',
            'sidebar-nav-edit-item-media',
            'sidebar-nav-edit-item-wishlist',
            'sidebar-nav-edit-item-collections',
            'sidebar-nav-edit-item-discoveries',
            'sidebar-nav-edit-item-market-watch',
            'sidebar-nav-edit-item-inbox',
            'sidebar-nav-edit-item-purchases',
            'sidebar-nav-edit-item-integrations',
            'sidebar-nav-edit-item-chats',
            'sidebar-nav-edit-item-users',
            'sidebar-nav-edit-item-reports',
          ])
        })
      cy.get('[data-testid="sidebar-nav-visibility-integrations"]').click()
    })
    visibleByTestId('sidebar-nav-cancel').click()

    visibleByTestId('sidebar-nav-group-general').within(() => {
      cy.get('[data-testid^="sidebar-nav-link-"]')
        .then(($links) =>
          [...$links].map((link) => link.getAttribute('data-testid') || '')
        )
        .should('deep.equal', [
          'sidebar-nav-link-dashboard',
          'sidebar-nav-link-inventory',
          'sidebar-nav-link-media',
          'sidebar-nav-link-collections',
          'sidebar-nav-link-wishlist',
          'sidebar-nav-link-discoveries',
          'sidebar-nav-link-market-watch',
          'sidebar-nav-link-inbox',
          'sidebar-nav-link-purchases',
          'sidebar-nav-link-integrations',
          'sidebar-nav-link-chats',
          'sidebar-nav-link-users',
          'sidebar-nav-link-reports',
        ])
    })

    openCustomiseNav()
    visibleByTestId('sidebar-nav-edit-panel').within(() => {
      cy.get('[data-testid="sidebar-nav-move-up-wishlist"]').click({
        force: true,
      })
      cy.get('[data-testid="sidebar-nav-visibility-integrations"]').click()
    })
    visibleByTestId('sidebar-nav-apply').click()

    visibleByTestId('sidebar-nav-group-general').within(() => {
      cy.get('[data-testid^="sidebar-nav-link-"]')
        .then(($links) =>
          [...$links].map((link) => link.getAttribute('data-testid') || '')
        )
        .should('deep.equal', [
          'sidebar-nav-link-dashboard',
          'sidebar-nav-link-inventory',
          'sidebar-nav-link-media',
          'sidebar-nav-link-wishlist',
          'sidebar-nav-link-collections',
          'sidebar-nav-link-discoveries',
          'sidebar-nav-link-market-watch',
          'sidebar-nav-link-inbox',
          'sidebar-nav-link-purchases',
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
          'sidebar-nav-link-media',
          'sidebar-nav-link-wishlist',
          'sidebar-nav-link-collections',
          'sidebar-nav-link-discoveries',
          'sidebar-nav-link-market-watch',
          'sidebar-nav-link-inbox',
          'sidebar-nav-link-purchases',
          'sidebar-nav-link-chats',
          'sidebar-nav-link-users',
          'sidebar-nav-link-reports',
        ])
    })

    openCustomiseNav()
    visibleByTestId('sidebar-nav-restore-hidden').click()
    visibleByTestId('sidebar-nav-apply').click()
    visibleByTestId('sidebar-nav-link-integrations').should('be.visible')

    openCustomiseNav()
    visibleByTestId('sidebar-nav-reset-defaults').click()
    visibleByTestId('sidebar-nav-apply').click()
    visibleByTestId('sidebar-nav-group-general')
      .find('[data-testid^="sidebar-nav-link-"]')
      .then(($links) =>
        [...$links].map((link) => link.getAttribute('data-testid') || '')
      )
      .should('deep.equal', [
        'sidebar-nav-link-dashboard',
        'sidebar-nav-link-inventory',
        'sidebar-nav-link-media',
        'sidebar-nav-link-collections',
        'sidebar-nav-link-wishlist',
        'sidebar-nav-link-discoveries',
        'sidebar-nav-link-market-watch',
        'sidebar-nav-link-inbox',
        'sidebar-nav-link-purchases',
        'sidebar-nav-link-integrations',
        'sidebar-nav-link-chats',
        'sidebar-nav-link-users',
        'sidebar-nav-link-reports',
      ])
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-007 reflects live nav edit order and saves the exact shown order', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    openCustomiseNav()

    visibleByTestId('sidebar-nav-edit-panel').within(() => {
      cy.get('[data-testid="sidebar-nav-move-up-wishlist"]').click({
        force: true,
      })
      cy.get('[data-testid^="sidebar-nav-edit-item-"]')
        .should(($items) => {
          const ids = [...$items].map(
            (item) => item.getAttribute('data-testid') || ''
          )
          expect(ids).to.deep.equal([
            'sidebar-nav-edit-item-dashboard',
            'sidebar-nav-edit-item-inventory',
            'sidebar-nav-edit-item-media',
            'sidebar-nav-edit-item-wishlist',
            'sidebar-nav-edit-item-collections',
            'sidebar-nav-edit-item-discoveries',
            'sidebar-nav-edit-item-market-watch',
            'sidebar-nav-edit-item-inbox',
            'sidebar-nav-edit-item-purchases',
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
        .should(($items) => {
          const ids = [...$items].map(
            (item) => item.getAttribute('data-testid') || ''
          )
          expect(ids).to.deep.equal([
            'sidebar-nav-edit-item-dashboard',
            'sidebar-nav-edit-item-inventory',
            'sidebar-nav-edit-item-wishlist',
            'sidebar-nav-edit-item-media',
            'sidebar-nav-edit-item-collections',
            'sidebar-nav-edit-item-discoveries',
            'sidebar-nav-edit-item-market-watch',
            'sidebar-nav-edit-item-inbox',
            'sidebar-nav-edit-item-purchases',
            'sidebar-nav-edit-item-integrations',
            'sidebar-nav-edit-item-chats',
            'sidebar-nav-edit-item-users',
            'sidebar-nav-edit-item-reports',
          ])
        })
    })

    visibleByTestId('sidebar-nav-apply').click()
    visibleByTestId('sidebar-nav-group-general')
      .find('[data-testid^="sidebar-nav-link-"]')
      .then(($links) => [...$links].map((link) => link.getAttribute('data-testid') || ''))
      .should('deep.equal', [
        'sidebar-nav-link-dashboard',
        'sidebar-nav-link-inventory',
        'sidebar-nav-link-wishlist',
        'sidebar-nav-link-media',
        'sidebar-nav-link-collections',
        'sidebar-nav-link-discoveries',
        'sidebar-nav-link-market-watch',
        'sidebar-nav-link-inbox',
        'sidebar-nav-link-purchases',
        'sidebar-nav-link-integrations',
        'sidebar-nav-link-chats',
        'sidebar-nav-link-users',
        'sidebar-nav-link-reports',
      ])
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-007 supports left-side drag-handle reorder with visible insertion feedback', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    const dataTransfer = new DataTransfer()

    openCustomiseNav()

    visibleByTestId('sidebar-nav-edit-item-wishlist').scrollIntoView()
    visibleByTestId('sidebar-nav-drag-handle-wishlist').should('be.visible')

    cy.get('[data-testid="sidebar-nav-edit-item-media"]').then(($target) => {
      const rect = $target[0].getBoundingClientRect()

      visibleByTestId('sidebar-nav-drag-handle-wishlist').trigger('dragstart', {
        dataTransfer,
      })
      cy.get('[data-testid="sidebar-nav-edit-dropzone-media"]').trigger('dragover', {
        dataTransfer,
        clientY: rect.top + 2,
      })
      visibleByTestId('sidebar-nav-drop-indicator-before-media').should('be.visible')
      cy.get('[data-testid="sidebar-nav-edit-dropzone-media"]').trigger('drop', {
        dataTransfer,
        clientY: rect.top + 2,
      })
    })

    visibleByTestId('sidebar-nav-edit-panel').within(() => {
      cy.get('[data-testid^="sidebar-nav-edit-item-"]')
        .then(($items) => [...$items].map((item) => item.getAttribute('data-testid') || ''))
        .should('deep.equal', [
          'sidebar-nav-edit-item-dashboard',
          'sidebar-nav-edit-item-inventory',
          'sidebar-nav-edit-item-wishlist',
          'sidebar-nav-edit-item-media',
          'sidebar-nav-edit-item-collections',
          'sidebar-nav-edit-item-discoveries',
          'sidebar-nav-edit-item-market-watch',
          'sidebar-nav-edit-item-inbox',
          'sidebar-nav-edit-item-purchases',
          'sidebar-nav-edit-item-integrations',
          'sidebar-nav-edit-item-chats',
          'sidebar-nav-edit-item-users',
          'sidebar-nav-edit-item-reports',
        ])

      cy.get('[data-testid="sidebar-nav-edit-item-wishlist"]').scrollIntoView({
        block: 'center',
      })
      cy.get('[data-testid="sidebar-nav-move-down-wishlist"]').should('be.visible')
      cy.get('[data-testid="sidebar-nav-visibility-wishlist"]').should('be.visible')
    })

    visibleByTestId('sidebar-nav-apply').click()
    visibleByTestId('sidebar-nav-group-general')
      .find('[data-testid^="sidebar-nav-link-"]')
      .then(($links) => [...$links].map((link) => link.getAttribute('data-testid') || ''))
      .should('deep.equal', [
        'sidebar-nav-link-dashboard',
        'sidebar-nav-link-inventory',
        'sidebar-nav-link-wishlist',
        'sidebar-nav-link-media',
        'sidebar-nav-link-collections',
        'sidebar-nav-link-discoveries',
        'sidebar-nav-link-market-watch',
        'sidebar-nav-link-inbox',
        'sidebar-nav-link-purchases',
        'sidebar-nav-link-integrations',
        'sidebar-nav-link-chats',
        'sidebar-nav-link-users',
        'sidebar-nav-link-reports',
      ])
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-003 updates collection browser context when folder selection changes', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    cy.get('[data-testid="collection-active-context"]').should(
      'have.text',
      'All Items'
    )

    cy.get('[data-testid="collection-folder-store-1"]').click()
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
    visibleByTestId('collections-new-action').click()
    visibleByTestId('collections-create-input').clear().type('Quick Create Shelf')
    visibleByTestId('collections-create-submit').click()
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

  it('UI-FOUNDATION-SHELL-NAVIGATION-008 uses explicit Database terminology in top switcher', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    visibleByTestId('team-switcher-trigger').click()
    cy.get('[data-slot="dropdown-menu-content"]').should('be.visible')
    cy.contains('[data-slot="dropdown-menu-label"]', /^Database$/).should('be.visible')
    cy.contains('[data-slot="dropdown-menu-item"]', /add database/i).should('be.visible')
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-009 switches active DB profile and reloads active data context', () => {
    cy.request('POST', '/api/test/reset', {})
    cy.request('POST', '/api/profiles', { name: 'Primary DB' }).then((primaryResp) => {
      expect(primaryResp.status).to.eq(201)
      const primaryID = primaryResp.body.id as string

      cy.request('POST', '/api/profiles', { name: 'Showcase DB' }).then((showcaseResp) => {
        expect(showcaseResp.status).to.eq(201)
        const showcaseID = showcaseResp.body.id as string

        cy.request('PUT', '/api/profiles/active', { profile_id: primaryID }).its('status').should('eq', 200)
        cy.request('POST', '/api/items', {
          part_number: 'PRI-001',
          title: 'Primary Item',
          brand: 'AFX',
          category: 'Cars',
        }).its('status').should('eq', 201)

        cy.request('PUT', '/api/profiles/active', { profile_id: showcaseID }).its('status').should('eq', 200)
        cy.request('POST', '/api/items', {
          part_number: 'SHW-001',
          title: 'Showcase Item',
          brand: 'AFX',
          category: 'Cars',
        }).its('status').should('eq', 201)

        cy.request('PUT', '/api/profiles/active', { profile_id: primaryID }).its('status').should('eq', 200)
      })
    })

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.contains('Primary Item').should('be.visible')
    cy.contains('Showcase Item').should('not.exist')

    visibleByTestId('team-switcher-trigger').click()
    cy.get('[data-testid="team-option-showcase-db"]').click()
    cy.contains('Showcase Item', { timeout: 20000 }).should('be.visible')
    cy.contains('Primary Item').should('not.exist')
    visibleByTestId('active-profile-name').should('contain', 'Showcase DB')
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-010 provides Showcase DB profile with seeded demo context', () => {
    cy.request('POST', '/api/test/reset', {})
    cy.request('POST', '/api/profiles', { name: 'Primary DB' }).then((primaryResp) => {
      const primaryID = primaryResp.body.id as string
      cy.request('POST', '/api/profiles', { name: 'Showcase DB' }).then((showcaseResp) => {
        const showcaseID = showcaseResp.body.id as string
        cy.request('PUT', '/api/profiles/active', { profile_id: showcaseID }).its('status').should('eq', 200)
        cy.request('POST', '/api/items', {
          part_number: 'SHW-100',
          title: 'Showcase Seed One',
          brand: 'AFX',
          category: 'Cars',
        }).its('status').should('eq', 201)
        cy.request('POST', '/api/items', {
          part_number: 'SHW-101',
          title: 'Showcase Seed Two',
          brand: 'Tyco',
          category: 'Cars',
        }).its('status').should('eq', 201)
        cy.request('PUT', '/api/profiles/active', { profile_id: primaryID }).its('status').should('eq', 200)
      })
    })

    signInTo('/inventory/')
    visibleByTestId('team-switcher-trigger').click()
    cy.get('[data-testid="team-option-primary-db-plan"]').should('contain', 'Database')
    cy.get('[data-testid="team-option-showcase-db-plan"]').should(
      'contain',
      'Showcase sample data'
    )
    cy.get('[data-testid="team-option-showcase-db"]').click()
    cy.contains('Showcase Seed One', { timeout: 20000 }).should('be.visible')
    cy.contains('Showcase Seed Two').should('be.visible')
    visibleByTestId('active-profile-name').should('contain', 'Showcase DB')
    cy.get('[data-testid="active-profile-status"]').should(
      'contain',
      'Showcase sample data'
    )
    cy.get('[data-testid="active-profile-db-icon"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Showcase DB database profile')
    cy.get('[data-testid="active-profile-db-icon-variant"]').should(
      'have.attr',
      'data-db-icon-variant',
      'dark'
    )
    visibleByTestId('team-switcher-trigger').click()
    cy.get('[data-testid="team-option-showcase-db-icon"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Showcase DB database profile')
    cy.get('[data-testid="team-option-showcase-db-icon-light"]').should(
      'have.attr',
      'data-db-icon-variant',
      'light'
    )
    cy.get('[data-testid="team-option-showcase-db-icon-dark"]').should(
      'have.attr',
      'data-db-icon-variant',
      'dark'
    )
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-011 fills available shell width on wide desktop viewport by default', () => {
    cy.viewport(1800, 1000)
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    cy.get('[data-slot="sidebar-inset"]').then(($inset) => {
      const insetRect = $inset[0].getBoundingClientRect()

      visibleByTestId('app-main-content').then(($main) => {
        const mainRect = $main[0].getBoundingClientRect()
        const availableWidth = insetRect.width
        const widthGap = availableWidth - mainRect.width

        expect(mainRect.width, 'main content width').to.be.greaterThan(1450)
        expect(widthGap, 'remaining unused inset width').to.be.lessThan(80)
      })
    })

    cy.get('[data-testid="inventory-header-title"]').should('be.visible')
    cy.contains('Folders').should('be.visible')
    cy.contains('Items').should('be.visible')
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-012 right-aligns sidebar notification badges as trailing affordances', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    cy.get('[data-testid="sidebar-nav-link-chats"]').should('be.visible')
    cy.get('[data-testid="sidebar-nav-link-chats"]').should(
      'have.attr',
      'aria-label',
      'Chats'
    )
    cy.get('[data-testid="sidebar-nav-label-chats"]').should('not.exist')
    cy.get('[data-testid="sidebar-nav-badge-chats"]').should('contain', '3')

    cy.get('[data-testid="sidebar-nav-link-chats"]').then(($link) => {
      const linkRect = $link[0].getBoundingClientRect()

      cy.get('[data-testid="sidebar-nav-badge-chats"]').then(($badge) => {
        const badgeRect = $badge[0].getBoundingClientRect()

        expect(linkRect.right - badgeRect.right, 'badge hugs row end').to.be.lessThan(24)
        expect(badgeRect.left, 'badge stays in trailing affordance area').to.be.greaterThan(
          linkRect.left + linkRect.width / 2
        )
      })
    })
  })

  it('UI-FOUNDATION-SHELL-NAVIGATION-018 renders primary navigation as icon-only accessible controls', () => {
    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

    const iconOnlyLinks = [
      ['dashboard', 'Dashboard'],
      ['inventory', 'Inventory'],
      ['media', 'Media'],
      ['collections', 'Collections'],
      ['wishlist', 'Wishlist'],
      ['discoveries', 'Discoveries'],
      ['market-watch', 'Market Watch'],
      ['purchases', 'Purchases'],
      ['integrations', 'Integrations'],
      ['chats', 'Chats'],
      ['users', 'Users'],
      ['reports', 'Reports'],
    ] as const

    iconOnlyLinks.forEach(([key, label]) => {
      cy.get(`[data-testid="sidebar-nav-link-${key}"]`)
        .scrollIntoView()
        .should('be.visible')
        .and('have.attr', 'aria-label', label)
        .within(() => {
          cy.get('svg').should('be.visible')
          cy.get(`[data-testid="sidebar-nav-label-${key}"]`).should('not.exist')
        })
    })

    cy.get('[data-testid="sidebar-nav-link-inventory"]')
      .should('have.attr', 'data-active', 'true')
      .focus()
      .should('be.focused')
  })
})
