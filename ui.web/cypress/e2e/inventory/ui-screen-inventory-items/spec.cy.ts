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
          {
            id: "item-2",
            part_number: "PN-002",
            title: "Second Item",
            status: "used",
            category: "documentation",
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
    cy.contains("button", "New").should("be.visible");
    cy.contains("button", "Create").should("be.visible");

    cy.contains("button", "Cards").click();
    cy.contains("Status:").should("be.visible");

    cy.contains("button", "Rows").click();
    cy.get("table").should("be.visible");
    cy.contains("th", "Part #").should("be.visible");
    cy.contains("th", "Title").should("be.visible");
    cy.contains("th", "Condition").should("be.visible");
    cy.contains("th", "Category").should("be.visible");
    cy.contains("th", "Task").should("not.exist");
    cy.contains("th", "Priority").should("not.exist");
    cy.contains("PN-001").should("be.visible");
    cy.contains("todo").should("be.visible");
    cy.contains("feature").should("be.visible");

    cy.get('input[placeholder="Filter by title or part number..."]').type(
      "no-matching-task-xyz"
    );
    cy.contains("No results.").should("be.visible");
  });

  it("renders empty inventory state without global 500 fallback", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsEmpty");

    signIn();
    cy.wait("@itemsEmpty");

    cy.contains("500").should("not.exist");
    cy.contains("Oops! Something went wrong").should("not.exist");
    cy.contains("Inventory").should("be.visible");
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

  it("UI-SCREEN-INVENTORY-ITEMS-004 keeps summary compact in Collection Browser header and removes duplicate strips", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-compact-1",
            part_number: "PN-COMPACT-1",
            title: "Compact Layout Item",
            status: "todo",
            category: "feature",
          },
        ],
      },
    }).as("itemsCompact");

    signIn();
    cy.wait("@itemsCompact");

    cy.contains("Command Row").should("not.exist");
    cy.contains("Summary Strip").should("not.exist");

    cy.contains("Collection Browser")
      .closest('[data-slot="card"]')
      .within(() => {
        cy.contains(/Folders:\s*\d+/).should("be.visible");
        cy.contains(/Items:\s*\d+/).should("be.visible");
        cy.contains(/Active Brand:\s*\w+/).should("be.visible");
        cy.contains(/Active Category:\s*\w+/).should("be.visible");

        cy.contains(/Active Category:/)
          .should("be.visible")
          .then(($summary) => {
            const summaryTop = $summary[0].getBoundingClientRect().top;
            cy.get('input[placeholder="Filter by title or part number..."]')
              .should("be.visible")
              .then(($input) => {
                const inputTop = $input[0].getBoundingClientRect().top;
                expect(summaryTop).to.be.lessThan(inputTop);
              });
          });
      });
  });

  it("UI-SCREEN-INVENTORY-ITEMS-005 shows New action with adjacent Create menu", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: [] },
    }).as("itemsActions");

    signIn();
    cy.wait("@itemsActions");

    cy.get('[data-testid="inventory-new-action"]')
      .should("be.visible")
      .and("contain", "New")
      .click();
    cy.get('[data-testid="inventory-item-create-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-item-create-cancel"]').click();

    cy.get('[data-testid="inventory-create-menu-trigger"]')
      .should("be.visible")
      .and("contain", "Create")
      .click();
    cy.get('[data-testid="inventory-create-menu-item"]').should("be.visible").click();
    cy.get('[data-testid="inventory-item-create-dialog"]').should("be.visible");
    cy.get('[data-testid="inventory-create-menu-folder"]').should("not.exist");
  });
+
+  it("UI-SCREEN-INVENTORY-ITEMS-007 opens create-item workflow from toolbar", () => {
+    cy.intercept("GET", "/api/items", {
+      statusCode: 200,
+      body: { items: [] },
+    }).as("itemsCreate");
+
+    signIn();
+    cy.wait("@itemsCreate");
+
+    cy.get('[data-testid="inventory-create-menu-trigger"]').click();
+    cy.get('[data-testid="inventory-create-menu-item"]').click();
+    cy.get('[data-testid="inventory-item-create-title"]').type("Inline Created Item");
+    cy.get('[data-testid="inventory-item-create-part-number"]').type("PN-CREATE-1");
+    cy.get('[data-testid="inventory-item-create-submit"]').click();
+
+    cy.get('[data-testid="inventory-item-create-dialog"]').should("not.exist");
+    cy.contains("Inline Created Item").should("be.visible");
+    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-CREATE-1");
+  });

  it("UI-SCREEN-INVENTORY-ITEMS-006 creates collection inline and auto-selects it", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-inline-1",
            part_number: "PN-INLINE-1",
            title: "Inline Collection Item",
            status: "todo",
            category: "feature",
          },
        ],
      },
    }).as("itemsInline");

    signIn();
    cy.wait("@itemsInline");

    cy.get('[data-testid="collection-inline-add-new"]').click();
    cy.get('[data-testid="collection-inline-new-name"]').type("Inline Alpha");
    cy.get('[data-testid="collection-inline-save"]').click();
    cy.get('[data-testid="collection-inline-picker-selected"]').should(
      "contain",
      "Inline Alpha"
    );
  });
});
