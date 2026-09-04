describe('settings/categories', () => {
  function signInToCategories() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap({ minimalProfile: true }).then(
      ({ profile_id, profile_name }) => {
        cy.useBootstrappedProfile(profile_id, profile_name, {
          path: '/settings/categories/',
        })
      }
    )
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('UI-SCREEN-SETTINGS-CATEGORIES-001 manages reusable taxonomy settings for the active profile', () => {
    cy.intercept('GET', '/api/profiles/e2e-profile-001/settings', {
      statusCode: 200,
      body: {
        settings: {
          'inventory.category-options.v1': JSON.stringify([
            'General',
            'Cars',
            'Model Kit',
          ]),
          'grading.enums.packaging': JSON.stringify([
            'Boxed',
            'Loose',
          ]),
          'grading.enums.item_type_condition_scales': JSON.stringify([
            {
              item_type: 'Model Car',
              conditions: ['Mint', 'Used'],
            },
          ]),
        },
      },
    }).as('profileSettings')

    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings', (req) => {
      const settings = req.body?.settings ?? {}
      const categories = JSON.parse(
        settings['inventory.category-options.v1'] ?? '[]'
      )
      const packagingGrades = JSON.parse(
        settings['grading.enums.packaging'] ?? '[]'
      )
      const itemTypeScales = JSON.parse(
        settings['grading.enums.item_type_condition_scales'] ?? '[]'
      )
      expect(categories).to.deep.equal(['General', 'Model Kit', 'Garage Kit'])
      expect(packagingGrades).to.deep.equal(['Boxed', 'Display'])
      expect(itemTypeScales).to.deep.include({
        item_type: 'Slot Car',
        conditions: ['New', 'Used'],
      })
      req.reply({ statusCode: 200, body: { settings } })
    }).as('saveProfileSettings')

    signInToCategories()
    cy.wait('@profileSettings')

    cy.contains('h3', 'Categories').should('be.visible')
    cy.get('[data-testid="settings-categories-list"]').should('contain', 'Cars')
    cy.get('[data-testid="settings-categories-new"]').type('Garage Kit')
    cy.get('[data-testid="settings-categories-add"]').click()
    cy.get('[data-testid="settings-category-remove-Cars"]').click()
    cy.get('[data-testid="settings-packaging-grade-new"]').type('Display')
    cy.get('[data-testid="settings-packaging-grade-add"]').click()
    cy.get('[data-testid="settings-packaging-grade-remove-Loose"]').click()
    cy.get('[data-testid="settings-item-type-new"]').type('Slot Car')
    cy.get('[data-testid="settings-item-type-add"]').click()
    cy.get('[data-testid="settings-categories-save"]').click()
    cy.wait('@saveProfileSettings')
    cy.get('[data-testid="settings-categories-status"]').should(
      'contain',
      'Saved categories'
    )
  })

  it('UI-SCREEN-SETTINGS-CATEGORIES-002 preserves taxonomy edits when save fails', () => {
    cy.intercept('GET', '/api/profiles/e2e-profile-001/settings', {
      statusCode: 200,
      body: {
        settings: {
          'inventory.category-options.v1': JSON.stringify([
            'General',
            'Cars',
          ]),
          'grading.enums.packaging': JSON.stringify([
            'Boxed',
            'Loose',
          ]),
          'grading.enums.item_type_condition_scales': JSON.stringify([
            {
              item_type: 'Model Car',
              conditions: ['Mint', 'Used'],
            },
          ]),
        },
      },
    }).as('profileSettings')

    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings', {
      statusCode: 503,
      body: { error: 'taxonomy_settings_save_unavailable' },
    }).as('saveProfileSettingsFailure')

    signInToCategories()
    cy.wait('@profileSettings')

    cy.get('[data-testid="settings-categories-new"]').type('Garage Kit')
    cy.get('[data-testid="settings-categories-add"]').click()
    cy.get('[data-testid="settings-category-remove-Cars"]').click()
    cy.get('[data-testid="settings-packaging-grade-new"]').type('Display')
    cy.get('[data-testid="settings-packaging-grade-add"]').click()
    cy.get('[data-testid="settings-packaging-grade-remove-Loose"]').click()
    cy.get('[data-testid="settings-item-type-new"]').type('Slot Car')
    cy.get('[data-testid="settings-item-type-add"]').click()
    cy.get('[data-testid="settings-categories-save"]').click()

    cy.wait('@saveProfileSettingsFailure')
      .its('request.body.settings')
      .then((settings) => {
        expect(
          JSON.parse(settings['inventory.category-options.v1'])
        ).to.deep.equal(['General', 'Garage Kit'])
        expect(
          JSON.parse(settings['grading.enums.packaging'])
        ).to.deep.equal(['Boxed', 'Display'])
        expect(
          JSON.parse(settings['grading.enums.item_type_condition_scales'])
        ).to.deep.include({
          item_type: 'Slot Car',
          conditions: ['New', 'Used'],
        })
      })
    cy.get('[data-testid="settings-categories-error"]').should(
      'contain',
      'profile_settings_save_503'
    )
    cy.get('[data-testid="settings-categories-status"]').should('not.exist')
    cy.location('pathname').should('match', /^\/settings\/categories\/?$/)
    cy.get('[data-testid="settings-categories-list"]')
      .should('contain', 'Garage Kit')
      .and('not.contain', 'Cars')
    cy.get('[data-testid="settings-packaging-grades-list"]')
      .should('contain', 'Display')
      .and('not.contain', 'Loose')
    cy.get('[data-testid="settings-item-type-scales-list"]')
      .should('contain', 'Slot Car')
    cy.get('[data-testid="settings-categories-save"]').should('not.be.disabled')
  })

  it('UI-SCREEN-SETTINGS-CATEGORIES-003 blocks taxonomy edits when active profile is missing', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 404,
      body: { error: 'active_profile_404' },
    }).as('activeProfileMissing')

    signInToCategories()
    cy.wait('@activeProfileMissing')

    cy.location('pathname').should('match', /^\/settings\/categories\/?$/)
    cy.get('[data-testid="settings-profile-context-blocked"]').should(
      'be.visible'
    )
    cy.contains('Active profile is required.').should('be.visible')
    cy.contains('button', 'Retry').should('be.visible')
    cy.contains('a', 'Create or Select Profile').should('be.visible')
    cy.get('[data-testid="settings-categories-new"]').should('be.disabled')
    cy.get('[data-testid="settings-categories-add"]').should('be.disabled')
    cy.get('[data-testid="settings-packaging-grade-new"]').should(
      'be.disabled'
    )
    cy.get('[data-testid="settings-packaging-grade-add"]').should(
      'be.disabled'
    )
    cy.get('[data-testid="settings-item-type-new"]').should('be.disabled')
    cy.get('[data-testid="settings-item-type-add"]').should('be.disabled')
    cy.get('[data-testid="settings-category-remove-General"]').should(
      'be.disabled'
    )
    cy.get('[data-testid="settings-categories-save"]').should('be.disabled')

    cy.contains('button', 'Retry').click()
    cy.wait('@activeProfileMissing')
    cy.location('pathname').should('match', /^\/settings\/categories\/?$/)
  })
})
