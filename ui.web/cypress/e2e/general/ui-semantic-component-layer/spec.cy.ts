describe("ui-semantic-component-layer", () => {
  function signInToInventory() {
    cy.e2eReset();
    cy.e2eBootstrap();
    cy.e2eSetSetupState("present");
    cy.useBootstrappedProfile("e2e-profile-001", "E2E Local", {
      path: "/inventory/",
    });
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/inventory\/?$/);
  }

  it("UI-SEMANTIC-COMPONENT-LAYER-001 keeps shell/workspace/overlay layers present", () => {
    signInToInventory();
    cy.get('[data-slot="sidebar"]').should("be.visible");
    cy.get("header").first().should("be.visible");
    cy.contains("Collection Browser").should("not.exist");
    cy.contains(/Folders:\s*\d+/).should("be.visible");
    cy.get('[data-testid="profile-dropdown-trigger"]:visible').first().click();
    cy.contains('[data-slot="dropdown-menu-item"]', "Sign out").click();
    cy.contains(/sign out/i).should("be.visible");
  });

  it("UI-SEMANTIC-COMPONENT-LAYER-002 reuses shared primitives across sections", () => {
    signInToInventory();
    cy.get('[data-slot="card"]').its("length").should("be.greaterThan", 1);
    cy.get('[data-testid="inventory-new-action"]')
      .should("be.visible")
      .and("have.attr", "aria-label", "New item");
    cy.contains(/Folders:\s*\d+/).closest('[data-slot="card"]').should("be.visible");
  });

  it("UI-SEMANTIC-COMPONENT-LAYER-003 keeps shell stable across top-level route switches", () => {
    signInToInventory();
    cy.get('[data-slot="sidebar"]').should("be.visible");
    cy.visit("/wishlist");
    cy.get('[data-slot="sidebar"]').should("be.visible");
    cy.get("header").first().should("be.visible");
    cy.get('[data-testid="sidebar-nav-link-wishlist"]')
      .should("have.attr", "data-active", "true");
    cy.contains("Wishlist Sample Grail Chase").should("be.visible");
  });

  it("UI-SEMANTIC-COMPONENT-LAYER-004 handles deterministic loading/error/ready transitions", () => {
    let attempt = 0;
    cy.intercept("GET", "/api/items", (req) => {
      attempt += 1;
      if (attempt === 1) {
        req.reply({ statusCode: 500, body: { error: "items_failed" } });
        return;
      }
      req.reply({ statusCode: 200, body: { items: [] } });
    }).as("itemsTransitions");

    signInToInventory();
    cy.wait("@itemsTransitions");
    cy.get('[data-testid="inventory-load-error"]').should("be.visible");
    cy.contains("button", "Retry").click();
    cy.wait("@itemsTransitions");
    cy.get('[data-testid="inventory-load-error"]').should("not.exist");
    cy.contains("No results.").should("be.visible");
  });

  it("UI-SEMANTIC-COMPONENT-LAYER-005 reuses domain blocks in inventory and wishlist", () => {
    signInToInventory();
    cy.get('input[placeholder^="Filter by title"]').should("be.visible");
    cy.visit("/wishlist");
    cy.get('input[placeholder="Filter by title or part number..."]').should("be.visible");
    cy.get("table").should("be.visible");
  });

  it("UI-SEMANTIC-COMPONENT-LAYER-006 enforces overlay confirm guardrails", () => {
    signInToInventory();
    cy.get('[data-testid="profile-dropdown-trigger"]:visible').first().click();
    cy.contains('[data-slot="dropdown-menu-item"]', "Sign out").click();
    cy.contains(/sign out/i).should("be.visible");
    cy.contains("button", "Cancel").click();
    cy.contains(/sign out/i).should("not.exist");
  });

  it("UI-SEMANTIC-COMPONENT-LAYER-007 keeps keyboard-accessible shell interactions", () => {
    signInToInventory();
    cy.get('[data-testid="folder-tree-item-store-1"]').click();
    cy.get('[data-testid="collection-active-context"]').should("contain", "Store 1");
    cy.get('[data-testid="folder-tree-item-store-2"]').click();
    cy.get('[data-testid="collection-active-context"]').should("contain", "Store 2");
  });
});
