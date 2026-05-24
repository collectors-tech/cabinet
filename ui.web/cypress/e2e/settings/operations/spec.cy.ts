describe('settings/operations', () => {
  function signInToOperations() {
    cy.visit('/sign-in?redirect=%2Fsettings%2Foperations')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/settings\/operations\/?$/
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-001 renders runtime metadata and recovery visibility', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-test123',
        build_date: '2026-04-22',
        bind_mode: 'lan',
        runtime_host: '192.168.1.53',
        runtime_port: 17882,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: true,
      },
    }).as('runtimeRecovery')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'rev-test123'
    )
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      '192.168.1.53:17882'
    )
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'lan'
    )
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'stable'
    )
    cy.get('[data-testid="settings-operations-recovery-card"]').should(
      'contain',
      'Recovery required'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-002 retries runtime visibility without route reload', () => {
    let runtimeAttempt = 0
    let recoveryAttempt = 0

    cy.intercept('GET', '/api/runtime', (req) => {
      runtimeAttempt += 1
      if (runtimeAttempt === 1) {
        req.reply(503, { error: 'runtime_unavailable' })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          app_version: 'rev-recovered',
          build_date: '2026-04-22',
          bind_mode: 'loopback',
          runtime_host: '127.0.0.1',
          runtime_port: 17880,
          update_channel: 'stable',
          update_public_key_configured: false,
        },
      })
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', (req) => {
      recoveryAttempt += 1
      if (recoveryAttempt === 1) {
        req.reply(503, { error: 'runtime_recovery_unavailable' })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          recovery_required: false,
        },
      })
    }).as('runtimeRecovery')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')
    cy.get('[data-testid="settings-operations-runtime-error"]').should(
      'contain',
      'Runtime information is unavailable right now.'
    )

    cy.get('[data-testid="settings-operations-retry"]').click()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-runtime-error"]').should('not.exist')
    cy.get('[data-testid="settings-operations-runtime-card"]').should(
      'contain',
      'rev-recovered'
    )
    cy.get('[data-testid="settings-operations-recovery-card"]').should(
      'contain',
      'No recovery required'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-002A pauses and resumes worker scheduling without route reload', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-queue-controls',
        build_date: '2026-04-23',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: {
        id: 'profile-ops-queue',
      },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/profile-ops-queue/settings', {
      statusCode: 200,
      body: {
        settings: {
          scanner_schedule: '0 */6 * * *',
        },
      },
    }).as('profileSettings')
    cy.intercept('PUT', '/api/profiles/profile-ops-queue/settings', (req) => {
      const settings = req.body?.settings ?? {}
      if (settings.scanner_schedule === 'manual') {
        expect(settings.operations_queue_resume_schedule).to.equal('0 */6 * * *')
        req.reply({
          statusCode: 200,
          body: {
            settings: {
              scanner_schedule: 'manual',
              operations_queue_resume_schedule: '0 */6 * * *',
            },
          },
        })
        return
      }

      expect(settings.scanner_schedule).to.equal('0 */6 * * *')
      expect(settings.operations_queue_resume_schedule).to.equal('0 */6 * * *')
      req.reply({
        statusCode: 200,
        body: {
          settings: {
            scanner_schedule: '0 */6 * * *',
            operations_queue_resume_schedule: '0 */6 * * *',
          },
        },
      })
    }).as('saveProfileSettings')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')
    cy.wait('@activeProfile')
    cy.wait('@profileSettings')

    cy.get('[data-testid="settings-operations-queue-card"]').should(
      'contain',
      'Workers scheduled: 0 */6 * * *'
    )
    cy.get('[data-testid="settings-operations-queue-resume"]').should('be.disabled')

    cy.get('[data-testid="settings-operations-queue-pause"]').click()
    cy.wait('@saveProfileSettings')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-queue-status"]').should(
      'contain',
      'Workers paused.'
    )
    cy.get('[data-testid="settings-operations-queue-resume"]').should('not.be.disabled')

    cy.get('[data-testid="settings-operations-queue-resume"]').click()
    cy.wait('@saveProfileSettings')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-queue-status"]').should(
      'contain',
      'Workers scheduled: 0 */6 * * *'
    )
    cy.get('[data-testid="settings-operations-queue-pause"]').should('not.be.disabled')
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-002B sets recovery passphrase and begins reset without route reload', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-auth-recovery',
        build_date: '2026-04-23',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: true,
      },
    }).as('runtimeRecovery')
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: {
        id: 'profile-ops-recovery',
      },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/profile-ops-recovery/settings', {
      statusCode: 200,
      body: {
        settings: {
          scanner_schedule: 'manual',
        },
      },
    }).as('profileSettings')
    cy.intercept('POST', '/api/auth/recovery/passphrase', (req) => {
      expect(req.body).to.deep.equal({
        profile_id: 'profile-ops-recovery',
        passphrase: 'reset-hunter2',
      })
      req.reply({
        statusCode: 200,
        body: {
          ok: true,
        },
      })
    }).as('setRecoveryPassphrase')
    cy.intercept('POST', '/api/auth/recovery/reset/begin', (req) => {
      expect(req.body).to.deep.equal({
        profile_id: 'profile-ops-recovery',
        passphrase: 'reset-hunter2',
      })
      req.reply({
        statusCode: 200,
        body: {
          session_id: 'recovery-session-1',
        },
      })
    }).as('beginRecoveryReset')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')
    cy.wait('@activeProfile')
    cy.wait('@profileSettings')

    cy.get('[data-testid="settings-operations-auth-recovery-card"]').should(
      'contain',
      'Recovery Access'
    )
    cy.get('[data-testid="settings-operations-recovery-passphrase-input"]')
      .clear()
      .type('reset-hunter2')
    cy.get('[data-testid="settings-operations-recovery-passphrase-submit"]').click()
    cy.wait('@setRecoveryPassphrase')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-auth-recovery-status"]').should(
      'contain',
      'Recovery passphrase saved.'
    )

    cy.get('[data-testid="settings-operations-recovery-reset-submit"]').click()
    cy.wait('@beginRecoveryReset')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-auth-recovery-status"]').should(
      'contain',
      'Recovery reset session started.'
    )
    cy.get('[data-testid="settings-operations-auth-recovery-summary"]').should(
      'contain',
      'recovery-session-1'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-003 exports snapshot data and shows dry-run conflicts', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-export',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('GET', '/api/data/export/json', {
      statusCode: 200,
      body: {
        schema_version: 1,
        exported_at: '2026-04-22T12:00:00Z',
        items: [
          {
            brand: 'AFX',
            category: 'Slot',
            part_number: 'OPS-001',
            title: 'Exported Item',
            instances: [],
          },
        ],
      },
    }).as('exportJson')
    cy.intercept('POST', '/api/data/import/json/dry-run', {
      statusCode: 200,
      body: {
        total_items: 2,
        new_items: 1,
        conflicts: 1,
        conflict_details: [
          {
            part_number: 'OPS-001',
            existing_id: 'item-1',
          },
        ],
      },
    }).as('importDryRun')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-export-json"]').click()
    cy.wait('@exportJson')
    cy.get('[data-testid="settings-operations-data-status"]').should(
      'contain',
      'Exported 1 item snapshot.'
    )

    cy.get('[data-testid="settings-operations-import-json-input"]')
      .clear()
      .type(
        '{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"OPS-001","title":"Conflict"},{"brand":"AFX","category":"Slot","part_number":"OPS-NEW","title":"New Item"}]}}',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-json-dry-run"]').click()
    cy.wait('@importDryRun')
    cy.get('[data-testid="settings-operations-import-summary"]').should(
      'contain',
      '2 items'
    )
    cy.get('[data-testid="settings-operations-import-summary"]').should(
      'contain',
      '1 conflict'
    )
    cy.get('[data-testid="settings-operations-import-summary"]').should(
      'contain',
      'OPS-001'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-004 reports dry-run failure without route reload', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-import-error',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/json/dry-run', {
      statusCode: 400,
      body: {
        error: 'failed_to_dry_run_import',
      },
    }).as('importDryRun')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-json-input"]')
      .clear()
      .type(
        '{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"OPS-ERR","title":"Broken"}]}}',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-json-dry-run"]').click()
    cy.wait('@importDryRun')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-data-status"]').should(
      'contain',
      'Import dry-run failed.'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-005 applies a reviewed JSON import with the selected conflict action', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-import-apply',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/json/dry-run', {
      statusCode: 200,
      body: {
        total_items: 2,
        new_items: 1,
        conflicts: 1,
        conflict_details: [
          {
            part_number: 'OPS-APPLY-001',
            existing_id: 'item-apply-1',
          },
        ],
      },
    }).as('importDryRun')
    cy.intercept('POST', '/api/data/import/json/apply', (req) => {
      expect(req.body).to.deep.equal({
        snapshot: {
          schema_version: 1,
          items: [
            {
              brand: 'AFX',
              category: 'Slot',
              part_number: 'OPS-APPLY-001',
              title: 'Conflict',
            },
            {
              brand: 'AFX',
              category: 'Slot',
              part_number: 'OPS-APPLY-NEW',
              title: 'New Item',
            },
          ],
        },
        options: {
          default_action: 'create',
        },
      })
      req.reply({
        statusCode: 200,
        body: {
          ok: true,
          total_items: 2,
          created: 1,
          merged: 1,
          skipped: 0,
          failed: 0,
        },
      })
    }).as('importApply')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-json-input"]')
      .clear()
      .type(
        '{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"OPS-APPLY-001","title":"Conflict"},{"brand":"AFX","category":"Slot","part_number":"OPS-APPLY-NEW","title":"New Item"}]}}',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-json-dry-run"]').click()
    cy.wait('@importDryRun')
    cy.get('[data-testid="settings-operations-import-default-action"]').click()
    cy.get('[role="option"]').contains('Create new item').click()
    cy.get('[data-testid="settings-operations-import-json-apply"]').click()
    cy.wait('@importApply')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-data-status"]').should(
      'contain',
      'Import applied: 2 items, 1 created, 1 merged, 0 skipped, 0 failed.'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-006 reports import apply failure without route reload', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-import-apply-error',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/json/dry-run', {
      statusCode: 200,
      body: {
        total_items: 1,
        new_items: 0,
        conflicts: 1,
        conflict_details: [
          {
            part_number: 'OPS-APPLY-ERR',
            existing_id: 'item-apply-err',
          },
        ],
      },
    }).as('importDryRun')
    cy.intercept('POST', '/api/data/import/json/apply', {
      statusCode: 400,
      body: {
        error: 'failed_to_apply_import',
      },
    }).as('importApply')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-json-input"]')
      .clear()
      .type(
        '{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"OPS-APPLY-ERR","title":"Broken"}]}}',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-json-dry-run"]').click()
    cy.wait('@importDryRun')
    cy.get('[data-testid="settings-operations-import-json-apply"]').click()
    cy.wait('@importApply')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-data-status"]').should(
      'contain',
      'Import apply failed.'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-007 exports CSV and reports CSV dry-run summary', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-csv',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('GET', '/api/data/export/csv/items', {
      statusCode: 200,
      body: 'brand,category,part_number,title,make,model,year,scale,series,description\nAFX,Slot,CSV-001,CSV Item,,,,,,\n',
      headers: {
        'content-type': 'text/csv',
      },
    }).as('exportCsv')
    cy.intercept('POST', '/api/data/import/csv/dry-run', (req) => {
      expect(req.body).to.deep.equal({
        csv: 'brand,category,part_number,title\nAFX,Slot,CSV-001,Existing Item\nAFX,Slot,CSV-NEW,New Item',
        mapping: {},
      })
      req.reply({
        statusCode: 200,
        body: {
          total_items: 2,
          new_items: 1,
          conflicts: 1,
          conflict_details: [
            {
              part_number: 'CSV-001',
              existing_id: 'item-csv-1',
            },
          ],
        },
      })
    }).as('importCsvDryRun')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-export-csv"]').click()
    cy.wait('@exportCsv')
    cy.get('[data-testid="settings-operations-csv-status"]').should(
      'contain',
      'Exported CSV with 1 item row.'
    )

    cy.get('[data-testid="settings-operations-import-csv-input"]')
      .clear()
      .type(
        'brand,category,part_number,title\nAFX,Slot,CSV-001,Existing Item\nAFX,Slot,CSV-NEW,New Item',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-csv-dry-run"]').click()
    cy.wait('@importCsvDryRun')
    cy.get('[data-testid="settings-operations-csv-summary"]').should(
      'contain',
      '2 items'
    )
    cy.get('[data-testid="settings-operations-csv-summary"]').should(
      'contain',
      '1 conflict'
    )
    cy.get('[data-testid="settings-operations-csv-summary"]').should(
      'contain',
      'CSV-001'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-008 reports CSV dry-run failure without route reload', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-csv-error',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/csv/dry-run', {
      statusCode: 400,
      body: {
        error: 'failed_to_parse_csv',
      },
    }).as('importCsvDryRun')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-csv-input"]')
      .clear()
      .type('broken csv body', { parseSpecialCharSequences: false })
    cy.get('[data-testid="settings-operations-import-csv-dry-run"]').click()
    cy.wait('@importCsvDryRun')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-csv-status"]').should(
      'contain',
      'CSV dry-run failed.'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-009 applies a reviewed CSV import with the selected conflict action', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-csv-apply',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/csv/dry-run', {
      statusCode: 200,
      body: {
        total_items: 2,
        new_items: 1,
        conflicts: 1,
        conflict_details: [
          {
            part_number: 'CSV-APPLY-001',
            existing_id: 'item-csv-apply-1',
          },
        ],
      },
    }).as('importCsvDryRun')
    cy.intercept('POST', '/api/data/import/csv/apply', (req) => {
      expect(req.body).to.deep.equal({
        csv_import: {
          csv: 'brand,category,part_number,title\nAFX,Slot,CSV-APPLY-001,Conflict\nAFX,Slot,CSV-APPLY-NEW,New Item',
          mapping: {},
        },
        options: {
          default_action: 'create',
        },
      })
      req.reply({
        statusCode: 200,
        body: {
          ok: true,
          total_items: 2,
          created: 1,
          merged: 1,
          skipped: 0,
          failed: 0,
        },
      })
    }).as('importCsvApply')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-csv-input"]')
      .clear()
      .type(
        'brand,category,part_number,title\nAFX,Slot,CSV-APPLY-001,Conflict\nAFX,Slot,CSV-APPLY-NEW,New Item',
        { parseSpecialCharSequences: false }
      )
    cy.get('[data-testid="settings-operations-import-csv-dry-run"]').click()
    cy.wait('@importCsvDryRun')
    cy.get('[data-testid="settings-operations-import-csv-default-action"]').click()
    cy.get('[role="option"]').contains('Create new item').click()
    cy.get('[data-testid="settings-operations-import-csv-apply"]').click()
    cy.wait('@importCsvApply')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-csv-status"]').should(
      'contain',
      'CSV import applied: 2 items, 1 created, 1 merged, 0 skipped, 0 failed.'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-010 reports CSV import apply failure without route reload', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-csv-apply-error',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/csv/dry-run', {
      statusCode: 200,
      body: {
        total_items: 1,
        new_items: 0,
        conflicts: 1,
        conflict_details: [
          {
            part_number: 'CSV-APPLY-ERR',
            existing_id: 'item-csv-apply-err',
          },
        ],
      },
    }).as('importCsvDryRun')
    cy.intercept('POST', '/api/data/import/csv/apply', {
      statusCode: 400,
      body: {
        error: 'failed_to_apply_import',
      },
    }).as('importCsvApply')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-csv-input"]')
      .clear()
      .type('brand,category,part_number,title\nAFX,Slot,CSV-APPLY-ERR,Broken', {
        parseSpecialCharSequences: false,
      })
    cy.get('[data-testid="settings-operations-import-csv-dry-run"]').click()
    cy.wait('@importCsvDryRun')
    cy.get('[data-testid="settings-operations-import-csv-apply"]').click()
    cy.wait('@importCsvApply')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-csv-status"]').should(
      'contain',
      'CSV import apply failed.'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-011 runs CSV dry-run with custom column mapping', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-csv-mapping',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/csv/dry-run', (req) => {
      expect(req.body).to.deep.equal({
        csv: 'maker,kind,pn,name\nAFX,Slot,MAP-001,Mapped Item',
        mapping: {
          brand: 'maker',
          category: 'kind',
          part_number: 'pn',
          title: 'name',
        },
      })
      req.reply({
        statusCode: 200,
        body: {
          total_items: 1,
          new_items: 1,
          conflicts: 0,
          conflict_details: [],
        },
      })
    }).as('importCsvDryRun')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-csv-input"]')
      .clear()
      .type('maker,kind,pn,name\nAFX,Slot,MAP-001,Mapped Item', {
        parseSpecialCharSequences: false,
      })
    cy.get('[data-testid="settings-operations-import-csv-mapping-brand"]')
      .clear()
      .type('maker')
    cy.get('[data-testid="settings-operations-import-csv-mapping-category"]')
      .clear()
      .type('kind')
    cy.get('[data-testid="settings-operations-import-csv-mapping-part-number"]')
      .clear()
      .type('pn')
    cy.get('[data-testid="settings-operations-import-csv-mapping-title"]')
      .clear()
      .type('name')
    cy.get('[data-testid="settings-operations-import-csv-dry-run"]').click()
    cy.wait('@importCsvDryRun')
    cy.get('[data-testid="settings-operations-csv-summary"]').should(
      'contain',
      '1 items'
    )
    cy.get('[data-testid="settings-operations-csv-summary"]').should(
      'contain',
      '0 conflicts'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-012 applies CSV import with the selected custom mapping', () => {
    const mappedCsv = 'maker,kind,pn,name\nAFX,Slot,MAP-APPLY-001,Mapped Apply Item'

    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-csv-mapping-apply',
        build_date: '2026-04-22',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/data/import/csv/dry-run', {
      statusCode: 200,
      body: {
        total_items: 1,
        new_items: 1,
        conflicts: 0,
        conflict_details: [],
      },
    }).as('importCsvDryRun')
    cy.intercept('POST', '/api/data/import/csv/apply', (req) => {
      expect(req.body).to.deep.equal({
        csv_import: {
          csv: mappedCsv,
          mapping: {
            brand: 'maker',
            category: 'kind',
            part_number: 'pn',
            title: 'name',
          },
        },
        options: {
          default_action: 'merge',
        },
      })
      req.reply({
        statusCode: 200,
        body: {
          ok: true,
          total_items: 1,
          created: 1,
          merged: 0,
          skipped: 0,
          failed: 0,
        },
      })
    }).as('importCsvApply')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-import-csv-input"]')
      .clear()
      .type(mappedCsv, {
        parseSpecialCharSequences: false,
      })
    cy.get('[data-testid="settings-operations-import-csv-mapping-brand"]')
      .clear()
      .type('maker')
    cy.get('[data-testid="settings-operations-import-csv-mapping-category"]')
      .clear()
      .type('kind')
    cy.get('[data-testid="settings-operations-import-csv-mapping-part-number"]')
      .clear()
      .type('pn')
    cy.get('[data-testid="settings-operations-import-csv-mapping-title"]')
      .clear()
      .type('name')
    cy.get('[data-testid="settings-operations-import-csv-dry-run"]').click()
    cy.wait('@importCsvDryRun')
    cy.get('[data-testid="settings-operations-import-csv-apply"]').click()
    cy.wait('@importCsvApply')
    cy.get('[data-testid="settings-operations-csv-status"]').should(
      'contain',
      'CSV import applied: 1 item, 1 created, 0 merged, 0 skipped, 0 failed.'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-013 exports diagnostic logs without leaving the Operations route', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-logs-export',
        build_date: '2026-04-23',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('GET', '/api/logs/export', {
      statusCode: 200,
      body: '{"entries":[{"level":"info","message":"[REDACTED]"}]}',
      headers: {
        'content-type': 'application/json',
      },
    }).as('exportLogs')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-export-logs"]').click()
    cy.wait('@exportLogs')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-logs-status"]').should(
      'contain',
      'Exported runtime logs successfully.'
    )
    cy.get('[data-testid="settings-operations-logs-preview"]').should(
      'contain',
      '[REDACTED]'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-014 imports runtime setup config without leaving the Operations route', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-setup-import',
        build_date: '2026-04-23',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: true,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/runtime/setup-import', (req) => {
      expect(req.body).to.deep.equal({
        source_path: 'C:\\cabinet\\recovery\\setup-import-source.json',
      })
      req.reply({
        statusCode: 200,
        body: {
          ok: true,
          setup_required: false,
          instance_name: 'Imported Recovery Instance',
          profile_key: 'imported-recovery',
          config_path: 'C:\\cabinet\\data\\cabinet.json',
          runtime_url: 'http://127.0.0.1:17880',
          runtime_port: 17880,
        },
      })
    }).as('setupImport')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-setup-import-source"]')
      .clear()
      .type('C:\\cabinet\\recovery\\setup-import-source.json')
    cy.get('[data-testid="settings-operations-setup-import-submit"]').click()
    cy.wait('@setupImport')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-setup-import-status"]').should(
      'contain',
      'Runtime setup imported successfully.'
    )
    cy.get('[data-testid="settings-operations-setup-import-summary"]').should(
      'contain',
      'Imported Recovery Instance'
    )
    cy.get('[data-testid="settings-operations-setup-import-summary"]').should(
      'contain',
      'imported-recovery'
    )
  })

  it('UI-SCREEN-SETTINGS-OPERATIONS-015 reports runtime setup import failure without leaving the Operations route', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-setup-import-error',
        build_date: '2026-04-23',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17880,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: true,
      },
    }).as('runtimeRecovery')
    cy.intercept('POST', '/api/runtime/setup-import', {
      statusCode: 400,
      body: {
        error: 'failed_to_import_setup_config',
      },
    }).as('setupImport')

    signInToOperations()
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')

    cy.get('[data-testid="settings-operations-setup-import-source"]')
      .clear()
      .type('C:\\cabinet\\recovery\\missing.json')
    cy.get('[data-testid="settings-operations-setup-import-submit"]').click()
    cy.wait('@setupImport')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.get('[data-testid="settings-operations-setup-import-status"]').should(
      'contain',
      'Runtime setup import failed.'
    )
  })
})
