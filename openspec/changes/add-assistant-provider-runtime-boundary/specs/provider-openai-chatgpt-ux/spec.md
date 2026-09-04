## MODIFIED Requirements

### Requirement: OpenAI / ChatGPT integration SHALL expose capability scoping for assistant workflows
Cabinet MUST define which assistant capabilities use OpenAI / ChatGPT integration (for example chat help, photo analysis, tagging/classification, generation) and expose that scope to operators/users where appropriate.

#### Scenario: Run governed assistant turn through configured OpenAI provider
- **GIVEN** the active profile has OpenAI / ChatGPT configured through the integration registry with a verified active method and a supported default assistant model
- **WHEN** governed Chat asks the assistant-provider runtime for a conversational turn
- **THEN** Cabinet SHALL resolve provider, credential, active method, model, and readiness from the active profile's integration registry/instance and secret boundary
- **AND** the OpenAI adapter SHALL return provider text and non-secret metadata through a provider-neutral assistant turn response
- **AND** the provider SHALL NOT receive Cabinet skill handles, database handles, filesystem handles, or app-control tool execution authority
- **AND** Cabinet's governed planner SHALL remain responsible for any skill selection, preview, confirmation, or apply behavior after provider text is returned

#### Scenario: Run deterministic fake assistant provider in tests
- **GIVEN** a test config selects the deterministic fake assistant provider
- **WHEN** governed Chat requests an assistant turn
- **THEN** Cabinet SHALL complete the turn without network calls
- **AND** the response SHALL use the same provider-neutral assistant turn shape as OpenAI-backed turns
- **AND** fake-provider output SHALL be explicit test evidence and SHALL NOT mark OpenAI, Anthropic, Google, or any live provider as ready

#### Scenario: Report assistant-provider runtime setup and failure states
- **GIVEN** governed Chat needs an assistant-provider turn
- **WHEN** the active profile is missing credentials, has an unhealthy provider, selects an unsupported model, times out, is cancelled, hits rate limits, or receives a transport/provider error
- **THEN** Cabinet SHALL return deterministic actionable setup or failure guidance
- **AND** the guidance SHALL use stable error classes for missing credentials, unhealthy provider, unsupported model, timeout, cancellation, rate limit, transport failure, and provider failure
- **AND** logs, errors, workflow evidence, health diagnostics, and user-facing responses SHALL omit credential values, raw provider payloads, arbitrary local paths, and unrelated profile context

#### Scenario: Keep placeholder assistant providers unavailable
- **GIVEN** the provider registry exposes Anthropic, Google, or another assistant placeholder without an implemented runtime adapter
- **WHEN** governed Chat asks for a provider-backed assistant turn
- **THEN** Cabinet SHALL report the provider as unavailable with setup/adapter guidance
- **AND** Cabinet SHALL NOT route the turn through OpenAI as a hidden fallback for that provider
- **AND** Cabinet SHALL NOT claim live readiness until a separate implemented adapter and validation evidence exist
