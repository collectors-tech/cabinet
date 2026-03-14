# Antfarm Worker Runbook (Cabinet)

Use this for worker IDs under `cabinet_*` and `cabinet-validator_*`.

## Required lifecycle (non-negotiable)
1. `step peek <workerId>`
2. If `NO_WORK`: stop.
3. If `HAS_WORK`: `step claim <workerId>`
4. Execute claimed `input` in repo context:
   - build lane: `C:\projects\collectors-tech\cabinet`
   - validator lane: `C:\projects\collectors-tech\cabinet`
5. Always end with exactly one:
   - `step complete <stepId>` with structured output, or
   - `step fail <stepId> "<error>"`

Never exit without complete/fail after a successful claim.

## Canonical commands
```powershell
node C:\projects\antfarm\dist\cli\cli.js step peek <workerId>
node C:\projects\antfarm\dist\cli\cli.js step claim <workerId>
node C:\projects\antfarm\dist\cli\cli.js step complete <stepId>
node C:\projects\antfarm\dist\cli\cli.js step fail <stepId> "<error>"
```

## Validator runtime standard
When validating Cabinet demo1 runtime, enforce:
- `-allow-parallel -no-open-browser`
- `-data-dir C:\projects\collectors-tech\cabinet\tmp\demo1\data`
- `-profile demo1-helper -instance-name demo1-helper -port 17881`
- source/target hash parity before launch.
