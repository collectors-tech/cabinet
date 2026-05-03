describe("inventory-shell", () => {
  function signIn() {
    cy.visit("/sign-in?redirect=%2Finventory%2F");
    cy.get('input[name="email"]').clear().type("e2e-inventory@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/inventory\/?$/
    );
  }

  it("UI-SCREEN-INVENTORY-SHELL-001 renders resolved inventory intro copy", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsIntro");

    signIn();
    cy.wait("@itemsIntro");

    cy.contains("Inventory").should("be.visible");
    cy.contains("Browse, organize, and update the items you already own.").should(
      "be.visible"
    );
    cy.contains("inventory.description").should("not.exist");
  });
});
