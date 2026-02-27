import "./commands";

beforeEach(() => {
  Cypress.on("uncaught:exception", () => false);
});
