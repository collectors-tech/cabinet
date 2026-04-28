# Cabinet Console Output Standard

Cabinet command output should feel like part of the product: recognizable,
calm, structured, and actionable. Scripts should avoid bare one-off
`Write-Host` lines when a shared Cabinet formatter can provide consistent
sections, status states, key/value facts, and hints.

## Status states

Use the shared PowerShell helper at `scripts/lib/cabinet-console.ps1` for
high-traffic scripts:

- `Write-CabinetBanner` identifies the command and purpose once at startup.
- `Write-CabinetSection` separates phases such as build, validation, and launch.
- `Write-CabinetStatus` reports `run`, `ok`, `warn`, `error`, or `info`.
- `Write-CabinetKeyValue` lists important runtime facts in aligned rows.
- `Write-CabinetHint` gives the next useful command or remediation step.

## Style rules

- Start with a banner for user-run commands.
- Group related work into short sections.
- Prefer clear verbs: `Building`, `Validating`, `Starting`, `Ready`.
- Include paths, ports, URLs, process IDs, and output locations as key/value rows.
- End successful flows with an `ok` status and one concrete verification hint.
- On failure, show the failed phase and the next action before surfacing raw tool
  output where practical.

## Color and CI

The helper uses PowerShell host colors when available and plain text when
`NO_COLOR` is set. It also emits a CI-compatible mode line when `CI` is present
so captured logs remain readable. Do not rely on color alone; every status must
also be visible in text.

## Example

```text
CABINET
Command: start-demo2
Purpose: Build and start the demo2 review runtime.

== Build ==
[RUN  ] Rebuilding Cabinet before launch.

== Runtime ==
  URL:         http://127.0.0.1:17882
  Mode:        restart
  Guard:       singleton endpoint guard enabled
[OK   ] Started in background.
  hint: Verify with: Invoke-WebRequest -UseBasicParsing http://127.0.0.1:17882/
```
