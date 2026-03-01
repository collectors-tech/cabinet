describe("ui-login-session", () => {
  it("UI-LOGIN-SESSION-001 redirects unauthenticated access to sign-in and returns to target after login", () => {
    cy.clearCookies();
    cy.clearLocalStorage();

    cy.visit("/inventory/");
    cy.location("pathname").should("eq", "/sign-in");
    cy.location("search").should("include", "redirect=");
    cy.location("search").should("include", "%2Finventory%2F");

    cy.get('input[name="email"]').type("e2e-login-session@example.com");
    cy.get('input[name="password"]').type("password123");
    cy.contains("button", "Sign in").click();

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/inventory\/?$/
    );
  });

  it("UI-LOGIN-SESSION-002 keeps inline validation errors and allows retry without refresh", () => {
    cy.clearCookies();
    cy.clearLocalStorage();

    cy.visit("/sign-in");
    cy.get('input[name="email"]').type("invalid-email");
    cy.get('input[name="password"]').type("short");
    cy.contains("button", "Sign in").click();

    cy.location("pathname").should("eq", "/sign-in");
    cy.contains(/invalid email|please enter your email/i).should("be.visible");
    cy.contains("Password must be at least 7 characters long").should(
      "be.visible"
    );

    cy.get('input[name="email"]').clear().type("e2e-login-session@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();

    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^(\/|\/inventory\/?|\/_authenticated\/?)$/
    );
  });
});

