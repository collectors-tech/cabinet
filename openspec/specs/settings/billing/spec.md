## Purpose
Define Settings Billing screen behavior for visible entitlement placeholders and disabled billing actions until cloud billing controls are enabled.

## Requirements
### Requirement UI-SCREEN-SETTINGS-BILLING-001: Billing screen SHALL render static entitlement state without mutating billing
The authenticated `/settings/billing` route SHALL render deterministic plan and license status information, expose the billing portal affordance as disabled while it is coming soon, and avoid presenting an enabled billing mutation action.

#### Scenario: Billing static state is inspectable and non-mutating
- **GIVEN** an authenticated user opens `/settings/billing`
- **WHEN** the billing section renders
- **THEN** the screen MUST show `Plan` and `License Status` information cards
- **AND** the `Open Billing Portal (Coming soon)` control MUST be disabled
- **AND** the screen MUST NOT expose an enabled billing portal action until a billing portal integration is available

## Traceability
| Use case | Contract | Verification |
| --- | --- | --- |
| UC-SET-BILL-01 | Billing static entitlement placeholders | `ui.web/cypress/e2e/settings/billing/spec.cy.ts` `UI-SCREEN-SETTINGS-BILLING-001 renders disabled static billing state without portal mutation` |
