describe("ui-foundation-auth-menus-shortcuts", () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
  });

  function signInToHome() {
    cy.e2eBootstrap({ minimalProfile: true }).then((state) => {
      cy.e2eEnsureSignedOut();
      cy.stubLocalServerSession(state.profile_id);
      cy.useBootstrappedProfile(state.profile_id, state.profile_name, {
        path: "/dashboard",
      });
      cy.wait("@localServerSession");
    });
    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard");
  }

  function openProfileMenu() {
    cy.get('[data-testid="profile-dropdown-trigger"]')
      .filter(":visible")
      .first()
      .click();
  }

  it("UI-FOUNDATION-AUTH-MENUS-SHORTCUTS-001 renders identity-backed platform-aware shortcut labels", () => {
    signInToHome();

    cy.window().then((win) => {
      const isMac = /Mac|iPhone|iPad|iPod/i.test(win.navigator.platform);
      const expectedAccount = isMac ? "⇧⌘P" : "Ctrl+Shift+P";
      const expectedNotifications = isMac ? "⌘," : "Ctrl+,";
      const expectedSignOut = isMac ? "⇧⌘Q" : "Ctrl+Shift+Q";

      openProfileMenu();
      cy.get('[data-testid="profile-shortcut-profile"]').should(
        "contain",
        expectedAccount
      );
      cy.get('[data-testid="profile-shortcut-settings"]').should(
        "contain",
        expectedNotifications
      );
      cy.get('[data-testid="profile-shortcut-signout"]').should(
        "contain",
        expectedSignOut
      );
    });
  });

  it("UI-FOUNDATION-AUTH-MENUS-SHORTCUTS-002 excludes unsupported template actions", () => {
    signInToHome();
    openProfileMenu();
    cy.contains(/new team/i).should("not.exist");
    cy.contains(/billing/i).should("not.exist");
  });

  it("UI-FOUNDATION-AUTH-MENUS-SHORTCUTS-003 excludes sidebar upsell rows", () => {
    signInToHome();
    cy.contains(/upgrade to pro/i).should("not.exist");
  });
});
