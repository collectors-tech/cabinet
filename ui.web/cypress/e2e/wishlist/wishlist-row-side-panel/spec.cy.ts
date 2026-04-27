describe("wishlist-row-side-panel", () => {
  function signInToWishlistWithRows() {
    let wishlistEntries = [
      {
        id: "wish-panel-1",
        item_id: "item-panel-1",
        priority: "medium",
        below_target_now: false,
        notes: "First panel note",
        target_price: 40,
      },
      {
        id: "wish-panel-2",
        item_id: "item-panel-2",
        priority: "high",
        below_target_now: false,
        notes: "Second panel note",
        target_price: 80,
      },
    ];
    let wishlistItems = [
      {
        id: "item-panel-1",
        title: "Panel Wishlist Alpha",
        part_number: "PANEL-001",
        status: "wishlist",
        category: "Cards",
        priority: "medium",
      },
      {
        id: "item-panel-2",
        title: "Panel Wishlist Beta",
        part_number: "PANEL-002",
        status: "wishlist",
        category: "Comics",
        priority: "high",
      },
    ];

    cy.intercept("GET", "/api/wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } });
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } });
    }).as("catalogItems");
    cy.intercept("PUT", "/api/items/item-panel-2", (req) => {
      wishlistItems = wishlistItems.map((item) =>
        item.id === "item-panel-2"
          ? {
              ...item,
              title: req.body.title,
              part_number: req.body.part_number,
              category: req.body.category,
            }
          : item
      );
      req.reply({
        statusCode: 200,
        body: wishlistItems.find((item) => item.id === "item-panel-2"),
      });
    }).as("updateWishlistItem");
    cy.intercept("PUT", "/api/wishlist", (req) => {
      wishlistEntries = wishlistEntries.map((entry) =>
        entry.id === req.body.id
          ? {
              ...entry,
              priority: req.body.priority,
              notes: req.body.notes,
              target_price: req.body.target_price,
            }
          : entry
      );
      req.reply({ statusCode: 204, body: "" });
    }).as("updateWishlistEntry");
    cy.intercept("GET", "/api/profiles/*/settings").as("profileSettings");

    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: "/wishlist/",
      });
    });
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.wait("@profileSettings");
    cy.wait("@profileSettings");
  }

  it("opens a right-side edit panel on double click and navigates visible records", () => {
    signInToWishlistWithRows();

    cy.contains("tr", "Panel Wishlist Alpha").click();
    cy.get('[data-testid="wishlist-edit-panel"]').should("not.exist");
    cy.get('[data-testid="row-edit-modal"]').should("not.exist");

    cy.contains("tr", "Panel Wishlist Alpha").dblclick();
    cy.get('[data-testid="wishlist-edit-panel"]')
      .should("be.visible")
      .should("have.attr", "data-side", "right");
    cy.get('input[name="title"]').should("have.value", "Panel Wishlist Alpha");

    cy.get('[data-testid="wishlist-edit-next"]').click();
    cy.get('input[name="title"]').should("have.value", "Panel Wishlist Beta");
    cy.get('[data-testid="wishlist-edit-previous"]').click();
    cy.get('input[name="title"]').should("have.value", "Panel Wishlist Alpha");
    cy.get('[data-testid="wishlist-edit-next"]').click();

    cy.get('input[name="title"]').clear().type("Panel Wishlist Beta Updated");
    cy.get('textarea[name="notes"]').clear().type("Updated from side panel");
    cy.contains("button", "Save changes").click();

    cy.wait("@updateWishlistItem");
    cy.wait("@updateWishlistEntry");
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.contains("Panel Wishlist Beta Updated").should("be.visible");
    cy.contains("Updated from side panel").should("be.visible");
  });
});
