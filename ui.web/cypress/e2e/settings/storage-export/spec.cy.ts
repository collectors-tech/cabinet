describe('settings/storage export downloads', () => {
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

  it('UI-SCREEN-SETTINGS-STORAGE-011 exposes JSON snapshot and item CSV download actions', () => {
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
    cy.intercept('GET', '/api/backup/list', {
      statusCode: 200,
      body: { backups: [] },
    }).as('backupList')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.wait('@backupList')

    cy.contains('Data exports').should('be.visible')
    cy.contains('Download the active profile as a JSON snapshot or item CSV.').should(
      'be.visible'
    )
    cy.get('[data-testid="settings-storage-export-json"]')
      .should('have.attr', 'href', '/api/data/export/json')
      .and('have.attr', 'download', 'cabinet-data-snapshot.json')
    cy.get('[data-testid="settings-storage-export-csv"]')
      .should('have.attr', 'href', '/api/data/export/csv/items')
      .and('have.attr', 'download', 'cabinet-items.csv')
  })

  it('UI-SCREEN-SETTINGS-STORAGE-011 disables export downloads while storage context is degraded', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'default' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/*/storage', {
      statusCode: 503,
      body: { error: 'storage_unavailable' },
    }).as('storageInfo')
    cy.intercept('GET', '/api/backup/list', {
      statusCode: 200,
      body: { backups: [] },
    }).as('backupList')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.wait('@backupList')

    cy.contains('Storage information is unavailable right now.').should(
      'be.visible'
    )
    cy.contains('Data exports').should('be.visible')
    cy.contains('button', 'JSON Snapshot').should('be.disabled')
    cy.contains('button', 'Item CSV').should('be.disabled')
    cy.get('[data-testid="settings-storage-export-json"]').should('not.exist')
    cy.get('[data-testid="settings-storage-export-csv"]').should('not.exist')
  })
})
