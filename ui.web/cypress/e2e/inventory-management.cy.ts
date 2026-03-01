describe("inventory-management", () => {
  function signIn() {
    cy.visit("/sign-in");
    cy.get('input[name="email"]').clear().type("e2e-inventory@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^(\/|\/_authenticated\/?)$/
    );
  }

  it("renders inventory workspace, supports view toggle and filtering, and avoids 500", () => {
    signIn();

    cy.visit("/inventory/");
    cy.contains("500").should("not.exist");
    cy.contains("Oops! Something went wrong").should("not.exist");

    cy.contains("Inventory").should("be.visible");
    cy.contains("Collection Browser").should("be.visible");
    cy.contains("button", "Add Item").should("be.visible");
    cy.contains("button", "Add Folder").should("be.visible");

    cy.contains("button", "Cards").click();
    cy.contains("Status:").should("be.visible");

    cy.contains("button", "Rows").click();
    cy.get("table").should("be.visible");

    cy.get('input[placeholder="Filter by title or ID..."]').type(
      "no-matching-task-xyz"
    );
    cy.contains("No results.").should("be.visible");
  });
});
