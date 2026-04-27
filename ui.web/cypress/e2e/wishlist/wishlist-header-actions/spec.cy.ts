describe("wishlist-header-actions", () => {
  function openWishlist() {
    cy.intercept("GET", "/api/wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "wish-header-1",
            item_id: "item-header-1",
            priority: "medium",
            below_target_now: false,
          },
        ],
      },
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-header-1",
            title: "Header Action Wishlist Item",
            part_number: "HDR-001",
            status: "wishlist",
            category: "Slot Cars",
            priority: "medium",
          },
        ],
      },
    }).as("catalogItems");
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
  }

  it("shows direct compact wishlist header actions without create menu", () => {
    openWishlist();

    cy.get('[data-testid="wishlist-global-header-actions"]').within(() => {
      cy.get('[data-testid="wishlist-new-action"]')
        .should("be.visible")
        .and("have.attr", "aria-label", "New wishlist item")
        .and("have.attr", "title", "New wishlist item")
        .and("not.contain.text", "New");
      cy.get('[data-testid="wishlist-create-collection-action"]')
        .should("be.visible")
        .and("have.attr", "aria-label", "Create collection")
        .and("have.attr", "title", "Create collection")
        .and("not.contain.text", "Create");
      cy.get('[data-testid="wishlist-import-action"]')
        .should("be.visible")
        .and("have.attr", "aria-label", "Import wishlist entries")
        .and("have.attr", "title", "Import wishlist entries")
        .and("not.contain.text", "Import");
      cy.get('[data-testid="wishlist-create-menu-trigger"]').should("not.exist");
    });
  });

  it("opens wishlist create, collection create, and import from direct buttons", () => {
    openWishlist();

    cy.get('[data-testid="wishlist-new-action"]').click();
    cy.contains("Create Wishlist Entry").should("be.visible");
    cy.contains("button", "Close").click();

    cy.get('[data-testid="wishlist-create-collection-action"]').click();
    cy.get('[data-testid="wishlist-table-new-collection-name"]').should(
      "be.visible"
    );

    cy.get('[data-testid="wishlist-import-action"]').click();
    cy.contains("Import Wishlist Entries").should("be.visible");
  });
});
