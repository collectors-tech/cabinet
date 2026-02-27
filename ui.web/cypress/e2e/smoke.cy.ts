describe("Cabinet smoke", () => {
  it("boots runtime and supports profile + collection API journey", () => {
    cy.visit("/");
    cy.get("#root").should("exist");

    cy.request("/healthz").its("status").should("eq", 200);

    cy.fixture("test-user.json").then((fx) => {
      const uniquePartNumber = `${fx.item.part_number}-${Date.now()}`;
      cy.request("POST", "/api/profiles", { name: fx.profileName }).then((createProfile) => {
        expect(createProfile.status).to.eq(201);
        const profileId = createProfile.body.id;
        expect(profileId).to.be.a("string").and.not.be.empty;

        cy.request("PUT", "/api/profiles/active", { profile_id: profileId }).its("status").should("eq", 200);
        cy.request("POST", "/api/items", { ...fx.item, part_number: uniquePartNumber }).its("status").should("eq", 201);

        cy.request("GET", "/api/items").then((itemsResp) => {
          expect(itemsResp.status).to.eq(200);
          const items = itemsResp.body.items as Array<{ part_number: string }>;
          expect(items.some((i) => i.part_number === uniquePartNumber)).to.eq(true);
        });
      });
    });
  });
});
