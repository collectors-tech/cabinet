describe("ui-screen-wishlist", () => {
  function stubWishlistData() {
    cy.intercept("GET", "/api/wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "wish-1",
            item_id: "item-collector-1",
            priority: "medium",
            below_target_now: false,
          },
          {
            id: "wish-2",
            item_id: "item-collector-2",
            priority: "high",
            below_target_now: true,
          },
        ],
      },
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-collector-1",
            title: "AFX Mega-G+ Camaro Wildfire",
            part_number: "22073",
          },
          {
            id: "item-collector-2",
            title: "F1 Silverline",
            part_number: "F1002",
          },
        ],
      },
    }).as("catalogItems");
  }

  function signInToWishlist() {
    stubWishlistData();
    cy.visit("/sign-in?redirect=%2Fwishlist%2F");
    cy.get('input[name="email"]').clear().type("e2e-wishlist@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/wishlist\/?$/
    );
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
  }

  it("UI-SCREEN-WISHLIST-001 filters list and persists row/card view mode", () => {
    signInToWishlist();

    cy.contains("Wishlist").should("be.visible");
    cy.get("table").should("be.visible");
    cy.contains("button", "Cards").click();
    cy.window().its("localStorage").invoke("getItem", "cabinet.viewMode.wishlist").should("eq", "cards");
    cy.contains("Status:").should("be.visible");
    cy.reload();
    cy.contains("Status:").should("be.visible");

    cy.contains("button", "Rows").click();
    cy.get('input[placeholder="Filter by title or ID..."]').type("no-match-wishlist");
    cy.contains("No results.").should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-003 supports multi-select with bulk action toolbar", () => {
    signInToWishlist();

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.get('button[aria-label="Select all"]').click();

    cy.contains(/selected/i).should("be.visible");
    cy.get('button[aria-label="Update status"]').should("be.visible");
    cy.get('button[aria-label="Update priority"]').should("be.visible");
    cy.get('button[aria-label="Export tasks"]').should("be.visible");
    cy.get('button[aria-label="Delete selected tasks"]').should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-002 opens create and import actions from Create menu", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-create-menu-trigger"]').click();
    cy.get('[data-testid="wishlist-create-menu-entry"]').should("be.visible");
    cy.get('[data-testid="wishlist-create-menu-import"]').should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-004 shows dedicated New action with adjacent Create menu", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-new-action"]')
      .should("be.visible")
      .and("contain", "New");
    cy.get('[data-testid="wishlist-create-menu-trigger"]')
      .should("be.visible")
      .and("contain", "Create");
  });

  it("UI-SCREEN-WISHLIST-005 supports inline collection create and auto-select", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-inline-add-new"]').click();
    cy.get('[data-testid="wishlist-inline-new-name"]').type("Wishlist Inline Alpha");
    cy.get('[data-testid="wishlist-inline-save"]').click();
    cy.get('[data-testid="wishlist-inline-picker-selected"]').should(
      "contain",
      "Wishlist Inline Alpha"
    );
  });

  it("UI-SCREEN-WISHLIST-007 renders wishlist collection semantics instead of task seed rows", () => {
    signInToWishlist();

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.contains("th", "Item ID").should("be.visible");
    cy.contains("th", "Title").should("be.visible");
    cy.contains("th", "Watch Status").should("be.visible");
    cy.contains("th", "Target Priority").should("be.visible");
    cy.contains("th", "Task").should("not.exist");
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("be.visible");
    cy.contains("item-collector-1").should("be.visible");
    cy.contains("Watching").should("be.visible");
    cy.contains("Below target").should("be.visible");
    cy.contains(/TASK-\d+/).should("not.exist");
    cy.contains("Backlog").should("not.exist");
  });
});
