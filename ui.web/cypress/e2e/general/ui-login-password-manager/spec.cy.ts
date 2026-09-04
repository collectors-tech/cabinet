describe("ui-login-password-manager", () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.request("POST", "/api/test/runtime/setup-status", { state: "present" })
      .its("status")
      .should("eq", 200);
    cy.e2eEnsureSignedOut();
  });

  it("UI-LOGIN-PASSWORD-MANAGER-001 keeps local-device sign-in out of password-manager autofill", () => {
    cy.visit("/sign-in");

    cy.get('[data-testid="local-device-auth-boundary"]')
      .should("be.visible")
      .and("contain.text", "Local device mode")
      .and("contain.text", "does not verify a")
      .and("contain.text", "password");
    cy.contains("button", "Open local workspace").should("be.visible");
    cy.get('[data-testid="identity-mode-indicator"]').should(
      "contain.text",
      "local-device"
    );
    cy.get("form").should("not.exist");
    cy.get('input[name="email"]').should("not.exist");
    cy.get('input[name="password"]').should("not.exist");
  });
});
