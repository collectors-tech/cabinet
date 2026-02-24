# CABINET - FULL FEATURE LIST

## 1. Application Core
- Desktop app (Windows + macOS)
- Single binary installer
- Signed application updates
- Update channel support (stable/beta)
- Embedded web UI
- Local SQLite database
- Automatic local backups
- Media folder for photos
- Import/export full collection (JSON + CSV)

## 2. Login and Authentication
Desktop-first authentication model:
- First launch creates a local user
- WebAuthn is required from v1
- First launch requires creating at least one WebAuthn credential
- App locks on startup
- Auto-lock after inactivity
- Add device credential
- Unlock via Touch ID / Windows Hello
- Multiple devices per user profile
- Recovery passphrase is optional (fallback only)
- Credential reset flow is required when a profile has no usable authenticator
- No cloud account required

This is local app authentication, not SaaS authentication.

## 3. User Profiles (Multi-user Support)
- Multiple local profiles supported
- Each profile has its own database
- Each profile has its own settings
- Each profile has its own API keys
- Each profile has its own license
- Profile switch from login screen

## 4. Collection Management
### Canonical Item
Fields:
- Brand
- Category
- Part number (primary identifier)
- Title
- Make
- Model
- Year
- Scale
- Series
- Description
- Tags
- Barcodes (multiple)
- Created/updated timestamps

### Instances
Fields:
- Condition
- Status (`sealed`, `blister`, `loose`, `custom`, `on_track`)
- Quantity
- Storage location
- Acquisition price
- Acquisition date
- Notes

Rules:
- One item can have many instances
- No auto-merge without user confirmation

## 5. Photo System
- Upload from desktop
- Upload from mobile browser
- Store original locally
- Generate thumbnail + preview
- Rebuild thumbnails
- Set primary image
- Delete images
- View full screen

## 6. Barcode System
- Detect barcode from uploaded image
- Manual barcode entry
- Multiple barcodes per item
- Lookup via local match
- Lookup via external search (eBay)
- Attach barcode to item
- Handle duplicate barcodes across variants

## 7. AI Assist (ChatGPT/OpenAI Only Initially)
User provides:
- OpenAI API key

Features:
- Identify item from photo
- Extract part number from listing title
- Normalize title into structured fields
- Suggest metadata improvements
- Confidence score
- Never auto-create without confirmation

Controls:
- Enable/disable AI

## 8. Scanner Engine
### Query Sets
User defines:
- Keywords
- Exclusions
- Max price
- Region
- Condition filter (if provider supports it)

### Execution
- Manual `Run Now`
- Scheduled runs
- Rate-limited execution

### Candidate Records
Stored fields:
- Listing ID
- Title
- Price
- Shipping
- URL
- Image
- Seller
- Stock status (`in_stock`, `low_stock`, `out_of_stock`, `unknown`)
- Stock count (if provider page exposes quantity)
- Stock count confidence/source note
- First seen
- Last seen
- Status

## 9. Matching Engine
For each candidate:
- Extract part number
- Match against canonical items
- Compute confidence

Output states:
- Matched
- Suggested
- Not in collection

## 10. Not In My Collection Panel
Shows:
- New items not owned

Filters:
- Price
- Query
- Date

Actions:
- Ignore
- Add to wishlist
- Track price
- Create item

## 11. Wishlist
- Linked to canonical item
- Target price
- Priority
- Notes
- Highlight scanner hits
- Below-target indicator

## 12. Price Tracking
- Mark item as tracked
- Daily price snapshot
- Store min, median, latest
- Store per-source stock count observation with timestamp
- Store in-stock/out-of-stock transitions by source
- Graph view
- Per-source breakdown
- Export price history

## 13. Dashboard
Displays:
- New discoveries
- Wishlist hits
- Price drops
- Low stock alerts for watchlisted/tracked items
- Restock alerts for previously out-of-stock items
- Recently added items

Collection stats:
- Total items
- Total instances
- Estimated value (optional)

## 14. Search and Filtering
- Full-text search
- Filter by brand, condition, status, tags, scale
- Saved filters
- Sort by date added, price, part number

## 15. Data Management
Export:
- JSON full backup
- CSV items

Import:
- JSON snapshot
- CSV mapping
- Dry-run preview before import apply
- Conflict resolution (merge/create-skip) with explicit user choice
- Schema version migration checks

Maintenance:
- Reindex search
- Repair database

## 16. Licensing System
- Free tier limit (example: 150 items)
- Pro unlocks unlimited items
- Pro unlocks scanner automation
- Pro unlocks price tracking
- Pro unlocks AI assist
- License file import
- Signature verification
- Offline validation
- Display license status

## 17. Settings
- Theme (light/dark)
- Scanner schedule
- Update channel preference (stable/beta)
- eBay credentials
- OpenAI API key
- Backup frequency
- Database location
- Rebuild thumbnails
- Reset ignore rules

## 18. Error Handling
- Scanner failure logs
- AI failure logs
- Provider health indicator
- Retry failed scans
- Clear error messages
- Crash recovery prompt after abnormal shutdown

## 19. Logging and Diagnostics
- Activity log
- Export logs
- Debug mode toggle
- Sensitive value redaction in logs (API keys, tokens, credentials)

## 20. Future Hooks (Designed, Not Active)
- Additional AI providers
- Additional scanner providers
- Share/export for comparison
- For-sale flag (disabled)
- Structured offers (disabled)

## 21. Non-Functional Requirements (v1)
- Startup time under 2.5 seconds on baseline hardware
- Search response under 200ms for 5k instances
- Scanner run (10 query sets) under 8 minutes
- Crash-free session rate target above 99 percent in beta
- Backup restore must be validated on Windows and macOS

## 22. Security and Privacy Requirements (v1)
- API keys and secrets stored in OS-backed secure storage
- No secrets stored in plaintext in SQLite
- WebAuthn credential operations logged without sensitive material
- License verification must run offline

## 23. In-App Chat Copilot (Collection Assistant)
- Persistent chat panel in app workspace (default right side)
- Toggle open/close from header
- Attach local files from disk into chat context (user selected only)
- Works across screens (global assistant, not tied to one tab)
- Context aware: active profile, active item, selected candidate, current filters
- User can ask:
  - Find items in collection
  - Explain gaps/missing variants
  - Suggest adding wishlist or tracked entries
  - Draft add-item payloads and metadata
- Assistant can propose actions, but never auto-commit destructive changes
- Confirm-before-apply flow for mutations (`create item`, `update item`, `track price`, `add wishlist`)
- Session/thread history stored locally per profile
- Chat settings:
  - Provider/model
  - System behavior mode (strict/suggestive)
  - Context sharing controls (what data can be sent)
- Keyboard support:
  - Open chat
  - Focus composer
  - Navigate history
- Docking behavior:
  - Right rail (default)
  - Optional pop-out panel
- Diagnostics:
  - Chat errors and tool-action logs in activity diagnostics
