## Purpose
Define the in-repo AntFarm workflow portability contract so Cabinet workflow execution does not depend on machine-specific or shared external paths.

## Requirements
### Requirement ANTFARM-WORKFLOW-001: Cabinet workflow definition MUST remain self-contained in repository paths
The AntFarm workflow configuration for Cabinet MUST reference only project-local agent profile paths.

#### Scenario: Workflow file references only local agent paths
- **GIVEN** the runtime loads `.antfarm/workflows/cabinet/workflow.yml`
- **WHEN** agent workspace references are parsed
- **THEN** each agent workspace MUST resolve inside `.antfarm/workflows/cabinet/agents/*`
- **AND** workflow content MUST NOT include shared-agent or absolute filesystem path references

### Requirement ANTFARM-WORKFLOW-002: Metadata MUST identify the Cabinet workflow deterministically
The workflow metadata file MUST declare identity values that match the Cabinet workflow bundle.

#### Scenario: Metadata identity contract
- **GIVEN** `.antfarm/workflows/cabinet/metadata.json` is present
- **WHEN** metadata is loaded
- **THEN** `workflowId` MUST equal `cabinet`
- **AND** `source` MUST equal `bundled:cabinet`

### Requirement ANTFARM-WORKFLOW-003: Local agent profile set MUST be complete for all workflow roles
Every role declared in the workflow MUST have local profile files required for execution.

#### Scenario: Required profile files exist for each agent role
- **GIVEN** workflow roles `planner`, `setup`, `developer`, `verifier`, `tester`, and `reviewer`
- **WHEN** role profile files are resolved
- **THEN** each role directory MUST include `AGENTS.md`, `SOUL.md`, and `IDENTITY.md`
