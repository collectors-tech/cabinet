describe("ui-login-password-manager", () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.request("POST", "/api/test/runtime/setup-status", { state: "present" })
      .its("status")
      .should("eq", 200);
    cy.e2eEnsureSignedOut();
  });

  it("UI-LOGIN-PASSWORD-MANAGER-001 exposes standard sign-in autofill markup", () => {
    cy.visit("/sign-in");

    cy.get("form").within(() => {
      cy.get('input[name="email"]')
        .should("have.attr", "id", "email")
        .and("have.attr", "type", "email")
        .and("have.attr", "autocomplete", "username")
        .and("be.visible");

      cy.get('label[for="email"]').should("contain", "Email");

      cy.get('input[name="password"]')
        .should("have.attr", "id", "password")
        .and("have.attr", "type", "password")
        .and("have.attr", "autocomplete", "current-password")
        .and("be.visible");

      cy.get('label[for="password"]').should("contain", "Password");
      cy.contains("button[type='submit']", "Sign in").should("be.visible");
    });
  });
});
