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

    cy.get('[data-testid="settings-storage-retry"]').click()
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

  it('UI-SCREEN-SETTINGS-STORAGE-006 reports Reindex Search failure without route reload', () => {
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
      statusCode: 500,
      body: { error: 'failed_to_reindex' },
    }).as('reindexSearch')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('button', 'Reindex Search').click()
    cy.wait('@reindexSearch')
    cy.location('pathname').should('match', /^\/settings\/storage\/?$/)
    cy.get('[data-testid="settings-storage-action-status"]').should(
      'contain',
      'Search reindex failed. Try again when runtime diagnostics are healthy.'
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

  it('UI-SCREEN-SETTINGS-STORAGE-006 reports Rebuild Thumbnails failure without route reload', () => {
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
      statusCode: 500,
      body: { error: 'failed_to_rebuild_thumbnails' },
    }).as('rebuildThumbnails')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('button', 'Rebuild Thumbnails').click()
    cy.wait('@rebuildThumbnails')
    cy.location('pathname').should('match', /^\/settings\/storage\/?$/)
    cy.get('[data-testid="settings-storage-action-status"]').should(
      'contain',
      'Thumbnail rebuild failed. Check diagnostics health and try again.'
    )
  })

  it('UI-SCREEN-SETTINGS-STORAGE-007 creates, lists, and restores backups from Settings Storage', () => {
    const backups = [
      {
        path: 'C:/cabinet/backups/cabinet-2026-04-21-120000.db',
        file_name: 'cabinet-2026-04-21-120000.db',
        size_bytes: 2048,
        created_at: '2026-04-21T12:00:00Z',
        archive_format: 'db',
        download_url: '/api/backup/download?file_name=cabinet-2026-04-21-120000.db',
        integrity_check: 'ok',
      },
    ]

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
    cy.intercept('GET', '/api/backup/list', (req) => {
      req.reply({
        statusCode: 200,
        body: { backups: [...backups] },
      })
    }).as('backupList')
    cy.intercept('POST', '/api/backup/run', (req) => {
      backups.unshift({
        path: 'C:/cabinet/backups/cabinet-backup-2026-04-21-130000.zip',
        file_name: 'cabinet-backup-2026-04-21-130000.zip',
        size_bytes: 4096,
        created_at: '2026-04-21T13:00:00Z',
        archive_format: 'zip',
        download_url: '/api/backup/download?file_name=cabinet-backup-2026-04-21-130000.zip',
        integrity_check: 'ok',
      })
      req.reply({
        statusCode: 200,
        body: { backup: backups[0] },
      })
    }).as('backupRun')
    cy.intercept('POST', '/api/backup/restore', (req) => {
      expect(req.body).to.deep.equal({
        backup_path: 'C:/cabinet/backups/cabinet-backup-2026-04-21-130000.zip',
        confirm_restore: true,
      })
      req.reply({
        statusCode: 200,
        body: {
          restore: {
            restored_path: 'C:/cabinet/backups/cabinet-backup-2026-04-21-130000.zip',
            restored_at: '2026-04-21T13:05:00Z',
            integrity_check: 'ok',
          },
        },
      })
    }).as('backupRestore')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.wait('@backupList')

    cy.get('[data-testid="settings-storage-backup-row"]').should('have.length', 1)
    cy.get('[data-testid="settings-storage-backup-run"]').click()
    cy.wait('@backupRun')
    cy.wait('@backupList')
    cy.get('[data-testid="settings-storage-backup-row"]').should('have.length', 2)
    cy.get('[data-testid="settings-storage-backup-row"]').first().should(
      'contain',
      'cabinet-backup-2026-04-21-130000.zip'
    )
    cy.get('[data-testid="settings-storage-backup-row"]')
      .first()
      .should('contain', 'ZIP archive')
      .find('[data-testid="settings-storage-backup-download"]')
      .should('have.attr', 'href', '/api/backup/download?file_name=cabinet-backup-2026-04-21-130000.zip')
      .and('have.attr', 'download', 'cabinet-backup-2026-04-21-130000.zip')

    cy.get('[data-testid="settings-storage-backup-row"]')
      .first()
      .find('[data-testid="settings-storage-backup-restore"]')
      .click()
    cy.get('[data-testid="settings-storage-restore-confirm"]').should('be.visible')
    cy.get('[data-testid="settings-storage-restore-submit"]').click()
    cy.wait('@backupRestore')
    cy.get('[data-testid="settings-storage-action-status"]').should(
      'contain',
      'Backup restored successfully.'
    )
  })

  it('UI-SCREEN-SETTINGS-STORAGE-008 reports restore failure without route reload', () => {
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
      body: {
        backups: [
          {
            path: 'C:/cabinet/backups/cabinet-2026-04-21-090000.db',
            file_name: 'cabinet-2026-04-21-090000.db',
            size_bytes: 1024,
            created_at: '2026-04-21T09:00:00Z',
            archive_format: 'db',
            download_url: '/api/backup/download?file_name=cabinet-2026-04-21-090000.db',
            integrity_check: 'ok',
          },
        ],
      },
    }).as('backupList')
    cy.intercept('POST', '/api/backup/restore', {
      statusCode: 400,
      body: { error: 'failed_to_restore_backup' },
    }).as('backupRestore')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.wait('@backupList')

    cy.get('[data-testid="settings-storage-backup-row"]')
      .first()
      .find('[data-testid="settings-storage-backup-restore"]')
      .click()
    cy.get('[data-testid="settings-storage-restore-submit"]').click()
    cy.wait('@backupRestore')
    cy.location('pathname').should('match', /^\/settings\/storage\/?$/)
    cy.get('[data-testid="settings-storage-action-status"]').should(
      'contain',
      'Backup restore failed.'
    )
    cy.get('[data-testid="settings-storage-restore-confirm"]').should('not.exist')
  })

  it('UI-SCREEN-SETTINGS-STORAGE-009 runs storage integrity check and shows healthy result', () => {
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
    cy.intercept('POST', '/api/data/repair', {
      statusCode: 200,
      body: { integrity_check: 'ok' },
    }).as('repairCheck')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.wait('@backupList')

    cy.get('[data-testid="settings-storage-repair-run"]').click()
    cy.wait('@repairCheck')
    cy.get('[data-testid="settings-storage-repair-result"]').should(
      'contain',
      'Database integrity check passed.'
    )
    cy.get('[data-testid="settings-storage-repair-result"]').should(
      'contain',
      'ok'
    )
  })

  it('UI-SCREEN-SETTINGS-STORAGE-010 reports integrity-check failure without route reload', () => {
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
    cy.intercept('POST', '/api/data/repair', {
      statusCode: 500,
      body: { error: 'failed_to_repair_check' },
    }).as('repairCheck')

    signInToStorage()
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.wait('@backupList')

    cy.get('[data-testid="settings-storage-repair-run"]').click()
    cy.wait('@repairCheck')
    cy.location('pathname').should('match', /^\/settings\/storage\/?$/)
    cy.get('[data-testid="settings-storage-repair-result"]').should(
      'contain',
      'Database integrity check failed.'
    )
  })

})
