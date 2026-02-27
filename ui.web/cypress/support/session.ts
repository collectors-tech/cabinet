export function ensureAuthenticatedWorkspace(profileId: string) {
  cy.window().then((win) => {
    win.localStorage.setItem(`cabinet.workspace.${profileId}`, "1");
  });
}
