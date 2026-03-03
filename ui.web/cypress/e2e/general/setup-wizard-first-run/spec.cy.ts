describe('SETUP-WIZ-007', () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true });
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'missing' })
      .its('status')
      .should('eq', 200);
  });

  it('UC-SW-08 setup-wizard-config-schema-write persists deterministic cabinet.json payload', () => {
    cy.intercept('GET', '/api/runtime/setup-status').as('setupStatus');
    cy.visit('/sign-in');
    cy.wait('@setupStatus');
    cy.get('[data-testid="setup-wizard"]').should('be.visible');

    cy.get('[data-testid="setup-instance-name"]').clear().type('Primary');
    cy.get('[data-testid="setup-profile-key"]').clear().type('primary');
    cy.get('[data-testid="setup-wizard-complete"]').click();

    cy.request('GET', '/api/test/runtime/setup-config')
      .its('body')
      .then((payload) => {
        expect(payload.instance.name).to.eq('Primary');
        expect(payload.instance.profile).to.eq('primary');
        expect(payload.storage.dataDir).to.be.a('string').and.not.empty;
        expect(payload.storage.mediaDir).to.be.a('string').and.not.empty;
        expect(payload.runtime.portMode).to.be.oneOf(['auto', 'fixed']);
        expect(payload.runtime.resolvedUrl).to.match(
          /^http:\/\/(127\.0\.0\.1|0\.0\.0\.0):/
        );
        expect(payload.auth.mode).to.eq('local');
        expect(payload.auth.clerk.enabled).to.eq(false);
        expect(payload.bootstrap.workspace).to.be.a('string').and.not.empty;
        expect(payload.features.chat).to.eq(true);
        expect(payload.features.providers).to.eq(true);
        expect(payload.features.scanner).to.eq(true);
        expect(payload.meta.wizardVersion).to.eq('1');
      });
  });

  it('UC-SW-09 setup-wizard-clerk-required-fields blocks completion when clerk key is missing', () => {
    cy.intercept('GET', '/api/runtime/setup-status').as('setupStatus');
    cy.visit('/sign-in');
    cy.wait('@setupStatus');
    cy.get('[data-testid="setup-wizard"]').should('be.visible');

    cy.get('[data-testid="setup-auth-mode"]').select('clerk');
    cy.get('[data-testid="setup-wizard-complete"]').click();
    cy.get('[data-testid="setup-wizard-error"]')
      .should('be.visible')
      .and('contain.text', 'Clerk publishable key is required');

    cy.request({
      method: 'GET',
      url: '/api/test/runtime/setup-config',
      failOnStatusCode: false,
    }).then((response) => {
      expect(response.status).to.eq(404);
    });

    cy.get('[data-testid="setup-clerk-publishable-key"]')
      .clear()
      .type('pk_test_123');
    cy.get('[data-testid="setup-wizard-complete"]').click();
    cy.contains('Sign in').should('be.visible');
  });
});
