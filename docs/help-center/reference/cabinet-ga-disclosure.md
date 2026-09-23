# Cabinet 1.0 capabilities and limitations

Release channel: ga
Release version: 1.0.0

This article is generated from the governed release disclosure source so the product UI and release notes stay aligned.

## Capability and limitation statements

- Windows portable package (supported): Cabinet 1.0 is distributed as an unsigned Windows portable ZIP, not an installer.
- Provider readiness (limited): Supported collector paths are identified in the packaged guidance; protected or user-present provider journeys remain action-required.
- Data ownership, export and recovery (supported): User data can use export, backup, and restore from local Cabinet surfaces and is not trapped behind a paid gate.

## Support and recovery pointers

- Back up from Settings Storage before replacing a portable build or reusing an existing data directory.
- Use Settings Operations for JSON/CSV export, logs, and recovery checks when validating a release.
- Treat release notes, package checksums, and the exact source commit as the support identity for a release.
