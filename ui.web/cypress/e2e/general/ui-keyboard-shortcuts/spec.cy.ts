describe("ui-keyboard-shortcuts", () => {
  function signInToHome() {
    cy.clearCookies();
    cy.clearLocalStorage();
    cy.visit("/sign-in?redirect=%2F");
    cy.get('input[name="email"]').clear().type("e2e-shortcuts@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard");
  }

  function signInToHomeWithShortcutOverrides(
    overrides: Record<string, string>
  ) {
    cy.clearCookies();
    cy.clearLocalStorage();
    cy.visit("/sign-in?redirect=%2F", {
      onBeforeLoad(win) {
        win.localStorage.setItem(
          "cabinet.shortcuts.overrides",
          JSON.stringify(overrides)
        );
      },
    });
    cy.get('input[name="email"]').clear().type("e2e-shortcuts@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard");
  }

  it("UI-KEYBOARD-SHORTCUTS-001 renders platform-aware shortcut labels in profile menu", () => {
    signInToHome();

    cy.window().then((win) => {
      const isMac = /Mac|iPhone|iPad|iPod/i.test(win.navigator.platform);
      const expectedProfile = isMac ? "⇧⌘P" : "Ctrl+Shift+P";
      const expectedSettings = isMac ? "⌘," : "Ctrl+,";
      const expectedSignOut = isMac ? "⇧⌘Q" : "Ctrl+Shift+Q";

      cy.get('[data-testid="profile-dropdown-trigger"]')
        .filter(":visible")
        .first()
        .click();
      cy.get('[data-testid="profile-shortcut-profile"]').should(
        "contain",
        expectedProfile
      );
      cy.get('[data-testid="profile-shortcut-settings"]').should(
        "contain",
        expectedSettings
      );
      cy.get('[data-testid="profile-shortcut-signout"]').should(
        "contain",
        expectedSignOut
      );
    });
  });

  it("UI-KEYBOARD-SHORTCUTS-002 toggles sidebar with global Ctrl/Cmd+B shortcut", () => {
    signInToHome();

    cy.get('[data-slot="sidebar"]').first().should("have.attr", "data-state", "expanded");
    cy.get("body").click(0, 0).type("{ctrl}b");
    cy.get('[data-slot="sidebar"]').first().should("have.attr", "data-state", "collapsed");
    cy.get("body").click(0, 0).type("{ctrl}b");
    cy.get('[data-slot="sidebar"]').first().should("have.attr", "data-state", "expanded");
  });

  it("UI-KEYBOARD-SHORTCUTS-003 rejects duplicate/reserved shortcut registrations and applies fallback policy", () => {
    signInToHomeWithShortcutOverrides({
      "command-palette": "k",
      "sidebar-toggle": "k",
    });

    cy.get('[data-slot="sidebar"]').first().should("have.attr", "data-state", "expanded");
    cy.get("body").click(0, 0).type("{ctrl}b");
    cy.get('[data-slot="sidebar"]').first().should("have.attr", "data-state", "collapsed");
    cy.get("body").click(0, 0).type("{ctrl}k");
    cy.get('input[placeholder="Type a command or search..."]').should("be.visible");
    cy.get('[data-slot="sidebar"]').first().should("have.attr", "data-state", "collapsed");
  });
});
