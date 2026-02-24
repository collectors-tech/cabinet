# 08 Gap and Intuitiveness Review (Per Section)

## Review Method
- Compare intended user goal per screen vs current interaction cost.
- Identify ambiguity, hidden actions, high-friction flows, and terminology issues.
- Classify gap severity: `Critical`, `High`, `Medium`, `Low`.

## Home
### Gaps
- `High`: Dashboard currently shows metrics but limited action-first triage.
- `Medium`: No consistent snooze/dismiss workflow for recurring alerts.

### Intuitiveness Risk
- Users see numbers but not clear next actions.

### Required Fixes
- Implement strict attention cards and action buttons from `03`.
- Add clear severity badges and card ordering.

## Inventory: Items
### Gaps
- `High`: Risk of form overload if quick and advanced metadata are not clearly separated.
- `Medium`: Selection persistence may be unclear after filter changes.

### Intuitiveness Risk
- New users may hesitate if first interaction shows too many fields.

### Required Fixes
- Default to quick-add minimal fields.
- Move advanced fields behind explicit expand section.

## Inventory: Photos
### Gaps
- `Medium`: Camera failure paths can feel abrupt without guidance.
- `Low`: Fullscreen behavior consistency across keyboard/mouse needs strict validation.

### Intuitiveness Risk
- Users may interpret permission errors as app failure.

### Required Fixes
- Add explicit permission guidance and fallback to file upload.
- Ensure close and escape behavior is obvious.

## Inventory: Barcodes
### Gaps
- `Medium`: Local vs external lookup outcomes can be ambiguous.

### Intuitiveness Risk
- Users may not know if barcode was saved, matched, or only searched externally.

### Required Fixes
- Show explicit status labels: `Saved`, `Matched local`, `No local match`.

## Inventory: AI Assist
### Gaps
- `High`: Confidence and confirmation semantics must be explicit to maintain trust.

### Intuitiveness Risk
- Users may fear silent data mutation by AI.

### Required Fixes
- Always present a review diff and explicit apply step.
- Surface confidence and fallback guidance.

## Discover
### Gaps
- `High`: Candidate action outcomes can be unclear without inline status updates.
- `Medium`: Filter impact discoverability can be weak at large volumes.

### Intuitiveness Risk
- Triage feels noisy if actions do not produce immediate visible state change.

### Required Fixes
- Inline row state updates after each action.
- Persistent filter chips with clear reset.

## Scanner
### Gaps
- `High`: Technical terms (query set, cron, provider health) may overwhelm non-technical users.

### Intuitiveness Risk
- Users avoid automation due to unclear mental model.

### Required Fixes
- Add plain-language helper text and templates.
- Surface "safe defaults" for scheduling.

## Reports
### Gaps
- `Medium`: Export and trend purpose separation may be unclear.

### Intuitiveness Risk
- Users may not know whether to use Reports vs Home.

### Required Fixes
- Add section descriptions and report presets.
- Keep Home for triage, Reports for analysis/export only.

## Settings
### Gaps
- `Critical`: Destructive actions can be risky if not consistently guarded.
- `Medium`: Diagnostics output may be too technical for first-line troubleshooting.

### Intuitiveness Risk
- Fear of breaking data or configuration.

### Required Fixes
- Confirm steps on destructive actions.
- Human-readable diagnostics summary before raw details.

## Cross-Cutting Gaps
1. Terminology consistency
- Current labels must align with IA terms.

2. Action feedback consistency
- Every action must show deterministic success/failure feedback.

3. Empty-state guidance
- Every empty screen needs a "what to do next" CTA.

4. Mobile parity
- Drawer and tab behavior must match desktop intent.

## Prioritized Remediation Backlog
1. Critical: Settings destructive guardrails.
2. High: Home action-first triage, Discover inline action feedback, Scanner simplification.
3. Medium: Inventory form simplification and report guidance.
4. Low: polish and terminology cleanup.

## Review Checklist (for per-section scrutiny sessions)
- [ ] Can a first-time user complete primary task in under 60 seconds?
- [ ] Is there one obvious next action at all times?
- [ ] Are errors actionable and recoverable without support?
- [ ] Are labels plain-language and consistent?
- [ ] Are mobile and desktop flows functionally equivalent?

