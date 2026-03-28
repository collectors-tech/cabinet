describe("ui-screen-collections", () => {
  function signInToCollections() {
    cy.visit("/sign-in?redirect=%2Fcollections%2F");
    cy.get('input[name="email"]').clear().type("e2e-collections@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/collections\/?$/
    );
  }

  it("UI-SCREEN-COLLECTIONS-001 renders top-level collections section with list and New action", () => {
    signInToCollections();

    cy.get('[data-testid="collections-section"]').should("be.visible");
    cy.get('[data-testid="collections-item-all-items"]').should("be.visible");
    cy.get('[data-testid="collections-new-action"]').should("be.visible");
  });

  it("UI-SCREEN-COLLECTIONS-002 exposes New flow and adjacent Create menu quick actions", () => {
    signInToCollections();

    cy.get('[data-testid="collections-create-guidance"]').should("be.visible");
    cy.get('[data-testid="collections-new-action"]').click();
    cy.get('[data-testid="collections-create-panel"]').should("be.visible");
    cy.get('[data-testid="collections-create-panel-description"]')
      .should("contain.text", "Saving creates the collection immediately")
    cy.contains("New opens the primary create flow for a named collection.")
      .should("be.visible")
    cy.get('[data-testid="collections-create-outcome"]').should("not.exist")

    cy.get('[data-testid="collections-create-menu-trigger"]').click();
    cy.get('[data-testid="collections-create-menu-new"]').should("be.visible");
    cy.get('[data-testid="collections-create-menu-starter"]').should("be.visible");
  });

  it("UI-SCREEN-COLLECTIONS-003 supports inline-style quick create and auto-activates new collection", () => {
    signInToCollections();

    cy.get('[data-testid="collections-new-action"]').click();
    cy.get('[data-testid="collections-new-input"]').type("Collections Inline Alpha");
    cy.get('[data-testid="collections-new-save"]').click();

    cy.contains("Collections Inline Alpha created and set as the active collection.")
      .should("be.visible")
    cy.get('[data-testid="collections-create-outcome"]').should("not.exist")
    cy.get('[data-testid="collections-item-collections-inline-alpha"]')
      .should("be.visible")
      .and("have.attr", "data-state", "active");
  });

  it("UI-SCREEN-COLLECTIONS-008 makes create outcomes and validation visible", () => {
    signInToCollections();

    cy.get('[data-testid="collections-new-action"]').click();
    cy.get('[data-testid="collections-new-save"]').click();
    cy.get('[data-testid="collections-create-error"]')
      .should("contain.text", "Enter a collection name before saving.")

    cy.get('[data-testid="collections-new-input"]').type("Collections Visible Beta");
    cy.get('[data-testid="collections-new-save"]').click();
    cy.get('[data-testid="collections-create-error"]').should("not.exist");
    cy.contains("Collections Visible Beta created and set as the active collection.")
      .should("be.visible")
    cy.get('[data-testid="collections-create-outcome"]').should("not.exist")

    cy.get('[data-testid="collections-create-menu-trigger"]').click();
    cy.get('[data-testid="collections-create-menu-starter"]').click();
    cy.contains("Starter collections added.").should("be.visible")
    cy.get('[data-testid="collections-create-outcome"]').should("not.exist")
  });

  it("UI-SCREEN-COLLECTIONS-006 exposes collection details and metadata summaries before selection", () => {
    signInToCollections();

    cy.get('[data-testid="collections-item-title-all-items"]')
      .should("contain.text", "All Items")
    cy.get('[data-testid="collections-item-description-all-items"]')
      .should("contain.text", "Everything currently tracked")
    cy.get('[data-testid="collections-item-metadata-all-items"]')
      .should("contain.text", "248 items")
      .and("contain.text", "Workspace default")
      .and("contain.text", "Updated 5m ago")
    cy.get('[data-testid="collections-item-status-all-items"]')
      .should("contain.text", "Broadest scope")

    cy.get('[data-testid="collections-item-description-watch-list"]')
      .should("contain.text", "Fast-moving")
    cy.get('[data-testid="collections-item-status-watch-list"]')
      .should("contain.text", "Needs review")
  });

  it("UI-SCREEN-COLLECTIONS-007 supports search, filtering, and ordering tools", () => {
    signInToCollections();

    cy.get('[data-testid="collections-management-tools"]').should("be.visible");
    cy.get('[data-testid="collections-management-summary"]')
      .should("contain.text", "Showing 6 of 6 collections.")

    cy.get('[data-testid="collections-search-input"]').type("watch");
    cy.get('[data-testid="collections-item-watch-list"]').should("be.visible");
    cy.get('[data-testid="collections-item-all-items"]').should("not.exist");
    cy.get('[data-testid="collections-management-summary"]')
      .should("contain.text", "Showing 1 of 6 collections.")

    cy.get('[data-testid="collections-search-input"]').clear();
    cy.get('[data-testid="collections-filter-storage"]').click();
    cy.get('[data-testid="collections-item-warehouse-1"]').should("be.visible");
    cy.get('[data-testid="collections-item-store-1"]').should("not.exist");

    cy.get('[data-testid="collections-filter-all"]').click();
    cy.get('[data-testid="collections-sort-items-desc"]').click();
    cy.get('[data-testid^="collections-item-"]').then(($items) => {
      const first = $items.first().attr('data-testid');
      expect(first).to.eq('collections-item-all-items');
    });
  });

  it("UI-SCREEN-COLLECTIONS-005 supports rename and remove actions with visible outcomes", () => {
    signInToCollections();

    cy.get('[data-testid="collections-rename-trigger-store-2"]').click();
    cy.get('[data-testid="collections-rename-panel"]').should("be.visible");
    cy.get('[data-testid="collections-rename-input"]').clear().type("Store 2 Prime");
    cy.get('[data-testid="collections-rename-save"]').click();
    cy.contains("Store 2 renamed to Store 2 Prime.").should("be.visible")
    cy.get('[data-testid="collections-create-outcome"]').should("not.exist")
    cy.get('[data-testid="collections-item-store-2-prime"]').should("be.visible");
    cy.get('[data-testid="collections-item-store-2"]').should("not.exist");

    cy.get('[data-testid="collections-remove-trigger-store-2-prime"]').click();
    cy.get('[data-testid="collections-remove-panel"]').should("be.visible");
    cy.get('[data-testid="collections-remove-confirm"]').click();
    cy.contains("Store 2 Prime removed from workspace collections.").should("be.visible")
    cy.get('[data-testid="collections-create-outcome"]').should("not.exist")
    cy.get('[data-testid="collections-item-store-2-prime"]').should("not.exist");
  });

  it("UI-SCREEN-COLLECTIONS-004 shows active collection context changes and persistence semantics", () => {
    signInToCollections();

    cy.get('[data-testid="collections-active-context-panel"]').should("be.visible");
    cy.get('[data-testid="collections-active-context-name"]')
      .should("contain.text", "All Items");
    cy.get('[data-testid="collections-active-context-persistence"]')
      .should("contain.text", "Persists for this signed-in profile");

    cy.get('[data-testid="collections-item-watch-list"] button').first().click();
    cy.get('[data-testid="collections-item-watch-list"]')
      .should("have.attr", "data-state", "active");
    cy.get('[data-testid="collections-active-context-name"]')
      .should("contain.text", "Watch List");
    cy.get('[data-testid="collections-active-context-message"]')
      .should("contain.text", "Active collection is Watch List")

    cy.reload();
    cy.get('[data-testid="collections-active-context-name"]')
      .should("contain.text", "Watch List");
    cy.get('[data-testid="collections-item-watch-list"]')
      .should("have.attr", "data-state", "active");
  });

  it("UI-SCREEN-COLLECTIONS-009 uses tag iconography for collections navigation and page identity", () => {
    signInToCollections();

    cy.get('[data-testid="sidebar-nav-link-collections"]').should("be.visible");
    cy.get('[data-testid="collections-page-icon"]').should("be.visible");

    cy.get('[data-testid="sidebar-nav-link-collections"] svg')
      .should("have.attr", "data-lucide", "tag");
    cy.get('[data-testid="collections-page-icon"]')
      .should("have.attr", "data-lucide", "tag");
  });
});
