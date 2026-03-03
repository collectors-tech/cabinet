describe('SETUP-WIZ', () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true });
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'missing' })
      .its('status')
      .should('eq', 200);
    cy.intercept('GET', '/api/runtime/setup-status').as('setupStatus');
    cy.visit('/sign-in');
    cy.wait('@setupStatus');
    cy.get('[data-testid="setup-wizard"]').should('be.visible');
  });

  it('UC-SW-06 setup-wizard-progress-template shows step header, percentage, and footer actions', () => {
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 1 OF 3');
    cy.get('[data-testid="setup-step-percent"]').should('contain.text', '33%');
    cy.get('[data-testid="setup-progress-bar"]')
      .should('have.attr', 'aria-valuenow')
      .and('eq', '33');
    cy.get('[data-testid="setup-prev"]').should('be.disabled');
    cy.get('[data-testid="setup-next"]').should('be.visible').and('not.be.disabled');
    cy.get('[data-testid="setup-complete"]').should('not.exist');

    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 2 OF 3');
    cy.get('[data-testid="setup-step-percent"]').should('contain.text', '67%');
    cy.get('[data-testid="setup-prev"]').should('not.be.disabled');
  });

  it('UC-SW-07 setup-wizard-complete-to-launch transitions to config complete with start action', () => {
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-step-indicator"]').should('contain.text', 'STEP 3 OF 3');
    cy.get('[data-testid="setup-complete"]').should('be.visible').click();

    cy.contains('Config complete').should('be.visible');
    cy.contains(/Start App|Open Cabinet/).should('be.visible');
    cy.contains(/registration-success|email activation/i).should('not.exist');
  });

  it('UC-SW-08 setup-wizard-config-schema-write persists deterministic cabinet.json payload', () => {
    cy.get('[data-testid="setup-instance-name"]').clear().type('Primary');
    cy.get('[data-testid="setup-profile-key"]').clear().type('primary');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();

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
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-auth-mode"]').select('clerk');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
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

    cy.get('[data-testid="setup-prev"]').click();
    cy.get('[data-testid="setup-clerk-publishable-key"]')
      .clear()
      .type('pk_test_123');
    cy.get('[data-testid="setup-next"]').click();
    cy.get('[data-testid="setup-complete"]').click();
    cy.contains('Config complete').should('be.visible');
  });
});
