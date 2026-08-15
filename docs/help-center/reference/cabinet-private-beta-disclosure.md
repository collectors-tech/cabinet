# Cabinet 0.1 private beta capabilities and limitations

Release channel: private-beta
Release version: 0.1.0-beta.6

This article is generated from the governed release disclosure source so the product UI and release notes stay aligned.

## Capability and limitation statements

- Windows portable package (supported): Cabinet 0.1 is distributed as a Windows portable ZIP for private beta validation, not an installer.
- Signing and installer limit (limited): The beta artefact is unsigned and no signed installer is claimed until signed installer evidence exists.
- Browser Companion package targets (supported): The Cabinet Browser Companion beta targets Chrome and Edge with developer-mode unpacked installation from exact package ZIPs.
- Browser Companion updates (limited): There are no silent automatic browser-store updates in this beta; replace the exact package when a newer validated candidate is issued.
- Voglers provider readiness (supported): Voglers is treated as a direct public storefront provider path for beta candidate Market Watch evidence.
- Hobbytech provider readiness (packaged unproven): Hobbytech is packaged-unproven until exact candidate evidence records the packaged provider journey.
- Frontline provider readiness (browser assisted): Frontline is browser-assisted and action-required when protected pages or user-present browser interaction are needed.
- Bonza provider readiness (browser assisted): Bonza is browser-assisted and action-required when protected pages or user-present browser interaction are needed.
- Authentication modes (supported): The candidate supports local account mode and ZITADEL mode where the runtime is configured for that authority.
- Assistant and Agent features (preview): Assistant and Agent features are preview and optional; unsafe or setup-dependent actions remain gated by confirmation and profile context.
- Telegram live-channel limitation (action required): Telegram live acceptance is not complete for this candidate until sender, chat, bot token, and webhook proof are recorded.
- Post-beta exclusions (excluded): Metadata Studio breadth, public identity and trust/P2P features, broad provider expansion, Telegram live acceptance, and eBay seller workflows are post-beta exclusions unless separate evidence says otherwise.
- Data ownership, export and recovery (supported): User data can use export, backup, and restore from local Cabinet surfaces and is not trapped behind a paid gate.

## Support and recovery pointers

- Back up from Settings Storage before replacing a portable build or reusing an existing data directory.
- Use Settings Operations for JSON/CSV export, logs, and recovery checks when validating a candidate.
- Treat release notes, package checksums, and the exact source commit as the support identity for a beta candidate.
