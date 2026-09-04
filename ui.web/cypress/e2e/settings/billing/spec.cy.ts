describe('settings billing screen', () => {
  function openBilling() {
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap({ minimalProfile: true }).then(
      ({ profile_id, profile_name }) => {
        cy.useBootstrappedProfile(profile_id, profile_name, {
          path: '/settings/billing/',
        })
      }
    )
  }

  it('UI-SCREEN-SETTINGS-BILLING-001 renders disabled static billing state without portal mutation', () => {
    cy.intercept('GET', '/api/auth/cloud/session/effective', {
      provider: 'zitadel',
      user_id: 'user_billing',
      email: 'billing@example.com',
      role: 'member',
      plan: 'free',
      features: ['collection_core'],
    }).as('effectiveSession')
    cy.intercept('GET', '/api/license/status*', {
      state: 'free',
      tier: 'free',
      features: [],
      expires_at: '',
    }).as('licenseStatus')

    openBilling()
    cy.wait(['@effectiveSession', '@licenseStatus'])

    cy.contains('h3', 'Billing').should('be.visible')
    cy.get('aside a[href="/settings/billing"]').should(
      'have.attr',
      'aria-current',
      'page'
    )

    cy.get('[data-testid="billing-plan-card"]').within(() => {
      cy.contains('Current plan').should('be.visible')
      cy.contains('Free').should('be.visible')
      cy.contains('Source: Cloud entitlement state').should('be.visible')
      cy.contains('Free includes 250 inventory items and export access.').should(
        'be.visible'
      )
    })

    cy.get('[data-testid="billing-license-card"]').within(() => {
      cy.contains('Founding license').should('be.visible')
      cy.contains('No signed license imported').should('be.visible')
      cy.contains('Paste a signed founding license payload below.').should(
        'be.visible'
      )
    })

    cy.contains('button', 'Open Billing Portal (Coming soon)')
      .scrollIntoView()
      .should('be.visible')
      .and('be.disabled')

    cy.get('[data-testid="founding-license-import"]').should('be.visible')
    cy.contains('button', 'Import signed license').should('be.disabled')
    cy.contains('a', 'Open Billing Portal').should('not.exist')
    cy.location('pathname').should('match', /^\/settings\/billing\/?$/)
  })
})
