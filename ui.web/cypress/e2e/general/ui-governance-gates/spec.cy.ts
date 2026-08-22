describe('general/ui-governance-gates', () => {
  function bootstrapAndSignIn(path = '/dashboard') {
    cy.e2eReset()
    cy.e2eBootstrap().then((bootstrap) => {
      cy.setCookie('sidebar_state', 'true')
      cy.intercept('GET', '/api/dashboard', {
        statusCode: 200,
        body: {
          new_discoveries: 4,
          wishlist_hits: 2,
          price_drops: 1,
          low_stock_discoveries: 1,
          restocks: 0,
          recently_added: ['Card 001', 'Card 002'],
          total_items: 100,
          total_instances: 220,
          estimated_value: 12345.67,
          cards: [
            { title: 'Discoveries', value: 4, link: '/discoveries' },
            { title: 'Wishlist', value: 2, link: '/wishlist' },
          ],
        },
      }).as('dashboard')
      cy.useBootstrappedProfile(bootstrap.profile_id, bootstrap.profile_name, { path })
      if (path === '/dashboard') {
        cy.wait('@dashboard')
      }
    })
  }

  it('UI-GOVERNANCE-GATES-001 enforces dominant above-fold primary outcome hierarchy', () => {
    bootstrapAndSignIn()
    cy.contains('h1', 'Home').should('be.visible')
    cy.contains('What needs action now in your collection').should('be.visible')
    cy.contains('Actions needed').should('be.visible')
    cy.contains('h1', 'Home').then(($h1) => {
      const h1Top = $h1[0].getBoundingClientRect().top
      cy.contains('Actions needed').then(($panelTitle) => {
        const panelTop = $panelTitle[0].getBoundingClientRect().top
        expect(h1Top).to.be.lessThan(panelTop)
      })
    })
  })

  it('UI-GOVERNANCE-GATES-002 keeps primary action visible and task-oriented', () => {
    bootstrapAndSignIn()
    cy.contains('button', 'Refresh dashboard').should('be.visible')
  })

  it('UI-GOVERNANCE-GATES-003 keeps shell stable while body scrolls', () => {
    bootstrapAndSignIn()
    cy.get('header').first().as('header')
    cy.get('@header').should('have.class', 'sticky')
    cy.get('[data-slot="sidebar"]').should('be.visible')
    cy.get('@header').then(($header) => {
      const initialTop = $header[0].getBoundingClientRect().top
      cy.scrollTo('bottom')
      cy.get('@header').then(($after) => {
        const afterTop = $after[0].getBoundingClientRect().top
        expect(Math.abs(afterTop - initialTop)).to.be.lessThan(10)
      })
    })
  })

  it('UI-GOVERNANCE-GATES-004 keeps diagnostics controls out of primary dashboard action path', () => {
    bootstrapAndSignIn()
    cy.contains('button', 'Open Settings Diagnostics').should('not.exist')
    cy.contains('button', 'Refresh dashboard').should('be.visible')
    cy.visit('/help-center')
    cy.contains('The Help Center now surfaces the available article set').should('be.visible')
  })

  it('UI-GOVERNANCE-GATES-005 proves structure/action/state coverage via deterministic dashboard transitions', () => {
    cy.e2eReset()
    cy.e2eBootstrap().then((bootstrap) => {
      cy.setCookie('sidebar_state', 'true')
      let attempts = 0
      cy.intercept('GET', '/api/dashboard', (req) => {
        attempts += 1
        if (attempts === 1) {
          req.reply({ statusCode: 500, body: { error: 'dashboard_failed' }, delay: 500 })
          return
        }
        req.reply({
          statusCode: 200,
          delay: 500,
          body: {
            new_discoveries: 0,
            wishlist_hits: 0,
            price_drops: 0,
            low_stock_discoveries: 0,
            restocks: 0,
            recently_added: ['Recovered card'],
            total_items: 10,
            total_instances: 10,
            estimated_value: 100,
            cards: [{ title: 'Discoveries', value: 0, link: '/discoveries' }],
          },
        })
      }).as('dashboardState')

      cy.useBootstrappedProfile(bootstrap.profile_id, bootstrap.profile_name, { path: '/dashboard' })
      cy.wait('@dashboardState')
      cy.contains('Dashboard unavailable').should('be.visible')
      cy.contains('button', 'Retry').click()
      cy.wait('@dashboardState')
      cy.contains('Recently added').should('be.visible')
      cy.contains('Recovered card').should('be.visible')
    })
  })

  it('UI-GOVERNANCE-GATES-006 attaches governance evidence through runtime metadata and support copy', () => {
    bootstrapAndSignIn()
    cy.get('[data-testid="sidebar-runtime-meta"]').should('be.visible')
    cy.get('[data-testid="sidebar-app-version"]').should('not.have.text', '')
    cy.get('[data-testid="sidebar-build-date"]').should('not.have.text', '')
    cy.visit('/help-center')
    cy.contains('The Help Center now surfaces the available article set').should('be.visible')
  })
})
