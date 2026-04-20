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
    cy.contains('button', 'Reindex Search').should('be.enabled')
    cy.contains('button', 'Rebuild Thumbnails').should('be.enabled')

    cy.visit('/settings/storage')
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('Storage information is unavailable right now.').should(
      'be.visible'
    )
    cy.contains('cabinet.db').should('be.visible')
    cy.contains('/default/media').should('be.visible')
    cy.contains('button', 'Reindex Search').should('be.disabled')
    cy.contains('button', 'Rebuild Thumbnails').should('be.disabled')
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

  it('UI-SCREEN-SETTINGS-STORAGE-006 runs Reindex Search and reports deterministic completion feedback', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'default' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/*/storage', {
      statusCode: 200,
      body: {
        db_path: 'C:/cabinet/profiles/default/cabinet.db',
        media_dir: 'C:/cabinet/profiles/default/media',
      },
    }).as('storageInfo')
    cy.intercept('POST', '/api/data/reindex', {
      statusCode: 200,
      body: { ok: true },
    }).as('reindexSearch')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('button', 'Reindex Search').click()
    cy.wait('@reindexSearch')
    cy.get('[data-testid="settings-storage-action-status"]').should(
      'contain',
      'Search reindex completed successfully.'
    )
  })

  it('UI-SCREEN-SETTINGS-STORAGE-006 runs Rebuild Thumbnails and reports deterministic feedback', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'default' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/*/storage', {
      statusCode: 200,
      body: {
        db_path: 'C:/cabinet/profiles/default/cabinet.db',
        media_dir: 'C:/cabinet/profiles/default/media',
      },
    }).as('storageInfo')
    cy.intercept('POST', '/api/data/rebuild-thumbnails', {
      statusCode: 200,
      body: { ok: true, rebuilt_items: 2, rebuilt_photos: 5 },
    }).as('rebuildThumbnails')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('button', 'Rebuild Thumbnails').click()
    cy.wait('@rebuildThumbnails')
    cy.get('[data-testid="settings-storage-action-status"]').should(
      'contain',
      'Thumbnail rebuild completed for 5 photos across 2 items.'
    )
  })
})
