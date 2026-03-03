describe('ONBOARDING-STARTER-DATA', () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true });
  });

  it('ONBOARDING-STARTER-DATA-001 gates sign-in behind setup wizard when runtime setup config is missing', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'missing' })
      .its('status')
      .should('eq', 200);
    cy.visit('/sign-in');

    cy.contains('Setup Wizard').should('be.visible');
    cy.contains('button', 'Complete Setup').should('be.visible');
    cy.contains('Sign in').should('not.exist');
  });

  it('ONBOARDING-STARTER-DATA-002 completes setup and restores sign-in controls deterministically', () => {
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'missing' })
      .its('status')
      .should('eq', 200);
    cy.visit('/sign-in');
    cy.contains('button', 'Complete Setup').click();
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
});
