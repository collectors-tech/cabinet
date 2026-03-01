describe("inventory-management", () => {
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

  it("renders inventory workspace, supports view toggle and filtering, and avoids 500", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-1",
            part_number: "PN-001",
            title: "Starter Item",
            status: "todo",
            category: "feature",
          },
        ],
      },
    }).as("items");
    signIn();
    cy.wait("@items");
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

  it("UI-SCREEN-INVENTORY-ITEMS-002 shows inline error state and recovers on retry", () => {
    let attempts = 0;
    cy.intercept("GET", "/api/items", (req) => {
      attempts += 1;
      if (attempts === 1) {
        req.reply({
          statusCode: 500,
          body: { error: "failed_to_list_items" },
        });
        return;
      }
      req.reply({
        statusCode: 200,
        body: {
          items: [
            {
              id: "item-retry-1",
              part_number: "PN-RETRY-1",
              title: "Recovered Item",
              status: "todo",
              category: "feature",
            },
          ],
        },
      });
    }).as("itemsRetry");

    signIn();
    cy.wait("@itemsRetry");
    cy.get('[data-testid="inventory-load-error"]').should("be.visible");
    cy.contains("button", "Retry").click();
    cy.wait("@itemsRetry");
    cy.get('[data-testid="inventory-load-error"]').should("not.exist");
    cy.contains("Recovered Item").should("be.visible");
    cy.contains("500").should("not.exist");
  });

  it("UI-SCREEN-INVENTORY-ITEMS-003 remains deterministic with bulk dataset filtering", () => {
    const bulk = Array.from({ length: 1200 }, (_, index) => ({
      id: `item-${index + 1}`,
      part_number: `PN-${index + 1}`,
      title: `Bulk Item ${index + 1}`,
      status: "todo",
      category: "feature",
    }));
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: bulk },
    }).as("itemsBulk");

    signIn();
    cy.wait("@itemsBulk");
    cy.contains("Items:").parent().contains("1200").should("be.visible");
    cy.contains("Page 1 of 120").should("exist");
    cy.contains("button", "Cards").click();
    cy.contains("Status:").should("be.visible");
    cy.contains("button", "Rows").click();
    cy.get("table").should("be.visible");
  });
});
