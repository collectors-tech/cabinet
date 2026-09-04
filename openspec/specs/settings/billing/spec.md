## Purpose
Define Settings Billing screen behavior for visible entitlement placeholders and disabled billing actions until cloud billing controls are enabled.

## Requirements
### Requirement UI-SCREEN-SETTINGS-BILLING-001: Billing screen SHALL render entitlement and founding-license state without mutating billing
The authenticated `/settings/billing` route SHALL render deterministic plan and license status information, expose the billing portal affordance as disabled while it is coming soon, support signed founding-license import, and avoid presenting an enabled billing mutation action.

#### Scenario: Billing static state is inspectable and non-mutating
- **GIVEN** an authenticated user opens `/settings/billing`
- **WHEN** the billing section renders
- **THEN** the screen MUST show `Plan` and `License Status` information cards
- **AND** the plan card MUST show the current Free/Plus/Pro plan, entitlement source, and tier guidance aligned with the canonical beta matrix
- **AND** the license card MUST show signed founding-license status and a signed license import control
- **AND** the `Open Billing Portal (Coming soon)` control MUST be disabled
- **AND** the screen MUST NOT expose an enabled billing portal action until a billing portal integration is available

## Traceability
| Use case | Contract | Verification |
| --- | --- | --- |
| UC-SET-BILL-01 | Billing visible entitlement and founding-license state | `ui.web/cypress/e2e/settings/billing/spec.cy.ts` `UI-SCREEN-SETTINGS-BILLING-001 renders disabled static billing state without portal mutation` |
