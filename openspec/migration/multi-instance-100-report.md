# Multi-instance 100 Scale Report

Generated: 2026-03-04T18:57:34.4663538+11:00

## Summary
- Requested instances: 100
- Healthy instances: 100
- Failed instances: 0
- Duration: 00:02:23.1505988
- Run directory: D:\projects\collectors-tech\cabinet\.\.tmp\runtime-multi-instance\20260304-185511
- Instance manifest: D:\projects\collectors-tech\cabinet\.\.tmp\runtime-multi-instance\20260304-185511\instances.json

## Guardrails
- Memory guardrail MB: 12000
- CPU seconds guardrail: 1800
- Backoff seconds: 3

## Failures
- none

## Acceptance Snapshot
- Unique URL/port per instance: pass
- Health-check visibility for all instances: pass
- Collision-safe lock/data isolation: pass (per-instance CABINET_DATA_DIR)
- Orchestration start/status/stop: pass
