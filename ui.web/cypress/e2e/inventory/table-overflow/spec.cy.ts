describe("inventory table overflow", () => {
  function signIn() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: "/inventory/",
      });
    });
  }

  it("UI-SCREEN-INVENTORY-ITEMS-008 truncates long row values instead of overlapping columns", () => {
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-long-row",
            part_number:
              "URL-BONZASLOTCARS-002-MATCHBOX-HOT-WHEELS-MOOISYRQ",
            title:
              "Bonza Matchbox Hot Wheels Mooi Syrq Long Showcase Title",
            status: "active",
            category: "Slot Cars",
          },
        ],
      },
    }).as("itemsLongRow");

    signIn();
    cy.wait("@itemsLongRow");

    cy.get('[data-testid="inventory-item-row-item-long-row"]')
      .should("exist")
      .scrollIntoView()
      .within(() => {
        cy.get("td")
          .eq(1)
          .then(($cell) => {
            expect($cell[0].scrollWidth, "part number cell overflow").to.be.at
              .most($cell[0].clientWidth + 1);
          });
        cy.get("td")
          .eq(2)
          .then(($cell) => {
            expect($cell[0].scrollWidth, "title cell overflow").to.be.at.most(
              $cell[0].clientWidth + 1
            );
          });
        cy.get("td")
          .eq(1)
          .find("[data-testid='inventory-row-part-number']")
          .should("have.css", "text-overflow", "ellipsis");
      });
  });
});
