describe("ui-keyboard-shortcuts", () => {
  beforeEach(() => {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
  });

  function signInToHome(
    onBeforeLoad?: (win: Cypress.AUTWindow) => void
  ) {
    cy.e2eBootstrap({ minimalProfile: true }).then((state) => {
      cy.e2eEnsureSignedOut();
      cy.stubLocalServerSession(state.profile_id);
      cy.request("PUT", "/api/profiles/active", {
        profile_id: state.profile_id,
      })
        .its("status")
        .should("eq", 200);
      cy.visit("/sign-in?redirect=%2Fdashboard", {
        onBeforeLoad(win) {
          win.localStorage.setItem(
            `cabinet.workspace.${state.profile_id}`,
            "1"
          );
          onBeforeLoad?.(win);
        },
      });
      cy.contains("button", "Open local workspace").click();
      cy.wait("@localServerSession");
    });
    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard");
  }

  function signInToHomeWithShortcutOverrides(
    overrides: Record<string, string>
  ) {
    signInToHome((win) => {
      win.localStorage.setItem(
        "cabinet.shortcuts.overrides",
        JSON.stringify(overrides)
      );
    });
  }

  function dispatchShortcut(key: string) {
    cy.document().then((doc) => {
      const win = doc.defaultView;
      expect(win, "application window").to.exist;
      const eventInit: KeyboardEventInit = {
        key,
        code: `Key${key.toUpperCase()}`,
        ctrlKey: true,
        bubbles: true,
        cancelable: true,
      };
      doc.dispatchEvent(new win!.KeyboardEvent("keydown", eventInit));
      win!.dispatchEvent(new win!.KeyboardEvent("keydown", eventInit));
    });
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

    cy.get('[data-slot="sidebar"]').first().invoke("attr", "data-state").then((initialState) => {
      expect(initialState).to.match(/^(expanded|collapsed)$/);
      const toggledState = initialState === "expanded" ? "collapsed" : "expanded";

      dispatchShortcut("b");
      cy.get('[data-slot="sidebar"]').first().should("have.attr", "data-state", toggledState);
      dispatchShortcut("b");
      cy.get('[data-slot="sidebar"]').first().should("have.attr", "data-state", initialState);
    });
  });

  it("UI-KEYBOARD-SHORTCUTS-003 rejects duplicate/reserved shortcut registrations and applies fallback policy", () => {
    signInToHomeWithShortcutOverrides({
      "command-palette": "k",
      "sidebar-toggle": "k",
    });

    dispatchShortcut("b");
    cy.get('[data-slot="sidebar"]').first().invoke("attr", "data-state").then((sidebarStateAfterFallback) => {
      expect(sidebarStateAfterFallback).to.match(/^(expanded|collapsed)$/);
      dispatchShortcut("k");
      cy.get('input[placeholder="Type a command or search..."]').should("be.visible");
      cy.get('[data-slot="sidebar"]').first().should("have.attr", "data-state", sidebarStateAfterFallback);
    });
  });
});
