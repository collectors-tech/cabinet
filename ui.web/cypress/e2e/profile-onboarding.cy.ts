describe("profile-onboarding", () => {
  it("supports auth entry validation and successful workspace unlock", () => {
    cy.visit("/sign-in");

    cy.contains("Sign in").should("be.visible");

    cy.get('input[name="email"]').type("invalid-email");
    cy.get('input[name="password"]').type("short");
    cy.contains("button", "Sign in").click();

    cy.location("pathname").should("eq", "/sign-in");
    cy.contains(/invalid email|please enter your email/i).should("be.visible");
    cy.contains("Password must be at least 7 characters long").should(
      "be.visible"
    );

    cy.get('input[name="email"]').clear().type("e2e-onboarding@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^(\/|\/_authenticated\/?)$/
    );
    cy.contains("Home").should("be.visible");
  });
});
