describe('UI-LOGIN-SESSION signed-out sign-in copy', () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true });
    cy.request('POST', '/api/test/runtime/setup-status', { state: 'present' })
      .its('status')
      .should('eq', 200);
  });

  it('UI-LOGIN-SESSION-008 keeps sign-in focused while preserving entry links', () => {
    cy.visit('/sign-in');

    cy.contains('Sign in to unlock your Cabinet workspace.').should(
      'not.exist'
    );
    cy.get('[data-testid="sign-in-profile-guidance"]').should('not.exist');
    cy.get('[data-testid="local-device-auth-boundary"]').should('be.visible');
    cy.get('input[name="email"]').should('not.exist');
    cy.get('input[name="password"]').should('not.exist');
    cy.contains('button', 'Open local workspace').should('be.visible');
    cy.contains('a', 'Create account').should('have.attr', 'href', '/sign-up');
    cy.contains('a', 'Forgot password?').should(
      'have.attr',
      'href',
      '/forgot-password'
    );
    cy.contains('a', 'Terms of Service').should('have.attr', 'href', '/terms');
    cy.contains('a', 'Privacy Policy').should('have.attr', 'href', '/privacy');
    cy.location('pathname').should('match', /^\/sign-in\/?$/);
  });
});
