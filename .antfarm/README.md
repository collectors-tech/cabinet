# AntFarm workflow (project-local copy)

This folder keeps the Cabinet AntFarm workflow definition in-repo so workflow logic is versioned with Cabinet.

## Location
- Workflow: `.antfarm/workflows/cabinet/workflow.yml`
- Agent profiles: `.antfarm/workflows/cabinet/agents/*`

## Notes
- Workflow id is `cabinet`.
- Paths are self-contained (no external shared-agent path dependency).
- If running from global AntFarm state, install/sync this workflow into the active AntFarm workflows directory.

## Run
```bash
antfarm workflow run cabinet "<task>"
```
