describe('ONBOARDING-STARTER-DATA', () => {
  let activeProfileId = '';

  beforeEach(() => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true }).then((state) => {
      activeProfileId = state.profile_id;
    });
  });

  it('ONBOARDING-STARTER-DATA-001 gates sign-in behind setup wizard when runtime setup config is missing', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'missing' })
      .its('status')
      .should('eq', 200);
    cy.visit('/sign-in');

    cy.contains('Setup Wizard').should('be.visible');
    cy.get('[data-testid="setup-start"]').should('be.visible');
    cy.get('[data-testid="setup-use-defaults"]').should('be.visible');
    cy.contains('Sign in').should('not.exist');
  });

  it('ONBOARDING-STARTER-DATA-002 completes setup and restores sign-in controls deterministically', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'missing' })
      .its('status')
      .should('eq', 200);
    cy.visit('/sign-in');
    cy.get('[data-testid="setup-use-defaults"]').click();
    cy.get('[data-testid="setup-wizard-complete-state"]').should('be.visible');
    cy.contains('button', 'Finish').click();
    cy.contains('Sign in').should('be.visible');
  });

  it('ONBOARDING-STARTER-DATA-003 bypasses setup wizard when runtime config is already present', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);
    cy.visit('/sign-in');
    cy.contains('Sign in').should('be.visible');
    cy.contains('Setup Wizard').should('not.exist');
  });

  it('ONBOARDING-STARTER-DATA-004 seeds showcase sample data with persisted provenance', () => {
    let firstTotalItems = 0;
    let firstTotalWishlist = 0;
    let sampleDisclosure = '';

    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);

    cy.request('POST', '/api/onboarding/sample-data')
      .its('body')
      .then((seed) => {
        expect(seed.dataset_kind).to.eq('sample_showcase');
        expect(seed.dataset_label).to.match(/sample/i);
        expect(seed.sample_data_disclosure).to.match(/example records/i);
        expect(seed.already_seeded_for_profile).to.eq(false);
        expect(seed.created_items).to.be.greaterThan(0);
        expect(seed.created_wishlist_entries).to.be.greaterThan(0);
        expect(seed.created_photos).to.be.greaterThan(0);
        expect(seed.total_items).to.be.at.least(30);
        expect(seed.total_wishlist_entries).to.be.at.least(3);

        firstTotalItems = seed.total_items;
        firstTotalWishlist = seed.total_wishlist_entries;
        sampleDisclosure = seed.sample_data_disclosure;
      });

    cy.request('GET', `/api/profiles/${activeProfileId}/settings`)
      .its('body.settings')
      .then((settings) => {
        expect(settings['onboarding.sample_data.dataset_kind']).to.eq('sample_showcase');
        expect(settings['onboarding.sample_data.disclosure']).to.eq(sampleDisclosure);
      });

    cy.request('POST', '/api/onboarding/sample-data')
      .its('body')
      .then((seed) => {
        expect(seed.dataset_kind).to.eq('sample_showcase');
        expect(seed.sample_data_disclosure).to.eq(sampleDisclosure);
        expect(seed.already_seeded_for_profile).to.eq(true);
        expect(seed.created_items).to.eq(0);
        expect(seed.created_wishlist_entries).to.eq(0);
        expect(seed.total_items).to.eq(firstTotalItems);
        expect(seed.total_wishlist_entries).to.eq(firstTotalWishlist);
      });
  });
});
