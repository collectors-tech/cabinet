# Cabinet – Technical Architecture

## Stack
- Backend: Go
- UI: Embedded Web UI (React or Svelte)
- DB: SQLite
- Search: SQLite FTS
- Media: Local file storage + thumbnail generation
- AI: OpenAI API (user-provided key)

## Package Structure
cmd/cabinet
internal/app
internal/api
internal/db
internal/scanner
internal/pricing
internal/licensing

## Scanner Engine
Pipeline:
1. Scheduled QuerySet
2. Provider Search (eBay)
3. Normalize Candidate
4. Deduplicate via fingerprint
5. Match to canonical item
6. Surface if not in collection

## Price Tracking
Daily snapshot job storing:
- Min
- Median
- Latest

## Licensing
- Signed JSON license file
- Public key embedded in binary
- Offline verification supported

