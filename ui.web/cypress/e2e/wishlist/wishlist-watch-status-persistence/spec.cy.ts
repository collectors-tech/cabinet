describe("wishlist-watch-status-persistence", () => {
  it("persists manual watch status changes across refresh", () => {
    cy.intercept("GET", "/api/wishlist").as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist").as("catalogItems");
    cy.intercept("GET", "/api/profiles/*/settings").as("profileSettings");

    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.request("PUT", "/api/profiles/active", { profile_id })
        .its("status")
        .should("eq", 200);

      cy.request("POST", "/api/items", {
        title: "Wishlist Manual Status Seed",
        part_number: "WISH-LIVE-001",
        category: "Slot Cars",
        priority: "medium",
      }).then((itemResp) => {
        expect(itemResp.status).to.eq(201);
        cy.request("POST", "/api/wishlist", {
          item_id: itemResp.body.id,
          target_price: 42,
          priority: "medium",
          notes: "Seeded for status persistence",
          below_target_now: false,
          highlight_hit: false,
        })
          .its("status")
          .should("eq", 201);
      });

      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: "/wishlist/",
      });
    });

    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.wait("@profileSettings");
    cy.wait("@profileSettings");

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.contains("Wishlist Manual Status Seed").should("be.visible");
    cy.get('button[aria-label="Select all"]').click();
    cy.get('button[aria-label="Update status"]').click();
    cy.contains('[role="menuitem"]', "Below target").click({ force: true });

    cy.contains("Wishlist Manual Status Seed")
      .closest("tr")
      .should("contain", "Below target");

    cy.reload();
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.contains("Wishlist Manual Status Seed")
      .closest("tr")
      .should("contain", "Below target");
  });
});
