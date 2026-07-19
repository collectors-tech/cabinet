# wsl-runner-pool Specification

## Purpose

Define Cabinet's optional repository-owned mechanism for provisioning isolated,
bounded WSL 2 GitHub Actions runner capacity.

## ADDED Requirements

### Requirement: CABINET-CI-RUNNER-001 Every runner SHALL have an isolated WSL boundary

The installer SHALL create one dedicated WSL distribution, Linux account,
runner work directory, Docker daemon, and VHD location for every requested pool
member.

#### Scenario: Add three runner pool members

- **GIVEN** no Cabinet runner distributions exist
- **WHEN** an operator requests `RunnerCount 3`
- **THEN** the installer SHALL plan `cabinet`, `cabinet-02`, and `cabinet-03`
- **AND** each member SHALL have a distinct account, install location, runner
  name, and custom runner label

### Requirement: CABINET-CI-RUNNER-002 Pool reconciliation SHALL be idempotent

The installer SHALL treat the requested runner count as desired state and SHALL
leave an existing member unchanged when an Actions runner systemd service is
already configured.

#### Scenario: Scale an existing pool

- **GIVEN** the unnumbered Cabinet runner is configured
- **WHEN** an operator requests `RunnerCount 3`
- **THEN** member 1 SHALL remain unchanged
- **AND** only missing or unconfigured members SHALL be provisioned

### Requirement: CABINET-CI-RUNNER-003 Provisioning SHALL install the Cabinet CI toolchain securely

The installer SHALL install Docker with `overlay2`, Node.js 22, the Go version
declared by `go.mod`, PowerShell, and official GitHub runner dependencies. It
SHALL keep the Linux password and registration token out of process command
lines, persistent service environment, and the repository.

#### Scenario: Register multiple missing members

- **GIVEN** GitHub CLI is authenticated with runner-administration access
- **WHEN** multiple members require configuration
- **THEN** a fresh repository registration token SHALL be requested for each
  member
- **AND** a caller-supplied single-use token SHALL be rejected for multi-member
  provisioning

### Requirement: CABINET-CI-RUNNER-004 Storage growth SHALL be bounded and reaped only while idle

The installer SHALL configure bounded Docker logs and an idle-aware cleanup
timer that removes unused Docker resources, completed runner workspaces,
package caches, and old journal data, then periodically trims free ext4 blocks.

#### Scenario: Cleanup encounters an active job

- **GIVEN** a runner job marker or `Runner.Worker` process is active
- **WHEN** scheduled cleanup runs
- **THEN** cleanup SHALL exit without pruning the job workspace or Docker state

### Requirement: CABINET-CI-RUNNER-005 Host capacity and lifecycle SHALL be explicit

The installer SHALL enforce a VHD ceiling and aggregate Windows free-space
headroom for missing members, support a non-mutating `WhatIf` plan, and
configure a hidden per-distro current-user logon task unless explicitly skipped.

#### Scenario: Requested pool exceeds safe host capacity

- **GIVEN** the target Windows drive cannot hold every missing VHD ceiling plus
  15 GB headroom
- **WHEN** the operator requests the pool without a low-space override
- **THEN** provisioning SHALL stop before creating a distribution
- **AND** the error SHALL explain how to change the location, ceiling, count, or
  explicit override
