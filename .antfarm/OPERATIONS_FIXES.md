# Antfarm Operations Fixes (Cabinet)

## Incident: Validator/build lanes pending without execution

### Symptoms
- `cabinet` / `cabinet-validator` runs show `running` with `pending` top step.
- `step peek cabinet-validator_validator` returns `HAS_WORK` but no claim completion.
- `cabinet.exe` may not launch if validator step never claimed.

### Root causes observed
1. Cron recreation drifted worker payloads to generic quick-check message.
2. Delivery mode drift and worker no-op behavior caused queue stalls.
3. Run/step state could appear active while no real process activity occurred.

### Applied fix pattern (mandatory)
1. Recreate workflow crons:
   - `workflow ensure-crons cabinet`
   - `workflow ensure-crons cabinet-validator`
2. Immediately patch all `antfarm/cabinet/*` and `antfarm/cabinet-validator/*` jobs:
   - `delivery.mode = none`
   - `payload.model = openai-codex/gpt-5.3-codex`
   - `payload.timeoutSeconds = 900`
   - explicit `peek -> claim -> execute -> step complete|step fail` instructions
3. Run medic:
   - `node C:\projects\antfarm\dist\cli\cli.js medic run`
4. Validate runtime parameters for validator launch:
   - source exe: `C:\projects\collectors-tech\cabinet\bin\cabinet.exe`
   - target exe: `C:\projects\collectors-tech\cabinet\tmp\demo1\bin\cabinet.exe`
   - args: `-allow-parallel -no-open-browser -data-dir ... -profile demo1-helper -instance-name demo1-helper -port 17881`
   - verify source/target hash parity.

### Emergency unstick (manual)
- Stop broken run, start fresh validator run, then:
  - `step claim cabinet-validator_validator`
- If launch not observed, relaunch strict demo1 runtime manually and capture evidence.

## Non-negotiable
A run is healthy only when step transitions (claim/complete/fail) are observed and runtime process evidence exists.
