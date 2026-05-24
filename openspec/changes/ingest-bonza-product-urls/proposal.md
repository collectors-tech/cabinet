## Why

Users need pasted product URLs to become actionable Cabinet item drafts instead of inert text. Bonza Slot Cars is already in the AU webshop provider family, and its product pages expose enough WooCommerce Store API data to reliably populate Cabinet item fields from a pasted URL.

## What Changes

- Add deterministic pasted URL routing for known provider product URLs.
- Route `bonzaslotcars.com.au/product/<slug>/` to the Bonza provider and WooCommerce family ingestion path.
- Add Bonza product URL ingestion that uses WooCommerce Store API first and page metadata/HTML only as fallback.
- Normalize Bonza product data into a Cabinet item draft with source/evidence fields.
- Wire Inventory create paste processing to call the provider ingestion path and prefill the create-item modal.
- Preserve provenance so the original URL, provider, product id, and source payload summary remain attached to the created item.
- Add duplicate detection for repeated provider product URLs/product ids before item creation.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `provider-api-families`: add deterministic product URL routing and WooCommerce product detail ingestion requirements.
- `provider-au-webshops`: add Bonza product URL ingestion and normalized product detail extraction requirements.
- `ui-screen-inventory-items`: require the Inventory create paste flow to process supported provider URLs into item drafts with confirm-before-create behavior.

## Impact

- Backend provider routing and Bonza/WooCommerce ingestion APIs.
- Inventory create modal paste workflow and create-item draft mapping.
- Item source URL, evidence/history, image staging, pricing, stock, and category/type field mapping.
- API tests for mocked Bonza Store API responses.
- Cypress coverage for pasted Bonza URL processing in Inventory.
- Live/manual verification for `https://bonzaslotcars.com.au/product/bonza-mug-white/`.
