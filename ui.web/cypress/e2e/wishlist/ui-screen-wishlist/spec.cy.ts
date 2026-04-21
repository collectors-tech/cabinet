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
    cy.intercept("GET", "/api/items?status=wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-collector-1",
            title: "AFX Mega-G+ Camaro Wildfire",
            part_number: "22073",
            status: "wishlist",
            category: "Slot Cars",
            priority: "medium",
          },
          {
            id: "item-collector-2",
            title: "F1 Silverline",
            part_number: "F1002",
            status: "wishlist",
            category: "Formula",
            priority: "high",
          },
        ],
      },
    }).as("catalogItems");
  }

  function signInToWishlist(options?: { skipStub?: boolean }) {
    if (!options?.skipStub) {
      stubWishlistData();
    }
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

  it("UI-SCREEN-WISHLIST-005 keeps inline create open and shows validation on blank save", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-inline-add-new"]').click();
    cy.get('[data-testid="wishlist-inline-save"]').click();
    cy.get('[data-testid="wishlist-inline-new-name"]').should(
      "have.attr",
      "aria-invalid",
      "true"
    );
    cy.get('[data-testid="wishlist-inline-validation"]')
      .should("be.visible")
      .and("contain", "Collection name is required.");
    cy.get('[data-testid="wishlist-inline-new-name"]').should("be.visible");
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

  it("UI-SCREEN-WISHLIST-006 moves acquired item into inventory and keeps it out of wishlist after refresh", () => {
    let wishlistEntries = [
      {
        id: "wish-1",
        item_id: "item-collector-1",
        priority: "high",
        below_target_now: false,
      },
    ];
    let wishlistItems = [
      {
        id: "item-collector-1",
        title: "AFX Mega-G+ Camaro Wildfire",
        part_number: "22073",
        status: "wishlist",
        category: "Slot Cars",
        priority: "high",
      },
    ];
    const inventoryItems = [
      {
        id: "item-collector-1",
        title: "AFX Mega-G+ Camaro Wildfire",
        part_number: "22073",
        status: "active",
        category: "Slot Cars",
        priority: "high",
      },
    ];

    cy.intercept("GET", "/api/wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } });
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } });
    }).as("catalogItems");
    cy.intercept("DELETE", "/api/wishlist?id=wish-1", (req) => {
      wishlistEntries = [];
      wishlistItems = [];
      req.reply({ statusCode: 204, body: "" });
    }).as("moveWishlistItemToOwned");
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items: inventoryItems },
    }).as("inventoryItems");

    signInToWishlist({ skipStub: true });

    cy.contains("tr", "AFX Mega-G+ Camaro Wildfire").within(() => {
      cy.get('[data-testid="task-row-actions-trigger"]').click();
    });
    cy.get('[data-testid="wishlist-mark-owned-action"]').click();

    cy.wait("@moveWishlistItemToOwned");
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("not.exist");

    cy.reload();
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("not.exist");

    cy.visit("/inventory/");
    cy.wait("@inventoryItems");
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/inventory\/?$/);
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-008 surfaces planning summary and persists selected focus across refresh and route return", () => {
    cy.intercept("GET", "/api/wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "wish-1",
            item_id: "item-collector-1",
            priority: "medium",
            below_target_now: false,
            notes: "Wait for convention restock",
          },
          {
            id: "wish-2",
            item_id: "item-collector-2",
            priority: "high",
            below_target_now: true,
            notes: "Buy if price drops again",
          },
          {
            id: "wish-3",
            item_id: "item-collector-3",
            priority: "critical",
            below_target_now: false,
            notes: "Need before championship season",
          },
        ],
      },
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-collector-1",
            title: "AFX Mega-G+ Camaro Wildfire",
            part_number: "22073",
            status: "wishlist",
            category: "Slot Cars",
            priority: "medium",
          },
          {
            id: "item-collector-2",
            title: "F1 Silverline",
            part_number: "F1002",
            status: "wishlist",
            category: "Formula",
            priority: "high",
          },
          {
            id: "item-collector-3",
            title: "Team Transport Hauler",
            part_number: "TT-88",
            status: "wishlist",
            category: "Haulers",
            priority: "critical",
          },
        ],
      },
    }).as("catalogItems");

    signInToWishlist({ skipStub: true });

    cy.get('[data-testid="wishlist-planning-summary"]').should("be.visible");
    cy.get('[data-testid="wishlist-planning-focus-all"]').should(
      "contain",
      "3"
    );
    cy.get('[data-testid="wishlist-planning-focus-high-priority"]').should(
      "contain",
      "2"
    );
    cy.get('[data-testid="wishlist-planning-focus-below-target"]').should(
      "contain",
      "1"
    );
    cy.get('[data-testid="wishlist-planning-focus-watchlist"]').should(
      "contain",
      "1"
    );

    cy.get('[data-testid="wishlist-planning-focus-below-target"]').click();
    cy.window()
      .its("localStorage")
      .invoke("getItem", "cabinet.wishlistPlanningFocus")
      .should("eq", "below-target");
    cy.get('[data-testid="wishlist-planning-focus-below-target"]').should(
      "have.attr",
      "aria-pressed",
      "true"
    );
    cy.contains("F1 Silverline").should("be.visible");
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("not.exist");
    cy.contains("Team Transport Hauler").should("not.exist");

    cy.reload();
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.get('[data-testid="wishlist-planning-focus-below-target"]').should(
      "have.attr",
      "aria-pressed",
      "true"
    );
    cy.contains("F1 Silverline").should("be.visible");
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("not.exist");

    cy.visit("/inventory/");
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/inventory\/?$/);
    cy.visit("/wishlist/");
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.get('[data-testid="wishlist-planning-focus-below-target"]').should(
      "have.attr",
      "aria-pressed",
      "true"
    );
    cy.contains("F1 Silverline").should("be.visible");
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("not.exist");
  });
});
