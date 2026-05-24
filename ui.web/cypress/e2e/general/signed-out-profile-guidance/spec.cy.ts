describe('UI-SCREEN-ONBOARDING-AUTH signed-out profile guidance', () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true });
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);
  });

  it('UI-SCREEN-ONBOARDING-AUTH-016 explains signed-out database/profile context before unlock', () => {
    cy.visit('/sign-in');

    cy.get('[data-testid="sign-in-profile-guidance"]')
      .should('be.visible')
      .and('contain.text', 'Sign in to unlock your Cabinet workspace')
      .and('contain.text', 'active database/profile')
      .and('contain.text', 'collections live inside that profile');
    cy.get('[data-testid="sign-in-profile-guidance"]')
      .contains('a', 'Create account')
      .should('have.attr', 'href', '/sign-up');
    cy.location('pathname').should('match', /^\/sign-in\/?$/);
  });
});
