## MODIFIED Requirements

### Requirement: Dashboard home English copy SHALL resolve current shell translations
Cabinet SHALL render the dashboard home shell using resolved English dashboard translations for the current actionable, loading, empty, and error states.

#### Scenario: Dashboard home renders resolved English shell copy
- **GIVEN** the dashboard home screen renders in the English locale
- **WHEN** it shows actionable, loading, empty, recent-items, or unavailable states
- **THEN** it SHALL render resolved English copy rather than raw `dashboard.*` translation keys