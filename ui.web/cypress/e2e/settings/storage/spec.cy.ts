describe('settings/storage', () => {
  function signInToStorage() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Fstorage')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/storage\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-SETTINGS-STORAGE-004 renders storage paths and keeps diagnostics actions disabled in degraded mode', () => {
    let storageAttempt = 0
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'default' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/*/storage', (req) => {
      storageAttempt += 1
      if (storageAttempt === 1) {
        req.reply(200, {
          db_path: 'C:/cabinet/profiles/default/cabinet.db',
          media_dir: 'C:/cabinet/profiles/default/media',
        })
        return
      }
      req.reply(503, { error: 'storage_unavailable' })
    }).as('storageInfo')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('cabinet.db').should('be.visible')
    cy.contains('/default/media').should('be.visible')
    cy.contains('button', 'Reindex Search (Diagnostics only)').should('be.disabled')
    cy.contains('button', 'Rebuild Thumbnails (Diagnostics only)').should(
      'be.disabled'
    )

    cy.visit('/settings/storage')
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('Storage information is unavailable right now.').should(
      'be.visible'
    )
    cy.contains('cabinet.db').should('be.visible')
    cy.contains('/default/media').should('be.visible')
    cy.contains('Diagnostics actions are unavailable while storage info is degraded.')
      .should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-STORAGE-005 retries storage fetch and recovers without route reload', () => {
    let storageAttempt = 0
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'default' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/*/storage', (req) => {
      storageAttempt += 1
      if (storageAttempt === 1) {
        req.reply(503, { error: 'storage_unavailable' })
        return
      }
      req.reply(200, {
        db_path: 'C:/cabinet/profiles/default/recovered.db',
        media_dir: 'C:/cabinet/profiles/default/recovered-media',
      })
    }).as('storageInfo')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('Storage information is unavailable right now.').should(
      'be.visible'
    )

    cy.contains('button', 'Retry').click()
    cy.wait('@storageInfo')
    cy.location('pathname').should('match', /^\/settings\/storage\/?$/)
    cy.contains('Storage information is unavailable right now.').should(
      'not.exist'
    )
    cy.contains('recovered.db').should('be.visible')
    cy.contains('/recovered-media').should('be.visible')
  })
})
