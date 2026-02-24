# Inventory Barcodes Screen Spec

## Use Cases
1. User adds barcodes manually to selected item.
2. User runs local lookup for barcode matches.
3. User launches external search when no local match exists.

## UI Sections
1. Barcode input/actions bar
2. Barcode list
3. Lookup result summary
4. External search output

## State Behavior
- Loading: barcodes loading state.
- Empty: "No barcodes" + add CTA.
- Error: add/lookup failures with retry path.
- Success: list + match summary.

## Acceptance Criteria
- [ ] Add barcode requires non-empty value.
- [ ] Lookup clearly shows local match count.
- [ ] External search URL is encoded safely.
- [ ] User sees explicit status: matched vs no match.

## Test Cases
- `INV-B-001` add and reload barcodes.
- `INV-B-002` local lookup match display.
- `INV-B-003` external search on no-match path.

